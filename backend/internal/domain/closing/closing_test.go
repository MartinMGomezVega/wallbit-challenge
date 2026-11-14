package closing_test

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

func TestNewClosing(t *testing.T) {
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	// Unconfirmed input -> PENDING_CLIENT_INPUT
	c1 := closing.NewClosing(
		"clo_1", "cli_1", p, false, "hash1",
		signal.SeverityGreen, nil, "Todo OK",
		nil, nil, closing.Reconciliation{}, now,
	)
	if c1.Status != closing.StatusPendingClientInput {
		t.Fatalf("esperaba PENDING_CLIENT_INPUT, obtuvo %s", c1.Status)
	}
	if c1.Version != 1 {
		t.Fatalf("esperaba versión 1, obtuvo %d", c1.Version)
	}

	// Confirmed input -> READY_FOR_REVIEW
	c2 := closing.NewClosing(
		"clo_2", "cli_2", p, true, "hash1",
		signal.SeverityGreen, nil, "Todo OK",
		nil, nil, closing.Reconciliation{}, now,
	)
	if c2.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba READY_FOR_REVIEW, obtuvo %s", c2.Status)
	}
}

func TestMarkReadyTransitions(t *testing.T) {
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_1", "cli_1", p, false, "hash1",
		signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)

	// PENDING_CLIENT_INPUT -> READY_FOR_REVIEW
	if err := c.MarkReady(); err != nil {
		t.Fatalf("esperaba éxito en MarkReady, obtuvo: %v", err)
	}
	if c.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba READY_FOR_REVIEW, obtuvo %s", c.Status)
	}

	// READY_FOR_REVIEW -> READY_FOR_REVIEW (idempotente)
	if err := c.MarkReady(); err != nil {
		t.Fatalf("esperaba éxito en MarkReady idempotente, obtuvo: %v", err)
	}

	// Pasar a APPROVED
	app := closing.Approval{
		ID:         "app_1",
		ApprovedBy: "contador",
		ApprovedAt: now,
		Note:       "ok",
	}
	if err := c.Approve(app); err != nil {
		t.Fatalf("error al aprobar: %v", err)
	}

	// APPROVED -> MarkReady debe fallar
	if err := c.MarkReady(); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition desde APPROVED, obtuvo: %v", err)
	}

	// Pasar a NOTIFIED
	if err := c.MarkNotified(); err != nil {
		t.Fatalf("error en MarkNotified: %v", err)
	}

	// NOTIFIED -> MarkReady debe fallar
	if err := c.MarkReady(); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition desde NOTIFIED, obtuvo: %v", err)
	}
}

func TestApproveTransitions(t *testing.T) {
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	// Aprobar directamente desde PENDING_CLIENT_INPUT (contador decide aprobar antes de que el cliente confirme)
	cPending := closing.NewClosing(
		"clo_1", "cli_1", p, false, "hash1",
		signal.SeverityYellow, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)
	app1 := closing.Approval{ID: "app_1", ApprovedBy: "c1", ApprovedAt: now}
	if err := cPending.Approve(app1); err != nil {
		t.Fatalf("se debe poder aprobar desde PENDING_CLIENT_INPUT: %v", err)
	}
	if cPending.Status != closing.StatusApproved {
		t.Fatalf("esperaba APPROVED, obtuvo %s", cPending.Status)
	}
	if len(cPending.Approvals) != 1 || cPending.Approvals[0].Version != 1 || cPending.Approvals[0].ResultHash != "hash1" {
		t.Fatalf("approval no quedó registrado correctamente: %+v", cPending.Approvals)
	}

	// Intentar aprobar cuando ya está APPROVED -> debe fallar
	if err := cPending.Approve(closing.Approval{ID: "app_2"}); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition al re-aprobar APPROVED: %v", err)
	}

	// Pasar a NOTIFIED
	if err := cPending.MarkNotified(); err != nil {
		t.Fatalf("error en MarkNotified: %v", err)
	}

	// Intentar aprobar cuando está NOTIFIED -> debe fallar
	if err := cPending.Approve(closing.Approval{ID: "app_3"}); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition al aprobar NOTIFIED: %v", err)
	}
}

