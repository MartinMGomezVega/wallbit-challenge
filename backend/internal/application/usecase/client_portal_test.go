package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestClientPortal_GetMe(t *testing.T) {
	ctx := context.Background()
	clientsRepo := fakes.NewClientRepository()

	cuit, _ := client.ParseCUIT("20-12345678-6")
	cliMono := client.Client{
		ID:                  "cli_01",
		Name:                "Paula Rojas",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
	}
	cliRI := client.Client{
		ID:     "cli_02",
		Name:   "Lucía Méndez",
		CUIT:   cuit,
		Regime: client.RegimeRI,
	}
	_ = clientsRepo.SaveAll(ctx, []client.Client{cliMono, cliRI})

	uc := &usecase.GetMe{Clients: clientsRepo}

	// 1. Monotributo con categoría y CUIT enmascarado
	me1, err := uc.Execute(ctx, "cli_01")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if me1.ID != "cli_01" || me1.Name != "Paula Rojas" {
		t.Errorf("datos de cliente incorrectos: %+v", me1)
	}
	if me1.CUITMasked != "20-****5678-6" {
		t.Errorf("CUITMasked = %s, want 20-****5678-6", me1.CUITMasked)
	}
	if me1.Regime != "MONOTRIBUTO" || me1.Category != "A" {
		t.Errorf("regimen/categoria incorrectos: %+v", me1)
	}

	// 2. RI sin categoría
	me2, err := uc.Execute(ctx, "cli_02")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if me2.Regime != "RI" || me2.Category != "" {
		t.Errorf("esperaba RI sin categoría, got %+v", me2)
	}

	// 3. Cliente inexistente -> 404
	_, err = uc.Execute(ctx, "non_existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperaba ErrNotFound, got %v", err)
	}
}

