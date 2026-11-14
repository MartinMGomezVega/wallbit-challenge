package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

type failNTimesSender struct {
	mu        sync.Mutex
	failCount int
	calls     int
}

func (s *failNTimesSender) Send(_ context.Context, _ notification.OutboundMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.calls <= s.failCount {
		return errors.New("simulated network failure")
	}
	return nil
}

func setupApprovalSuite(t *testing.T) (
	*usecase.ApproveClosing,
	*usecase.ApproveBatch,
	*usecase.DispatchOutbox,
	*usecase.CreateActionRequiredNotices,
	*usecase.SendPaymentReminders,
	*fakes.ClosingRepository,
	*fakes.ClientRepository,
	*fakes.NotificationRepository,
	*fakes.OutboxRepository,
	*failNTimesSender,
	*fakes.Clock,
) {
	ctx := context.Background()
	clk := &fakes.Clock{Current: time.Date(2026, 11, 1, 10, 0, 0, 0, period.Location())}
	idGen := fakes.NewIDGenerator()

	closingsRepo := fakes.NewClosingRepository()
	clientsRepo := fakes.NewClientRepository()
	notifRepo := fakes.NewNotificationRepository()
	outboxRepo := fakes.NewOutboxRepository()
	paramsRepo := &fakes.ParametersRepository{}
	sender := &failNTimesSender{failCount: 0}

	// Clients
	cli1 := client.Client{ID: "cli_01", Name: "Ignacio Sosa", Regime: client.RegimeMonotributo}
	cli2 := client.Client{ID: "cli_02", Name: "Lucía Méndez", Regime: client.RegimeRI}
	_ = clientsRepo.SaveAll(ctx, []client.Client{cli1, cli2})

	dispatchOutbox := &usecase.DispatchOutbox{
		Outbox: outboxRepo,
		Sender: sender,
		Clock:  clk,
	}

	approveClosing := &usecase.ApproveClosing{
		Closings:       closingsRepo,
		Clients:        clientsRepo,
		Notifications:  notifRepo,
		Outbox:         outboxRepo,
		Sender:         sender,
		Clock:          clk,
		IDGen:          idGen,
		DispatchOutbox: dispatchOutbox,
	}

	approveBatch := &usecase.ApproveBatch{
		ApproveClosing: approveClosing,
		Closings:       closingsRepo,
		NumWorkers:     4,
	}

	createActionNotices := &usecase.CreateActionRequiredNotices{
		Closings:       closingsRepo,
		Notifications:  notifRepo,
		Outbox:         outboxRepo,
		IDGen:          idGen,
		Clock:          clk,
		DispatchOutbox: dispatchOutbox,
	}

	sendReminders := &usecase.SendPaymentReminders{
		Closings:       closingsRepo,
		Notifications:  notifRepo,
		Outbox:         outboxRepo,
		Parameters:     paramsRepo,
		Clock:          clk,
		IDGen:          idGen,
		DispatchOutbox: dispatchOutbox,
	}

	return approveClosing, approveBatch, dispatchOutbox, createActionNotices, sendReminders, closingsRepo, clientsRepo, notifRepo, outboxRepo, sender, clk
}