func TestMarkNotifiedTransitions(t *testing.T) {
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_1", "cli_1", p, false, "hash1",
		signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)

	// Desde PENDING_CLIENT_INPUT -> debe fallar
	if err := c.MarkNotified(); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition desde PENDING_CLIENT_INPUT, obtuvo: %v", err)
	}

	// Desde READY_FOR_REVIEW -> debe fallar
	_ = c.MarkReady()
	if err := c.MarkNotified(); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition desde READY_FOR_REVIEW, obtuvo: %v", err)
	}

	// Desde APPROVED -> éxito
	_ = c.Approve(closing.Approval{ID: "app_1", ApprovedBy: "c1", ApprovedAt: now})
	if err := c.MarkNotified(); err != nil {
		t.Fatalf("esperaba éxito en MarkNotified desde APPROVED, obtuvo: %v", err)
	}
	if c.Status != closing.StatusNotified {
		t.Fatalf("esperaba NOTIFIED, obtuvo %s", c.Status)
	}

	// Desde NOTIFIED -> debe fallar
	if err := c.MarkNotified(); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("esperaba ErrInvalidTransition desde NOTIFIED, obtuvo: %v", err)
	}
}

func TestMarkPaid(t *testing.T) {
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_1", "cli_1", p, false, "hash1",
		signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)

	paidDate := time.Date(2026, 11, 20, 14, 0, 0, 0, period.Location())
	if err := c.MarkPaid(paidDate); err != nil {
		t.Fatalf("error en MarkPaid: %v", err)
	}
	if c.PaidAt == nil || !c.PaidAt.Equal(paidDate) {
		t.Fatalf("PaidAt no coincide: got %v, want %v", c.PaidAt, paidDate)
	}
}

func TestApplyResult(t *testing.T) {
	p := period.MustParse("2026-10")
	t1 := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())
	t2 := time.Date(2026, 11, 2, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_1", "cli_1", p, true, "hash_v1",
		signal.SeverityGreen, nil, "desc v1", nil, nil, closing.Reconciliation{}, t1,
	)

	// Mismo hash -> false, no incrementa versión
	updated := c.ApplyResult("hash_v1", signal.SeverityGreen, nil, "desc v1", nil, nil, closing.Reconciliation{}, t2)
	if updated {
		t.Fatal("esperaba false para mismo hash")
	}
	if c.Version != 1 {
		t.Fatalf("esperaba versión 1, obtuvo %d", c.Version)
	}
	if !c.ComputedAt.Equal(t2) {
		t.Fatalf("ComputedAt debió actualizarse a t2")
	}

	// Cambio de hash en READY_FOR_REVIEW -> versión 2, queda READY_FOR_REVIEW
	updated = c.ApplyResult("hash_v2", signal.SeverityYellow, nil, "desc v2", nil, nil, closing.Reconciliation{}, t2)
	if !updated {
		t.Fatal("esperaba true para hash modificado")
	}
	if c.Version != 2 {
		t.Fatalf("esperaba versión 2, obtuvo %d", c.Version)
	}
	if c.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba READY_FOR_REVIEW, obtuvo %s", c.Status)
	}

	// Aprobar versión 2
	app := closing.Approval{ID: "app_v2", ApprovedBy: "c1", ApprovedAt: t2}
	_ = c.Approve(app)
	if c.Status != closing.StatusApproved {
		t.Fatalf("esperaba APPROVED, obtuvo %s", c.Status)
	}

	// Cambio de hash en APPROVED -> versión 3, reabre a READY_FOR_REVIEW y preserva historial de aprobaciones
	t3 := time.Date(2026, 11, 3, 0, 0, 0, 0, period.Location())
	updated = c.ApplyResult("hash_v3", signal.SeverityRed, nil, "desc v3", nil, nil, closing.Reconciliation{}, t3)
	if !updated {
		t.Fatal("esperaba true al modificar hash en APPROVED")
	}
	if c.Version != 3 {
		t.Fatalf("esperaba versión 3, obtuvo %d", c.Version)
	}
	if c.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba reapertura a READY_FOR_REVIEW, obtuvo %s", c.Status)
	}
	if len(c.Approvals) != 1 || c.Approvals[0].Version != 2 {
		t.Fatalf("el historial de aprobaciones no se preservó: %+v", c.Approvals)
	}

	// Aprobar y notificar versión 3
	_ = c.Approve(closing.Approval{ID: "app_v3", ApprovedBy: "c1", ApprovedAt: t3})
	_ = c.MarkNotified()
	if c.Status != closing.StatusNotified {
		t.Fatalf("esperaba NOTIFIED, obtuvo %s", c.Status)
	}

	// Cambio de hash en NOTIFIED -> versión 4, reabre a READY_FOR_REVIEW y preserva aprobaciones
	t4 := time.Date(2026, 11, 4, 0, 0, 0, 0, period.Location())
	updated = c.ApplyResult("hash_v4", signal.SeverityYellow, nil, "desc v4", nil, nil, closing.Reconciliation{}, t4)
	if !updated {
		t.Fatal("esperaba true al modificar hash en NOTIFIED")
	}
	if c.Version != 4 {
		t.Fatalf("esperaba versión 4, obtuvo %d", c.Version)
	}
	if c.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba reapertura a READY_FOR_REVIEW desde NOTIFIED, obtuvo %s", c.Status)
	}
	if len(c.Approvals) != 2 {
		t.Fatalf("esperaba 2 aprobaciones preservadas, obtuvo %d", len(c.Approvals))
	}
}