func TestClientPortal_ListMyClosings(t *testing.T) {
	ctx := context.Background()
	closingsRepo := fakes.NewClosingRepository()
	clientInputsRepo := fakes.NewClientInputRepository()
	invoicesRepo := fakes.NewInvoiceRepository()

	pOct := period.MustParse("2026-10")
	pSep := period.MustParse("2026-09")
	now := time.Date(2026, 11, 1, 10, 0, 0, 0, period.Location())

	// 1. Cierre de Sep: Notificado y listo (READY)
	cSep := closing.NewClosing(
		"clo_sep", "cli_paula", pSep, true, "h_sep", signal.SeverityGreen,
		nil, "Sep OK", nil, nil, closing.Reconciliation{}, now,
	)
	_ = cSep.Approve(closing.Approval{Version: 1})
	pubSep := closing.Published{
		Version:     1,
		Title:       "Cierre Septiembre listo",
		Body:        "Detalle",
		PublishedAt: now,
	}
	_ = cSep.Publish(pubSep)
	paidAt := now.Add(24 * time.Hour)
	_ = cSep.MarkPaid(paidAt)
	_ = closingsRepo.Save(ctx, cSep)

	inputSep := clientinput.ClientPeriodInput{
		ClientID:                 "cli_paula",
		Period:                   pSep,
		DeclaredNoExternalIncome: true,
		ConfirmedAt:              &now,
	}
	_ = clientInputsRepo.Save(ctx, inputSep)

	// 2. Cierre de Oct: Sin confirmar input (ACTION_REQUIRED)
	cOct := closing.NewClosing(
		"clo_oct", "cli_paula", pOct, false, "h_oct", signal.SeverityYellow,
		nil, "Oct pending", nil, nil, closing.Reconciliation{}, now,
	)
	_ = closingsRepo.Save(ctx, cOct)

	// Factura en Octubre
	inv := invoice.Invoice{
		ID:          "inv_1",
		ClientID:    "cli_paula",
		Type:        invoice.TypeC,
		PointOfSale: 1,
		Number:      1,
		IssueDate:   time.Date(2026, 10, 15, 12, 0, 0, 0, period.Location()),
		NetAmount:   money.New(100000, money.ARS),
		TotalAmount: money.New(100000, money.ARS),
	}
	_ = invoicesRepo.Save(ctx, inv)

	uc := &usecase.ListMyClosings{
		Closings:     closingsRepo,
		ClientInputs: clientInputsRepo,
		Invoices:     invoicesRepo,
	}

	res, err := uc.Execute(ctx, "cli_paula")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("esperaba 2 cierres, got %d", len(res))
	}

	// Más reciente primero (Octubre)
	octDTO := res[0]
	if octDTO.Period != "2026-10" || octDTO.State != "ACTION_REQUIRED" {
		t.Errorf("Octubre state = %s, want ACTION_REQUIRED", octDTO.State)
	}
	if octDTO.Input.InvoicesCount != 1 {
		t.Errorf("Octubre InvoicesCount = %d, want 1", octDTO.Input.InvoicesCount)
	}
	if octDTO.Input.Confirmed {
		t.Errorf("Octubre Confirmed debió ser false")
	}

	// Septiembre
	sepDTO := res[1]
	if sepDTO.Period != "2026-09" || sepDTO.State != "READY" {
		t.Errorf("Septiembre state = %s, want READY", sepDTO.State)
	}
	if sepDTO.Published == nil || sepDTO.Published.Title != "Cierre Septiembre listo" {
		t.Errorf("Septiembre Published = %+v", sepDTO.Published)
	}
	if sepDTO.PaidAt == nil || !sepDTO.PaidAt.Equal(paidAt) {
		t.Errorf("Septiembre PaidAt = %v, want %v", sepDTO.PaidAt, paidAt)
	}

	// 3. Caso IN_REVIEW con reapertura (updating: true)
	// Confirmamos input en Octubre y simulamos que tenía un Published anterior pero fue reabierto
	cOct.Published = &closing.Published{Version: 1, Title: "Snapshot viejo"}
	cOct.Status = closing.StatusReadyForReview
	_ = closingsRepo.Save(ctx, cOct)
	inputOct := clientinput.ClientPeriodInput{
		ClientID:                 "cli_paula",
		Period:                   pOct,
		DeclaredNoExternalIncome: true,
		ConfirmedAt:              &now,
	}
	_ = clientInputsRepo.Save(ctx, inputOct)

	resAfter, err := uc.Execute(ctx, "cli_paula")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resAfter[0].State != "IN_REVIEW" {
		t.Errorf("Octubre reabierto state = %s, want IN_REVIEW", resAfter[0].State)
	}
	if !resAfter[0].Updating {
		t.Errorf("Octubre reabierto debió tener updating = true")
	}
	if resAfter[0].Published == nil || resAfter[0].Published.Title != "Snapshot viejo" {
		t.Errorf("Octubre reabierto debe mostrar snapshot viejo al cliente")
	}
}

func TestClientPortal_MarkClosingPaid(t *testing.T) {
	ctx := context.Background()
	closingsRepo := fakes.NewClosingRepository()
	clk := &fakes.Clock{Current: time.Date(2026, 11, 2, 10, 0, 0, 0, period.Location())}

	p := period.MustParse("2026-10")
	pSep := period.MustParse("2026-09")

	// Cierre listo (READY) de cli_01 (Octubre)
	cReady := closing.NewClosing(
		"clo_ready", "cli_01", p, true, "h1", signal.SeverityGreen,
		nil, "desc", nil, nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = cReady.Approve(closing.Approval{Version: 1})
	pub := closing.Published{Version: 1, Title: "Listo"}
	_ = cReady.Publish(pub)
	_ = closingsRepo.Save(ctx, cReady)

	// Cierre no listo (PENDING_CLIENT_INPUT) de cli_01 (Septiembre)
	cNotReady := closing.NewClosing(
		"clo_not_ready", "cli_01", pSep, false, "h2", signal.SeverityYellow,
		nil, "desc", nil, nil, closing.Reconciliation{}, clk.Now(),
	)
	_ = closingsRepo.Save(ctx, cNotReady)

	uc := &usecase.MarkClosingPaid{
		Closings: closingsRepo,
		Clock:    clk,
	}

	// 1. Intentar pagar cierre que pertenece a otro cliente -> 403
	_, err := uc.Execute(ctx, "clo_ready", "cli_attacker")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperaba ErrForbidden, got %v", err)
	}

	// 2. Intentar pagar cierre inexistente -> 404
	_, err = uc.Execute(ctx, "clo_inexistente", "cli_01")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperaba ErrNotFound, got %v", err)
	}

	// 3. Intentar pagar cierre no READY -> 409
	_, err = uc.Execute(ctx, "clo_not_ready", "cli_01")
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("esperaba ErrConflict, got %v", err)
	}

	// 4. Pagar cierre READY -> éxito
	paidClosing, err := uc.Execute(ctx, "clo_ready", "cli_01")
	if err != nil {
		t.Fatalf("error pagando cierre: %v", err)
	}
	if paidClosing.PaidAt == nil || !paidClosing.PaidAt.Equal(clk.Now()) {
		t.Errorf("paidAt incorrecto: got %v, want %v", paidClosing.PaidAt, clk.Now())
	}

	// 5. Idempotencia: segundo llamado mantiene la fecha original
	clk.Current = clk.Current.Add(1 * time.Hour)
	paidClosing2, err := uc.Execute(ctx, "clo_ready", "cli_01")
	if err != nil {
		t.Fatalf("error en llamada idempotente: %v", err)
	}
	if !paidClosing2.PaidAt.Equal(clk.Now().Add(-1 * time.Hour)) {
		t.Errorf("llamada idempotente modificó PaidAt")
	}
}

