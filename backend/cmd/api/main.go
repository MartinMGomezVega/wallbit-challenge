// Command api levanta el backend: config → infra → use cases → http.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/config"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	arcamock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/arca/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	fxmock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
	notifylog "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/notify/log"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/scheduler"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
	apihttp "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/interfaces/http"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fatal", "error", err.Error())
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	var clk ports.Clock = clock.Real{}
	if cfg.AppNow != nil {
		clk = clock.NewFixed(*cfg.AppNow)
	}

	// Infraestructura
	var arcaOpts []arcamock.Option
	if cfg.MockLatency > 0 {
		arcaOpts = append(arcaOpts, arcamock.WithLatency(cfg.MockLatency))
	}
	arcaClient, err := arcamock.New(arcaOpts...)
	if err != nil {
		return err
	}

	var fxOpts []fxmock.Option
	if cfg.MockLatency > 0 {
		fxOpts = append(fxOpts, fxmock.WithLatency(cfg.MockLatency))
	}
	fxClient, err := fxmock.New(fxOpts...)
	if err != nil {
		return err
	}

	paramsRepo := memory.NewParametersRepository()
	clientRepo := memory.NewClientRepository()
	movementRepo := memory.NewMovementRepository()
	classificationRepo := memory.NewClassificationRepository()
	invoiceRepo := memory.NewInvoiceRepository()
	purchaseInvoiceRepo := memory.NewPurchaseInvoiceRepository()
	clientInputRepo := memory.NewClientInputRepository()
	closingRepo := memory.NewClosingRepository()
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()
	notifSender := notifylog.NewSender(log)

	// Generación de datos semilla determinísticos
	idGen := idgen.NewSequential()
	seedStart := time.Now()
	ds, err := seed.Generate(context.Background(), seed.SeedConfig{
		Seed:    cfg.Seed,
		Clients: cfg.SeedClients,
		AppNow:  clk.Now(),
	}, fxClient, idGen)
	if err != nil {
		return fmt.Errorf("error generando seed: %w", err)
	}

	if err := clientRepo.SaveAll(context.Background(), ds.Clients); err != nil {
		return fmt.Errorf("error guardando clientes seed: %w", err)
	}
	if err := movementRepo.SaveAll(context.Background(), ds.Movements); err != nil {
		return fmt.Errorf("error guardando movimientos seed: %w", err)
	}
	if err := invoiceRepo.SaveAll(context.Background(), ds.Invoices); err != nil {
		return fmt.Errorf("error guardando facturas seed: %w", err)
	}
	if err := purchaseInvoiceRepo.SaveAll(context.Background(), ds.PurchaseInvoices); err != nil {
		return fmt.Errorf("error guardando facturas de compra seed: %w", err)
	}
	if err := clientInputRepo.SaveAll(context.Background(), ds.ClientInputs); err != nil {
		return fmt.Errorf("error guardando inputs de clientes seed: %w", err)
	}

	// Clasificación inicial concurrente con worker pool (T002.10)
	classifications := classifyMovementsWorkerPool(ds.Movements, cfg.ClosingWorkers)
	if err := classificationRepo.SaveAll(context.Background(), classifications); err != nil {
		return fmt.Errorf("error guardando clasificaciones iniciales: %w", err)
	}

	log.Info("seed inicializado",
		"clientes", len(ds.Clients),
		"movimientos", len(ds.Movements),
		"facturas", len(ds.Invoices),
		"facturasCompra", len(ds.PurchaseInvoices),
		"clientInputs", len(ds.ClientInputs),
		"clasificaciones", len(classifications),
		"duracionMs", time.Since(seedStart).Milliseconds(),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	computeClosing := &usecase.ComputeClosing{
		Clients:          clientRepo,
		Movements:        movementRepo,
		Classifications:  classificationRepo,
		Invoices:         invoiceRepo,
		PurchaseInvoices: purchaseInvoiceRepo,
		ClientInputs:     clientInputRepo,
		Parameters:       paramsRepo,
		Closings:         closingRepo,
		TaxScales:        arcaClient,
		DueDates:         arcaClient,
		ExchangeRates:    fxClient,
		Clock:            clk,
		IDGen:            idGen,
	}

	notifier := &usecase.RecalcOnChange{
		Compute:  computeClosing,
		Closings: closingRepo,
	}

	dispatchOutbox := &usecase.DispatchOutbox{
		Outbox: outboxRepo,
		Sender: notifSender,
		Clock:  clk,
	}

	actionNotices := &usecase.CreateActionRequiredNotices{
		Closings:       closingRepo,
		Notifications:  notifRepo,
		Outbox:         outboxRepo,
		Clock:          clk,
		IDGen:          idGen,
		DispatchOutbox: dispatchOutbox,
	}

	sendRemindersUC := &usecase.SendPaymentReminders{
		Closings:       closingRepo,
		Notifications:  notifRepo,
		Outbox:         outboxRepo,
		Parameters:     paramsRepo,
		Clock:          clk,
		IDGen:          idGen,
		DispatchOutbox: dispatchOutbox,
	}

	// Ejecución inicial del cierre del mes anterior (criterio 7)
	prevMonth := period.FromTime(clk.Now()).Prev()
	runClosingUC := &usecase.RunMonthlyClosing{
		Clients:       clientRepo,
		Compute:       computeClosing,
		Workers:       cfg.ClosingWorkers,
		ActionNotices: actionNotices,
	}

	closingStart := time.Now()
	summary, err := runClosingUC.Execute(context.Background(), prevMonth)
	if err != nil {
		return fmt.Errorf("error ejecutando cierre del mes anterior: %w", err)
	}

	log.Info("cierre inicial completado",
		"periodo", prevMonth.String(),
		"procesados", summary.Processed,
		"creados", summary.Created,
		"actualizados", summary.Updated,
		"sinCambios", summary.Unchanged,
		"fallidos", len(summary.Failed),
		"duracionMs", time.Since(closingStart).Milliseconds(),
	)

	// Iniciar scheduler automático (criterio 8)
	sched := scheduler.NewScheduler(runClosingUC, clk)
	sched.ReminderRunner = sendRemindersUC
	go sched.Start(ctx)

	handler := apihttp.NewRouter(apihttp.Deps{
		Clock:                     clk,
		Logger:                    log,
		CORSOrigin:                cfg.CORSOrigin,
		ParametersRepo:            paramsRepo,
		ClientRepo:                clientRepo,
		MovementRepo:              movementRepo,
		ClassificationRepo:        classificationRepo,
		InvoiceRepo:               invoiceRepo,
		PurchaseInvoiceRepo:       purchaseInvoiceRepo,
		ClientInputRepo:           clientInputRepo,
		LatestExclusionDateReader: closingRepo,
		ClosingRepo:               closingRepo,
		TaxScales:                 arcaClient,
		DueDates:                  arcaClient,
		ExchangeRates:             fxClient,
		Workers:                   cfg.ClosingWorkers,
		IDGen:                     idGen,
		ChangeNotifier:            notifier,
		NotificationRepo:          notifRepo,
		OutboxRepo:                outboxRepo,
		NotificationSender:        notifSender,
		RealNow:                   clock.Real{}.Now,
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "port", cfg.Port, "now", clk.Now().Format(time.RFC3339))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func classifyMovementsWorkerPool(movements []movement.Movement, workers int) []movement.Classification {
	if workers <= 0 {
		workers = 4
	}
	results := make([]movement.Classification, len(movements))
	ch := make(chan int, len(movements))
	for i := range movements {
		ch <- i
	}
	close(ch)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for idx := range ch {
				results[idx] = movement.Classify(movements[idx], nil)
			}
		}()
	}
	wg.Wait()
	return results
}
