package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// rollbackFakeTxManager simula rollback de estado en fakes cuando fn devuelve error.
type rollbackFakeTxManager struct {
	rollback   func()
	rolledBack bool
}

func (t *rollbackFakeTxManager) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		t.rolledBack = true
		if t.rollback != nil {
			t.rollback()
		}
		return err
	}
	return nil
}

type failingOutboxRepo struct {
	*fakes.OutboxRepository
	failOnEnqueue error
}

func (r *failingOutboxRepo) Enqueue(ctx context.Context, m notification.OutboundMessage) error {
	if r.failOnEnqueue != nil {
		return r.failOnEnqueue
	}
	return r.OutboxRepository.Enqueue(ctx, m)
}

type failOnNthInvoiceRepo struct {
	*fakes.InvoiceRepository
	failAt int
	calls  int
}

func (r *failOnNthInvoiceRepo) Save(ctx context.Context, inv invoice.Invoice) error {
	r.calls++
	if r.calls == r.failAt {
		return errors.New("simulated invoice save failure")
	}
	return r.InvoiceRepository.Save(ctx, inv)
}

func TestApproveClosing_RollsBackOnFailureInsideTx(t *testing.T) {
	ctx := context.Background()
	_, _, _, _, _, closingsRepo, clientsRepo, notifRepo, outboxRepo, sender, clk := setupApprovalSuite(t)

	c := closing.NewClosing(
		"clo_tx", "cli_01", period.MustParse("2026-10"), true, "hash_tx", signal.SeverityGreen,
		nil, "OK",
		&tax.MonotributoAssessment{
			CurrentCategory:   tax.CatA,
			Fee:               money.New(100000, money.ARS),
			FeeDueDate:        clk.Now().AddDate(0, 0, 10),
			BilledInPeriodUSD: money.New(10000, money.USD),
			BilledInPeriodARS: money.New(10000000, money.ARS),
		},
		nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, c)

	txMgr := &rollbackFakeTxManager{
		rollback: func() {
			notifRepo.Notifications = make(map[string]notification.Notification)
			outboxRepo.Messages = make(map[string]notification.OutboundMessage)
			_ = closingsRepo.Save(ctx, c)
		},
	}

	apprUC := &usecase.ApproveClosing{
		Closings:      closingsRepo,
		Clients:       clientsRepo,
		Notifications: notifRepo,
		Outbox:        &failingOutboxRepo{OutboxRepository: outboxRepo, failOnEnqueue: errors.New("outbox down")},
		Sender:        sender,
		Clock:         clk,
		IDGen:         fakes.NewIDGenerator(),
		Tx:            txMgr,
	}

	_, err := apprUC.Execute(ctx, usecase.ApproveClosingRequest{
		ClosingID: "clo_tx",
		Version:   1,
	})
	if err == nil {
		t.Fatal("expected error when outbox fails")
	}
	if !txMgr.rolledBack {
		t.Fatal("expected transaction rollback")
	}

	notifs, _ := notifRepo.ByClient(ctx, "cli_01")
	if len(notifs) != 0 {
		t.Fatalf("expected 0 notifications after rollback, got %d", len(notifs))
	}
	got, _ := closingsRepo.Get(ctx, "clo_tx")
	if got.Status != closing.StatusReadyForReview {
		t.Fatalf("expected closing status READY_FOR_REVIEW after rollback, got %s", got.Status)
	}
}

func TestImportInvoicesCSV_RollsBackOnFailureInsideTx(t *testing.T) {
	ctx := context.Background()
	clk := &fakes.Clock{Current: time.Date(2026, 11, 1, 10, 0, 0, 0, period.Location())}
	clientsRepo := fakes.NewClientRepository()
	_ = clientsRepo.Save(ctx, client.Client{
		ID:     "cli_csv",
		Name:   "CSV Client",
		Regime: client.RegimeMonotributo,
	})
	invRepo := fakes.NewInvoiceRepository()

	txMgr := &rollbackFakeTxManager{
		rollback: func() {
			invRepo.Invoices = make(map[string]invoice.Invoice)
		},
	}

	uc := &usecase.ImportInvoicesCSV{
		Clients:  clientsRepo,
		Invoices: &failOnNthInvoiceRepo{InvoiceRepository: invRepo, failAt: 2},
		IDGen:    fakes.NewIDGenerator(),
		Clock:    clk,
		Tx:       txMgr,
	}

	csvData := strings.Join([]string{
		"Fecha;Tipo;Punto de Venta;Número Desde;Nro. Doc. Receptor;Denominación Receptor;Tipo Cambio;Moneda;Imp. Neto Gravado;IVA;Imp. Total",
		"01/10/2026;11 - Factura C;0001;00000001;20111111111;Cliente Uno;1;PES;1000;0;1000",
		"02/10/2026;11 - Factura C;0001;00000002;20111111111;Cliente Uno;1;PES;2000;0;2000",
	}, "\n")

	_, err := uc.Execute(ctx, "cli_csv", strings.NewReader(csvData))
	if err == nil {
		t.Fatal("expected error on second invoice save")
	}
	if !txMgr.rolledBack {
		t.Fatal("expected transaction rollback")
	}
	list, _ := invRepo.ByClient(ctx, "cli_csv")
	if len(list) != 0 {
		t.Fatalf("expected 0 persisted invoices after rollback, got %d", len(list))
	}
}