func TestPublishedPreservedOnReopen(t *testing.T) {
	p := period.MustParse("2026-10")
	t1 := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	c := closing.NewClosing(
		"clo_1", "cli_1", p, true, "hash_v1",
		signal.SeverityGreen, nil, "desc v1", nil, nil, closing.Reconciliation{}, t1,
	)

	// Aprobar
	app := closing.Approval{ID: "app_1", ApprovedBy: "c1", ApprovedAt: t1}
	if err := c.Approve(app); err != nil {
		t.Fatalf("error al aprobar: %v", err)
	}

	// Publicar (011 Criterio 3)
	pub1 := closing.Published{
		Version:     1,
		Title:       "Cierre de Octubre 2026 listo",
		Body:        "Tu cuota de monotributo es $26.600",
		PublishedAt: t1,
	}
	if err := c.Publish(pub1); err != nil {
		t.Fatalf("error al publicar: %v", err)
	}

	if c.Published == nil || c.Published.Version != 1 || c.Published.Title != "Cierre de Octubre 2026 listo" {
		t.Fatalf("c.Published no se guardó correctamente: %+v", c.Published)
	}

	// Reabrir por nuevo cálculo (ApplyResult con hash diferente)
	t2 := time.Date(2026, 11, 5, 0, 0, 0, 0, period.Location())
	updated := c.ApplyResult("hash_v2", signal.SeverityYellow, nil, "desc v2", nil, nil, closing.Reconciliation{}, t2)
	if !updated {
		t.Fatal("esperaba true para ApplyResult con nuevo hash")
	}

	// Estado reabierto
	if c.Status != closing.StatusReadyForReview {
		t.Fatalf("esperaba READY_FOR_REVIEW tras cambio, obtuvo %s", c.Status)
	}
	if c.Version != 2 {
		t.Fatalf("esperaba versión 2, obtuvo %d", c.Version)
	}

	// Criterio 3: El cliente sigue viendo el Published viejo (versión 1) y NO el cálculo vigente sin aprobar
	if c.Published == nil {
		t.Fatal("c.Published no debe ser nil al reabrirse")
	}
	if c.Published.Version != 1 {
		t.Fatalf("c.Published.Version debió mantenerse en 1, obtuvo %d", c.Published.Version)
	}
	if c.Published.Title != "Cierre de Octubre 2026 listo" {
		t.Fatalf("c.Published.Title debió mantenerse intacto, obtuvo %s", c.Published.Title)
	}
}
