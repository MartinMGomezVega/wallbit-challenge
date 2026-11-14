package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestRecalcOnChange(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p10 := period.MustParse("2026-10")
	p11 := period.MustParse("2026-11")

	cuit, _ := client.ParseCUIT("20-12345678-9")
	cli := client.Client{
		ID:                  "cli_recalc",
		Name:                "Recalc Client",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// Crear cierres previos para p10 y p11
	c10 := closing.NewClosing("clo_10", cli.ID, p10, false, "h10", 0, nil, "", nil, nil, closing.Reconciliation{}, now)
	c11 := closing.NewClosing("clo_11", cli.ID, p11, false, "h11", 0, nil, "", nil, nil, closing.Reconciliation{}, now)
	_ = h.closings.Save(ctx, c10)
	_ = h.closings.Save(ctx, c11)

	recalc := &usecase.RecalcOnChange{
		Compute:  h.uc,
		Closings: h.closings,
	}

	// Notificar cambio en p10
	err := recalc.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
		ClientID: cli.ID,
		Period:   p10,
	})
	if err != nil {
		t.Fatalf("error en NotifyClosingInputsChanged: %v", err)
	}

	// Verificar que ambos cierres (p10 y p11) fueron recalculados (actualizados)
	after10, found, err := h.closings.ByClientPeriod(ctx, cli.ID, p10)
	if err != nil || !found {
		t.Fatalf("cierre p10 no encontrado: %v", err)
	}
	if after10.ComputedAt.IsZero() {
		t.Fatal("cierre p10 debió ser recalculado")
	}

	after11, found, err := h.closings.ByClientPeriod(ctx, cli.ID, p11)
	if err != nil || !found {
		t.Fatalf("cierre p11 no encontrado: %v", err)
	}
	if after11.ComputedAt.IsZero() {
		t.Fatal("cierre p11 (período siguiente) debió ser recalculado")
	}
}

func TestRecalcPeriod(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	// Crear cliente 1 (no aprobado)
	cuit1, _ := client.ParseCUIT("20-11111111-2")
	cli1 := client.Client{
		ID:                  "cli_1",
		Name:                "Cliente 1",
		CUIT:                cuit1,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli1)
	clo1 := closing.NewClosing("clo_1", "cli_1", p, false, "h1", 0, nil, "", nil, nil, closing.Reconciliation{}, now)
	_ = h.closings.Save(ctx, clo1)

	// Crear cliente 2 (aprobado)
	cuit2, _ := client.ParseCUIT("20-22222222-3")
	cli2 := client.Client{
		ID:                  "cli_2",
		Name:                "Cliente 2",
		CUIT:                cuit2,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli2)
	clo2 := closing.NewClosing("clo_2", "cli_2", p, true, "h2", 0, nil, "", nil, nil, closing.Reconciliation{}, now)
	_ = clo2.Approve(closing.Approval{ID: "app_1", ApprovedBy: "contador", ApprovedAt: now})
	_ = h.closings.Save(ctx, clo2)

	recalcPeriod := &usecase.RecalcPeriod{
		Closings: h.closings,
		Compute:  h.uc,
	}

	count, err := recalcPeriod.Execute(ctx, p)
	if err != nil {
		t.Fatalf("error en RecalcPeriod: %v", err)
	}
	// Solo 1 cierre (cli_1) debe ser recalculado; cli_2 está aprobado y se ignora
	if count != 1 {
		t.Fatalf("esperaba 1 cierre recalculado, obtuvo %d", count)
	}

	// Verificar que cli_2 sigue aprobado
	after2, _, _ := h.closings.ByClientPeriod(ctx, "cli_2", p)
	if after2.Status != closing.StatusApproved {
		t.Fatalf("cierre 2 debió permanecer APPROVED, obtuvo %s", after2.Status)
	}
}

func TestConfirmPeriodInput_MovesPendingToReady(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	cuit, _ := client.ParseCUIT("20-11111111-2")
	cli := client.Client{
		ID:                  "cli_conf",
		Name:                "Cliente Confirm",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// Cierre en PENDING_CLIENT_INPUT
	resInit, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en Execute inicial: %v", err)
	}
	if resInit.Closing.Status != closing.StatusPendingClientInput {
		t.Fatalf("esperaba PENDING_CLIENT_INPUT, obtuvo %s", resInit.Closing.Status)
	}

	// Configurar ChangeNotifier real
	recalcNotifier := &usecase.RecalcOnChange{
		Compute:  h.uc,
		Closings: h.closings,
	}

	confirmUC := &usecase.ConfirmPeriodInput{
		Inputs:   h.clientInputs,
		Clock:    h.clock,
		Notifier: recalcNotifier,
	}

	// Guardar input contestado con "sin ingresos externos"
	_ = h.clientInputs.Save(ctx, clientinput.ClientPeriodInput{
		ClientID:                 cli.ID,
		Period:                   p,
		DeclaredNoExternalIncome: true,
	})

	// Ejecutar confirmación del cliente
	_, err = confirmUC.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error al confirmar input: %v", err)
	}

	// Verificar que el cierre pasó automáticamente a READY_FOR_REVIEW
	updatedClo, found, err := h.closings.ByClientPeriod(ctx, cli.ID, p)
	if err != nil || !found {
		t.Fatalf("cierre no encontrado: %v", err)
	}
	if updatedClo.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba transición a READY_FOR_REVIEW tras confirmar input, obtuvo %s", updatedClo.Status)
	}
}

func TestUpdateParameters_TriggersRecalc(t *testing.T) {
	ctx := context.Background()
	// Fecha actual: 2026-11-05 -> último mes cerrado: 2026-10
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p10 := period.MustParse("2026-10")

	cuit, _ := client.ParseCUIT("20-11111111-2")
	cli := client.Client{
		ID:                  "cli_param",
		Name:                "Cliente Params",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// Crear cierre en 2026-10 sin aprobar
	c10 := closing.NewClosing("clo_10", cli.ID, p10, false, "h10", 0, nil, "", nil, nil, closing.Reconciliation{}, now)
	_ = h.closings.Save(ctx, c10)

	recalcPeriod := &usecase.RecalcPeriod{
		Closings: h.closings,
		Compute:  h.uc,
	}

	updateParamsUC := &usecase.UpdateParameters{
		Repo:         h.params,
		Clock:        h.clock,
		RecalcPeriod: recalcPeriod,
	}

	newParams := h.params.Params
	newParams.CategoryCapWarningBp = money.BasisPoints(8500)

	res, err := updateParamsUC.Execute(ctx, newParams)
	if err != nil {
		t.Fatalf("error en UpdateParameters: %v", err)
	}
	if res.Recalculated != 1 {
		t.Fatalf("esperaba 1 cierre recalculado por cambio de parámetros, obtuvo %d", res.Recalculated)
	}
}