func TestApproveClosing_IndividualSuccess(t *testing.T) {
	ctx := context.Background()
	apprUC, _, _, _, _, closingsRepo, _, notifRepo, outboxRepo, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)
	feeDue := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_01", "cli_01", p, true, "hash123", signal.SeverityGreen,
		nil, "Todo ok",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(150000, money.USD),
			BilledInPeriodARS: money.New(225000000, money.ARS),
			CurrentCategory:   tax.CatK,
			Fee:               money.New(16144460, money.ARS),
			FeeDueDate:        feeDue,
		},
		nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, c)

	res, err := apprUC.Execute(ctx, usecase.ApproveClosingRequest{
		ClosingID:  "clo_01",
		Version:    1,
		ClientNote: "Aprobado sin observaciones",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Warning != "" {
		t.Errorf("expected empty warning, got %q", res.Warning)
	}
	if res.Closing.Status != closing.StatusNotified {
		t.Errorf("status = %s, want NOTIFIED", res.Closing.Status)
	}
	if len(res.Closing.Approvals) != 1 {
		t.Fatalf("expected 1 approval, got %d", len(res.Closing.Approvals))
	}
	if res.Closing.Published == nil {
		t.Fatalf("expected Published to be set")
	}

	// Verify in-app notification
	notifs, _ := notifRepo.ByClient(ctx, "cli_01")
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifs))
	}
	if notifs[0].Kind != notification.KindClosingReady {
		t.Errorf("kind = %s, want CLOSING_READY", notifs[0].Kind)
	}
	if !strings.Contains(notifs[0].Body, "Facturaste US$ 1.500,00") {
		t.Errorf("in-app notification body missing figures: %s", notifs[0].Body)
	}

	// Verify outbound outbox messages
	pending, _ := outboxRepo.Pending(ctx, 10)
	if len(pending) != 0 {
		t.Errorf("expected outbox to be dispatched, but %d remain pending", len(pending))
	}
}

func TestApproveClosing_PendingClientInput_Warning(t *testing.T) {
	ctx := context.Background()
	apprUC, _, _, _, _, closingsRepo, _, _, _, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)
	feeDue := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_pending", "cli_01", p, false, "hash123", signal.SeverityYellow,
		nil, "Esperando",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.Zero(money.USD),
			BilledInPeriodARS: money.Zero(money.ARS),
			CurrentCategory:   tax.CatA,
			Fee:               money.New(500000, money.ARS),
			FeeDueDate:        feeDue,
		},
		nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, c)

	res, err := apprUC.Execute(ctx, usecase.ApproveClosingRequest{
		ClosingID: "clo_pending",
		Version:   1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Warning != "CLIENT_INPUT_NOT_CONFIRMED" {
		t.Errorf("warning = %q, want CLIENT_INPUT_NOT_CONFIRMED", res.Warning)
	}
	if res.Closing.Status != closing.StatusNotified {
		t.Errorf("status = %s, want NOTIFIED", res.Closing.Status)
	}
}

func TestApproveClosing_Errors(t *testing.T) {
	ctx := context.Background()
	apprUC, _, _, _, _, closingsRepo, _, _, _, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)
	feeDue := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_err", "cli_01", p, true, "hash123", signal.SeverityGreen,
		nil, "Todo ok",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.Zero(money.USD),
			BilledInPeriodARS: money.Zero(money.ARS),
			CurrentCategory:   tax.CatA,
			Fee:               money.New(500000, money.ARS),
			FeeDueDate:        feeDue,
		},
		nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, c)

	// 1. Not found
	_, err := apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "non_existent", Version: 1})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// 2. Version conflict
	_, err = apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_err", Version: 99})
	if !errors.Is(err, domain.ErrVersionConflict) {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}

	// 3. Note too long (> 500 chars)
	tooLongNote := strings.Repeat("a", 501)
	_, err = apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_err", Version: 1, ClientNote: tooLongNote})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for long note, got %v", err)
	}

	// 4. Approve successfully
	_, err = apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_err", Version: 1})
	if err != nil {
		t.Fatalf("first approval failed: %v", err)
	}

	// 5. Already approved
	_, err = apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_err", Version: 1})
	if !errors.Is(err, domain.ErrAlreadyApproved) {
		t.Errorf("expected ErrAlreadyApproved, got %v", err)
	}
}