func TestClientPortal_Notifications(t *testing.T) {
	ctx := context.Background()
	notifRepo := fakes.NewNotificationRepository()
	clk := &fakes.Clock{Current: time.Date(2026, 11, 2, 12, 0, 0, 0, period.Location())}

	n1 := notification.Notification{
		ID:        "notif_1",
		ClientID:  "cli_01",
		Kind:      notification.KindClosingReady,
		Title:     "Aviso 1",
		CreatedAt: clk.Now().Add(-2 * time.Hour),
	}
	n2 := notification.Notification{
		ID:        "notif_2",
		ClientID:  "cli_01",
		Kind:      notification.KindClosingReady,
		Title:     "Aviso 2",
		CreatedAt: clk.Now().Add(-1 * time.Hour),
	}
	nOther := notification.Notification{
		ID:        "notif_other",
		ClientID:  "cli_other",
		Kind:      notification.KindClosingReady,
		Title:     "Aviso Other",
		CreatedAt: clk.Now(),
	}
	_ = notifRepo.Save(ctx, n1)
	_ = notifRepo.Save(ctx, n2)
	_ = notifRepo.Save(ctx, nOther)

	listUC := &usecase.ListMyNotifications{Notifications: notifRepo}
	readUC := &usecase.MarkNotificationRead{Notifications: notifRepo, Clock: clk}

	// 1. Listar notificaciones de cli_01 ordenadas descendentemente
	notifs, err := listUC.Execute(ctx, "cli_01")
	if err != nil {
		t.Fatalf("error listando: %v", err)
	}
	if len(notifs) != 2 {
		t.Fatalf("esperaba 2 notificaciones, got %d", len(notifs))
	}
	if notifs[0].ID != "notif_2" || notifs[1].ID != "notif_1" {
		t.Errorf("orden incorrecto: got %+v", notifs)
	}

	// 2. Intentar marcar notificación de otro cliente -> 403
	err = readUC.Execute(ctx, "notif_other", "cli_01")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("esperaba ErrForbidden al marcar notificación ajena, got %v", err)
	}

	// 3. Intentar marcar notificación inexistente -> 404
	err = readUC.Execute(ctx, "notif_inexistente", "cli_01")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("esperaba ErrNotFound, got %v", err)
	}

	// 4. Marcar propia -> éxito
	err = readUC.Execute(ctx, "notif_2", "cli_01")
	if err != nil {
		t.Fatalf("error marcando leída: %v", err)
	}

	updated, _ := notifRepo.Get(ctx, "notif_2")
	if updated.ReadAt == nil || !updated.ReadAt.Equal(clk.Now()) {
		t.Errorf("ReadAt incorrecto: got %v, want %v", updated.ReadAt, clk.Now())
	}
}
