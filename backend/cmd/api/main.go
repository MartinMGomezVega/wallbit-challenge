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
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/postgres"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/scheduler"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
	apihttp "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
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

	r, err := buildRepos(context.Background(), cfg, log)
	if err != nil {
		return err
	}
	defer r.close()

	paramsRepo := r.params
	clientRepo := r.clients
	movementRepo := r.movements
	classificationRepo := r.classifications
	invoiceRepo := r.invoices
	purchaseInvoiceRepo := r.purchaseInvoices
	clientInputRepo := r.clientInputs
	closingRepo := r.closings
	closingQuery := r.closingQuery
	statsQuery := r.statsQuery
	txManager := r.txManager
	notifRepo := r.notifications
	outboxRepo := r.outbox
	notifSender := notifylog.NewSender(log)

	// Generación y carga de datos semilla determinísticos
	idGen := idgen.NewSequential()
	if err := loadSeed(context.Background(), cfg, log, r, fxClient, idGen, clk); err != nil {
		return err
	}
	if cfg.Storage == "postgres" && r.pool != nil {
		_ = postgres.InitIDGen(context.Background(), r.pool, idGen)
	}

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
		Tx:               txManager,
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

	// Ejecución inicial de cierres: 2026-09 (cerrado y notificado) y 2026-10 (en curso)
	pSep := period.MustParse("2026-09")
	pOct := period.MustParse("2026-10")
	runClosingUC := &usecase.RunMonthlyClosing{
		Clients:       clientRepo,
		Compute:       computeClosing,
		Workers:       cfg.ClosingWorkers,
		ActionNotices: actionNotices,
	}

	if _, err := runClosingUC.Execute(context.Background(), pSep); err != nil {
		return fmt.Errorf("error ejecutando cierre de septiembre: %w", err)
	}
	if err := usecase.SeedSeptemberClosings(context.Background(), closingRepo, clientRepo, idGen, uint64(cfg.Seed)); err != nil {
		return fmt.Errorf("error cerrando período de septiembre: %w", err)
	}

	closingStart := time.Now()
	summary, err := runClosingUC.Execute(context.Background(), pOct)
	if err != nil {
		return fmt.Errorf("error ejecutando cierre de octubre: %w", err)
	}
	if cfg.Storage == "postgres" && r.pool != nil {
		_ = postgres.InitIDGen(context.Background(), r.pool, idGen)
	}

	log.Info("cierre inicial completado",
		"periodo", pOct.String(),
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
		ClosingQuery:              closingQuery,
		StatsQuery:                statsQuery,
		TxManager:                 txManager,
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

type repos struct {
	pool             *pgxpool.Pool
	txManager        ports.TxManager
	params           ports.ParametersRepository
	clients          ports.ClientRepository
	movements        ports.MovementRepository
	classifications  ports.ClassificationRepository
	invoices         ports.InvoiceRepository
	purchaseInvoices ports.PurchaseInvoiceRepository
	clientInputs     ports.ClientInputRepository
	closings         ports.ClosingRepository
	exclusions       ports.LatestExclusionDateReader
	closingQuery     ports.ClosingQuery
	statsQuery       ports.StatsQuery
	notifications    ports.NotificationRepository
	outbox           ports.OutboxRepository
	close            func()
}

func buildRepos(ctx context.Context, cfg config.Config, log *slog.Logger) (repos, error) {
	if cfg.Storage == "postgres" {
		pool, err := postgres.OpenPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return repos{}, err
		}
		if err := postgres.Migrate(ctx, pool); err != nil {
			pool.Close()
			return repos{}, err
		}
		log.Info("storage inicializado", "driver", "postgres")
		cr := postgres.NewClosingRepository(pool)
		return repos{
			pool:             pool,
			txManager:        postgres.NewTxManager(pool),
			params:           postgres.NewParametersRepository(pool),
			clients:          postgres.NewClientRepository(pool),
			movements:        postgres.NewMovementRepository(pool),
			classifications:  postgres.NewClassificationRepository(pool),
			invoices:         postgres.NewInvoiceRepository(pool),
			purchaseInvoices: postgres.NewPurchaseInvoiceRepository(pool),
			clientInputs:     postgres.NewClientInputRepository(pool),
			closings:         cr,
			exclusions:       cr,
			closingQuery:     postgres.NewClosingQuery(pool),
			statsQuery:       postgres.NewStatsQuery(pool),
			notifications:    postgres.NewNotificationRepository(pool),
			outbox:           postgres.NewOutboxRepository(pool),
			close:            pool.Close,
		}, nil
	}

	log.Info("storage inicializado", "driver", "memory")
	clientRepo := memory.NewClientRepository()
	closingRepo := memory.NewClosingRepository()
	return repos{
		txManager:        memory.NewTxManager(),
		params:           memory.NewParametersRepository(),
		clients:          clientRepo,
		movements:        memory.NewMovementRepository(),
		classifications:  memory.NewClassificationRepository(),
		invoices:         memory.NewInvoiceRepository(),
		purchaseInvoices: memory.NewPurchaseInvoiceRepository(),
		clientInputs:     memory.NewClientInputRepository(),
		closings:         closingRepo,
		exclusions:       closingRepo,
		closingQuery:     memory.NewClosingQuery(closingRepo, clientRepo),
		statsQuery:       memory.NewStatsQuery(closingRepo),
		notifications:    memory.NewNotificationRepository(),
		outbox:           memory.NewOutboxRepository(),
		close:            func() {},
	}, nil
}

func loadSeed(
	ctx context.Context,
	cfg config.Config,
	log *slog.Logger,
	r repos,
	fxClient ports.ExchangeRateProvider,
	idGen ports.IDGenerator,
	clk ports.Clock,
) error {
	seedStart := time.Now()
	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    cfg.Seed,
		Clients: cfg.SeedClients,
		AppNow:  clk.Now(),
	}, fxClient, idGen)
	if err != nil {
		return fmt.Errorf("error generando seed: %w", err)
	}
	classifications := classifyMovementsWorkerPool(ds.Movements, cfg.ClosingWorkers)

	if cfg.Storage == "postgres" {
		inserted, err := postgres.LoadSeed(ctx, r.pool, &ds, classifications, cfg.SeedReset)
		if err != nil {
			return fmt.Errorf("error cargando seed en postgres: %w", err)
		}
		log.Info("seed postgres",
			"insertado", inserted,
			"reset", cfg.SeedReset,
			"clientes", len(ds.Clients),
			"duracionMs", time.Since(seedStart).Milliseconds(),
		)
		return nil
	}

	if err := r.clients.SaveAll(ctx, ds.Clients); err != nil {
		return fmt.Errorf("error guardando clientes seed: %w", err)
	}
	if err := r.movements.SaveAll(ctx, ds.Movements); err != nil {
		return fmt.Errorf("error guardando movimientos seed: %w", err)
	}
	if err := r.invoices.SaveAll(ctx, ds.Invoices); err != nil {
		return fmt.Errorf("error guardando facturas seed: %w", err)
	}
	if err := r.purchaseInvoices.SaveAll(ctx, ds.PurchaseInvoices); err != nil {
		return fmt.Errorf("error guardando facturas de compra seed: %w", err)
	}
	if err := r.clientInputs.SaveAll(ctx, ds.ClientInputs); err != nil {
		return fmt.Errorf("error guardando inputs de clientes seed: %w", err)
	}
	if err := r.classifications.SaveAll(ctx, classifications); err != nil {
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
	return nil
}