func TestApproveClosing_Reapproval_GeneratesClosingUpdated(t *testing.T) {
	ctx := context.Background()
	apprUC, _, _, _, _, closingsRepo, _, notifRepo, _, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)
	feeDue := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_reopen", "cli_01", p, true, "hash1", signal.SeverityGreen,
		nil, "Todo ok",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.Zero(money.USD),
			BilledInPeriodARS: money.Zero(money.ARS),
			CurrentCategory:   tax.CatA,
			Fee:               money.New(500000, money.ARS),
			FeeDueDate:        feeDue,
		},
		nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, c)

	// First approval
	_, _ = apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_reopen", Version: 1})

	// Advance clock
	clk.Current = clk.Current.Add(time.Hour)

	// Reopen by applying changes
	saved, _ := closingsRepo.Get(ctx, "clo_reopen")
	saved.ApplyResult("hash2", signal.SeverityYellow, nil, "Modificado", saved.Mono, nil, saved.Reconciliation, clk.Now())
	_ = closingsRepo.Save(ctx, saved)

	// Second approval on version 2
	res, err := apprUC.Execute(ctx, usecase.ApproveClosingRequest{ClosingID: "clo_reopen", Version: 2})
	if err != nil {
		t.Fatalf("reapproval failed: %v", err)
	}
	if res.Closing.Version != 2 {
		t.Errorf("version = %d, want 2", res.Closing.Version)
	}

	notifs, _ := notifRepo.ByClient(ctx, "cli_01")
	if len(notifs) < 2 {
		t.Fatalf("expected at least 2 notifications, got %d", len(notifs))
	}
	if notifs[0].Kind != notification.KindClosingUpdated {
		t.Errorf("expected most recent notification to be CLOSING_UPDATED, got %s", notifs[0].Kind)
	}
}

func TestDispatchOutbox_Retries(t *testing.T) {
	ctx := context.Background()
	_, _, dispatchOutbox, _, _, _, _, _, outboxRepo, sender, clk := setupApprovalSuite(t)

	// Enqueue test message
	msg := notification.OutboundMessage{
		ID:        "msg_fail_test",
		Channel:   notification.ChannelEmail,
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Subject:   "Test",
		Body:      "Body",
		Attempts:  0,
		CreatedAt: clk.Now(),
	}
	_ = outboxRepo.Enqueue(ctx, msg)

	// Case 1: fails 2 times and succeeds on 3rd attempt
	sender.failCount = 2

	// Run 1: attempt 1 fails
	sent, _ := dispatchOutbox.Execute(ctx, 10)
	if sent != 0 {
		t.Errorf("run 1: expected 0 sent, got %d", sent)
	}
	p1, _ := outboxRepo.Pending(ctx, 10)
	if len(p1) != 1 || p1[0].Attempts != 1 || p1[0].LastError == "" {
		t.Errorf("run 1: expected 1 attempt with LastError, got %+v", p1)
	}

	// Run 2: attempt 2 fails
	sent, _ = dispatchOutbox.Execute(ctx, 10)
	if sent != 0 {
		t.Errorf("run 2: expected 0 sent, got %d", sent)
	}
	p2, _ := outboxRepo.Pending(ctx, 10)
	if len(p2) != 1 || p2[0].Attempts != 2 {
		t.Errorf("run 2: expected 2 attempts, got %+v", p2)
	}

	// Run 3: attempt 3 succeeds
	sent, _ = dispatchOutbox.Execute(ctx, 10)
	if sent != 1 {
		t.Errorf("run 3: expected 1 sent, got %d", sent)
	}
	p3, _ := outboxRepo.Pending(ctx, 10)
	if len(p3) != 0 {
		t.Errorf("run 3: expected 0 pending, got %d", len(p3))
	}

	// Case 2: fails 4 times -> reaches 3 attempts and stays with LastError
	sender.calls = 0
	sender.failCount = 10

	msg2 := notification.OutboundMessage{
		ID:        "msg_fail_perm",
		Channel:   notification.ChannelPush,
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Subject:   "Test Perm",
		Body:      "Body",
		Attempts:  0,
		CreatedAt: clk.Now(),
	}
	_ = outboxRepo.Enqueue(ctx, msg2)

	for i := 0; i < 4; i++ {
		_, _ = dispatchOutbox.Execute(ctx, 10)
	}

	// After attempts reach 3, pending returns 0 because Attempts < 3 is required
	pFinal, _ := outboxRepo.Pending(ctx, 10)
	if len(pFinal) != 0 {
		t.Errorf("expected 0 pending after exhausting attempts, got %d", len(pFinal))
	}
}

func TestApproveBatch_ParallelAndRace(t *testing.T) {
	ctx := context.Background()
	_, approveBatch, _, _, _, closingsRepo, clientsRepo, _, _, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)
	feeDue := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	// Create 10 closings with distinct clients:
	// 5 GREEN READY -> OK
	// 1 RED READY -> NOT_GREEN
	// 1 GREEN PENDING -> NOT_READY
	// 1 GREEN READY version mismatch -> VERSION_CONFLICT
	// non-existent -> NOT_FOUND
	for i := 1; i <= 5; i++ {
		cliID := fmt.Sprintf("cli_green_%d", i)
		_ = clientsRepo.Save(ctx, client.Client{ID: cliID, Name: cliID, Regime: client.RegimeMonotributo})
		c := closing.NewClosing(
			fmt.Sprintf("clo_green_%d", i), cliID, p, true, fmt.Sprintf("h%d", i), signal.SeverityGreen,
			nil, "Ok",
			&tax.MonotributoAssessment{
				BilledInPeriodUSD: money.Zero(money.USD),
				BilledInPeriodARS: money.Zero(money.ARS),
				CurrentCategory:   tax.CatA,
				Fee:               money.New(500000, money.ARS),
				FeeDueDate:        feeDue,
			},
			nil, closing.Reconciliation{}, clk.Now(),
		)
		_ = closingsRepo.Save(ctx, c)
	}

	// Red ready
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_red", Name: "Red", Regime: client.RegimeMonotributo})
	_ = closingsRepo.Save(ctx, closing.NewClosing(
		"clo_red", "cli_red", p, true, "h_red", signal.SeverityRed,
		nil, "Red", nil, nil, closing.Reconciliation{}, clk.Now(),
	))

	// Green pending
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_pending", Name: "Pending", Regime: client.RegimeMonotributo})
	_ = closingsRepo.Save(ctx, closing.NewClosing(
		"clo_pending", "cli_pending", p, false, "h_pend", signal.SeverityGreen,
		nil, "Pending", nil, nil, closing.Reconciliation{}, clk.Now(),
	))

	// Version mismatch
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_v2", Name: "V2", Regime: client.RegimeMonotributo})
	_ = closingsRepo.Save(ctx, closing.NewClosing(
		"clo_v2", "cli_v2", p, true, "h_v2", signal.SeverityGreen,
		nil, "V2", nil, nil, closing.Reconciliation{}, clk.Now(),
	))

	req := usecase.ApproveBatchRequest{
		Items: []usecase.ApproveBatchItemRequest{
			{ID: "clo_green_1", Version: 1},
			{ID: "clo_green_2", Version: 1},
			{ID: "clo_green_3", Version: 1},
			{ID: "clo_green_4", Version: 1},
			{ID: "clo_green_5", Version: 1},
			{ID: "clo_red", Version: 1},
			{ID: "clo_pending", Version: 1},
			{ID: "clo_v2", Version: 99},
			{ID: "clo_not_found", Version: 1},
		},
	}

	res, err := approveBatch.Execute(ctx, p, req)
	if err != nil {
		t.Fatalf("ApproveBatch failed: %v", err)
	}

	if res.Approved != 5 {
		t.Errorf("approved = %d, want 5", res.Approved)
	}
	if res.Failed != 4 {
		t.Errorf("failed = %d, want 4", res.Failed)
	}

	// Verify error codes
	errMap := make(map[string]string)
	for _, it := range res.Results {
		if !it.OK {
			errMap[it.ID] = it.Error
		}
	}
	if errMap["clo_red"] != "NOT_GREEN" {
		t.Errorf("clo_red error = %s, want NOT_GREEN", errMap["clo_red"])
	}
	if errMap["clo_pending"] != "NOT_READY" {
		t.Errorf("clo_pending error = %s, want NOT_READY", errMap["clo_pending"])
	}
	if errMap["clo_v2"] != "VERSION_CONFLICT" {
		t.Errorf("clo_v2 error = %s, want VERSION_CONFLICT", errMap["clo_v2"])
	}
	if errMap["clo_not_found"] != "NOT_FOUND" {
		t.Errorf("clo_not_found error = %s, want NOT_FOUND", errMap["clo_not_found"])
	}

	// Max limit > 500 error
	hugeItems := make([]usecase.ApproveBatchItemRequest, 501)
	_, err = approveBatch.Execute(ctx, p, usecase.ApproveBatchRequest{Items: hugeItems})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for batch > 500, got %v", err)
	}
}

func TestCreateActionRequiredNotices(t *testing.T) {
	ctx := context.Background()
	_, _, _, createActionNotices, _, closingsRepo, _, notifRepo, _, _, clk := setupApprovalSuite(t)

	p := period.New(2026, time.October)

	// Pending client input closing
	_ = closingsRepo.Save(ctx, closing.NewClosing(
		"clo_act", "cli_01", p, false, "h1", signal.SeverityYellow,
		nil, "Pending", nil, nil, closing.Reconciliation{}, clk.Now(),
	))

	// First execution creates 1 notice
	count, err := createActionNotices.Execute(ctx, p)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	notifs, _ := notifRepo.ByClient(ctx, "cli_01")
	if len(notifs) != 1 || notifs[0].Kind != notification.KindActionRequired {
		t.Errorf("expected 1 ACTION_REQUIRED notification, got %+v", notifs)
	}

	// Second execution is deduplicated (creates 0)
	count2, err := createActionNotices.Execute(ctx, p)
	if err != nil || count2 != 0 {
		t.Errorf("expected 0 notices on second run, got %d (err: %v)", count2, err)
	}
}

func TestSendPaymentReminders_Borders(t *testing.T) {
	ctx := context.Background()
	_, _, _, _, sendReminders, closingsRepo, clientsRepo, notifRepo, _, _, clk := setupApprovalSuite(t)

	now := clk.Now()
	p := period.FromTime(now)

	// Distinct clients for distinct closings
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_thresh", Name: "Threshold", Regime: client.RegimeMonotributo})
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_exp", Name: "Expired", Regime: client.RegimeMonotributo})
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_far", Name: "Far", Regime: client.RegimeMonotributo})
	_ = clientsRepo.Save(ctx, client.Client{ID: "cli_paid", Name: "Paid", Regime: client.RegimeMonotributo})

	// ReminderDaysBeforeDue = 3
	// Umbral exacto: now + 3 days -> debe recordar
	dueOnThreshold := now.AddDate(0, 0, 3)
	cThreshold := closing.NewClosing(
		"clo_thresh", "cli_thresh", p, true, "h1", signal.SeverityGreen,
		nil, "Ok",
		&tax.MonotributoAssessment{
			CurrentCategory: tax.CatA,
			Fee:             money.New(500000, money.ARS),
			FeeDueDate:      dueOnThreshold,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = cThreshold.Approve(closing.Approval{Version: 1})
	_ = cThreshold.MarkNotified()
	_ = closingsRepo.Save(ctx, cThreshold)

	// Vencido: due < now -> NO debe recordar
	dueExpired := now.AddDate(0, 0, -1)
	cExpired := closing.NewClosing(
		"clo_expired", "cli_exp", p, true, "h2", signal.SeverityGreen,
		nil, "Expired",
		&tax.MonotributoAssessment{
			CurrentCategory: tax.CatA,
			Fee:             money.New(500000, money.ARS),
			FeeDueDate:      dueExpired,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = cExpired.Approve(closing.Approval{Version: 1})
	_ = cExpired.MarkNotified()
	_ = closingsRepo.Save(ctx, cExpired)

	// Lejos: due > now + 3 days -> NO debe recordar
	dueFar := now.AddDate(0, 0, 5)
	cFar := closing.NewClosing(
		"clo_far", "cli_far", p, true, "h3", signal.SeverityGreen,
		nil, "Far",
		&tax.MonotributoAssessment{
			CurrentCategory: tax.CatA,
			Fee:             money.New(500000, money.ARS),
			FeeDueDate:      dueFar,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = cFar.Approve(closing.Approval{Version: 1})
	_ = cFar.MarkNotified()
	_ = closingsRepo.Save(ctx, cFar)

	// Ya pagado: paidAt != nil -> NO debe recordar
	cPaid := closing.NewClosing(
		"clo_paid", "cli_paid", p, true, "h4", signal.SeverityGreen,
		nil, "Paid",
		&tax.MonotributoAssessment{
			CurrentCategory: tax.CatA,
			Fee:             money.New(500000, money.ARS),
			FeeDueDate:      dueOnThreshold,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = cPaid.Approve(closing.Approval{Version: 1})
	_ = cPaid.MarkNotified()
	_ = cPaid.MarkPaid(now)
	_ = closingsRepo.Save(ctx, cPaid)

	// Execute SendPaymentReminders
	reminded, err := sendReminders.Execute(ctx, now)
	if err != nil {
		t.Fatalf("SendPaymentReminders failed: %v", err)
	}

	if reminded != 1 {
		t.Errorf("reminded = %d, want 1 (only the threshold one)", reminded)
	}

	notifs, _ := notifRepo.ByClient(ctx, "cli_thresh")
	if len(notifs) != 1 || notifs[0].Kind != notification.KindPaymentReminder {
		t.Errorf("expected 1 PAYMENT_REMINDER, got %+v", notifs)
	}

	// Segundo run deduplicado: 0
	reminded2, err := sendReminders.Execute(ctx, now)
	if err != nil || reminded2 != 0 {
		t.Errorf("second run reminded = %d, want 0 (err: %v)", reminded2, err)
	}
}

func TestApprovalPublishedSnapshotCriterion3(t *testing.T) {
	approveClosing, _, _, _, _, closingsRepo, _, _, _, _, clk := setupApprovalSuite(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")
	now := clk.Now()

	// Cierre verde inicial
	c := closing.NewClosing(
		"clo_pub", "cli_01", p, true, "hash_v1", signal.SeverityGreen,
		nil, "Green",
		&tax.MonotributoAssessment{
			CurrentCategory:   tax.CatA,
			Fee:               money.New(2660000, money.ARS),
			FeeDueDate:        now.AddDate(0, 0, 10),
			BilledInPeriodUSD: money.New(100000, money.USD),
			BilledInPeriodARS: money.New(120000000, money.ARS),
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = closingsRepo.Save(ctx, c)

	// Aprobar cierre v1
	res, err := approveClosing.Execute(ctx, usecase.ApproveClosingRequest{
		ClosingID:  "clo_pub",
		Version:    1,
		ClientNote: "Nota importante",
	})
	if err != nil {
		t.Fatalf("error al aprobar: %v", err)
	}

	saved, _ := closingsRepo.Get(ctx, "clo_pub")
	if saved.Published == nil {
		t.Fatal("Published no debe ser nil tras aprobación")
	}
	if saved.Published.Version != 1 {
		t.Fatalf("Published.Version = %d, want 1", saved.Published.Version)
	}
	if saved.Published.Figures.Total.Amount != 2660000 {
		t.Fatalf("Published.Figures.Total = %d, want 2660000", saved.Published.Figures.Total.Amount)
	}
	if res.Closing.Status != closing.StatusNotified {
		t.Fatalf("status = %s, want NOTIFIED", res.Closing.Status)
	}

	// Reabrir por nuevo cálculo (ApplyResult cambia hash)
	later := now.Add(24 * time.Hour)
	saved.ApplyResult("hash_v2", signal.SeverityYellow, nil, "Recalculated",
		&tax.MonotributoAssessment{
			CurrentCategory:   tax.CatB,
			Fee:               money.New(3500000, money.ARS), // Nueva cuota no aprobada
			FeeDueDate:        later.AddDate(0, 0, 10),
			BilledInPeriodUSD: money.New(200000, money.USD),
			BilledInPeriodARS: money.New(240000000, money.ARS),
		},
		nil, closing.Reconciliation{}, later)
	_ = closingsRepo.Save(ctx, saved)

	// Verificar Criterio 3: El cierre reabierto mantiene el snapshot Published viejo
	reopened, _ := closingsRepo.Get(ctx, "clo_pub")
	if reopened.Status != closing.StatusReadyForReview {
		t.Fatalf("status = %s, want READY_FOR_REVIEW", reopened.Status)
	}
	if reopened.Published == nil {
		t.Fatal("Published no debe ser nil al reabrirse")
	}
	if reopened.Published.Version != 1 {
		t.Fatalf("reopened.Published.Version = %d, want 1", reopened.Published.Version)
	}
	// La cifra publicada sigue siendo 2660000, a pesar de que el cálculo vigente tiene 3500000
	if reopened.Published.Figures.Total.Amount != 2660000 {
		t.Fatalf("Published.Figures.Total debió mantenerse en 2660000, got %d", reopened.Published.Figures.Total.Amount)
	}
	if reopened.Mono.Fee.Amount != 3500000 {
		t.Fatalf("reopened.Mono.Fee (cálculo en vivo) debió ser 3500000, got %d", reopened.Mono.Fee.Amount)
	}
}

type barrierClientRepo struct {
	*fakes.ClientRepository
	wg sync.WaitGroup
}

func (b *barrierClientRepo) Get(ctx context.Context, id string) (client.Client, error) {
	b.wg.Done()
	b.wg.Wait()
	return b.ClientRepository.Get(ctx, id)
}

func TestApproveClosing_ConcurrentSameVersion(t *testing.T) {
	ctx := context.Background()
	_, _, _, _, _, closingsRepo, clientsRepo, notifRepo, outboxRepo, sender, clk := setupApprovalSuite(t)

	c := closing.NewClosing(
		"clo_conc", "cli_01", period.MustParse("2026-10"), true, "hash_conc", signal.SeverityGreen,
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

	barrierClients := &barrierClientRepo{ClientRepository: clientsRepo}
	barrierClients.wg.Add(2)

	apprUC := &usecase.ApproveClosing{
		Closings:      closingsRepo,
		Clients:       barrierClients,
		Notifications: notifRepo,
		Outbox:        outboxRepo,
		Sender:        sender,
		Clock:         clk,
		IDGen:         fakes.NewIDGenerator(),
	}

	errs := make([]error, 2)
	var done sync.WaitGroup
	done.Add(2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			defer done.Done()
			_, errs[idx] = apprUC.Execute(ctx, usecase.ApproveClosingRequest{
				ClosingID: "clo_conc",
				Version:   1,
			})
		}(i)
	}
	done.Wait()

	okCount := 0
	conflictCount := 0
	for _, err := range errs {
		if err == nil {
			okCount++
		} else if errors.Is(err, domain.ErrVersionConflict) {
			conflictCount++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if okCount != 1 || conflictCount != 1 {
		t.Fatalf("expected 1 ok and 1 ErrVersionConflict, got ok=%d conflict=%d errs=%v", okCount, conflictCount, errs)
	}
}
