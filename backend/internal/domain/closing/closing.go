// Package closing implementa las entidades, la máquina de estados y las reglas de negocio
// del cierre contable mensual según domain-context.md §10 y architecture.md §2.2.
package closing

import (
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ErrInvalidTransition indica una transición de estado no permitida en la máquina de estados.
var ErrInvalidTransition = domain.ErrInvalidTransition

// ErrVersionConflict indica una versión inconsistente al intentar aprobar un cierre (concurrencia optimista).
var ErrVersionConflict = domain.ErrVersionConflict

// ErrAlreadyApproved indica que el cierre ya fue aprobado o notificado en la misma versión.
var ErrAlreadyApproved = domain.ErrAlreadyApproved

// Status representa los estados posibles de un cierre mensual.
type Status string

const (
	StatusPendingClientInput Status = "PENDING_CLIENT_INPUT"
	StatusReadyForReview     Status = "READY_FOR_REVIEW"
	StatusApproved           Status = "APPROVED"
	StatusNotified           Status = "NOTIFIED"
)

// Approval registra la aprobación formal realizada por un contador.
type Approval struct {
	ID         string    `json:"id"`
	Version    int       `json:"version"`
	ResultHash string    `json:"resultHash"`
	ApprovedBy string    `json:"approvedBy"`
	ApprovedAt time.Time `json:"approvedAt"`
	Note       string    `json:"note,omitempty"`
	Batch      bool      `json:"batch"`
}

// MatchedPair asocia un movimiento con una factura y la diferencia de monto.
type MatchedPair struct {
	MovementID string `json:"movementId"`
	InvoiceID  string `json:"invoiceId"`
	DiffAmount int64  `json:"diffAmount"`
}

// Reconciliation contiene los IDs de la conciliación del período.
type Reconciliation struct {
	MatchedPairs        []MatchedPair `json:"matchedPairs"`
	UnmatchedIncomeIDs  []string      `json:"unmatchedIncomeIds"`
	UnmatchedInvoiceIDs []string      `json:"unmatchedInvoiceIds"`
}

// PublishedFigureItem representa un concepto individual a pagar publicado al cliente.
type PublishedFigureItem struct {
	Concept string      `json:"concept"`
	Amount  money.Money `json:"amount"`
	DueDate time.Time   `json:"dueDate"`
}

// PublishedFigures agrupa las cifras publicadas de facturación y conceptos a pagar.
type PublishedFigures struct {
	BilledUSD money.Money           `json:"billedUSD"`
	BilledARS money.Money           `json:"billedARS"`
	Items     []PublishedFigureItem `json:"items"`
	Total     money.Money           `json:"total"`
}

// Published contiene la información oficial congelada que el cliente ve una vez notificado el cierre (011 §3).
type Published struct {
	Version     int              `json:"version"`
	Title       string           `json:"title"`
	Body        string           `json:"body"`
	Figures     PublishedFigures `json:"figures"`
	PublishedAt time.Time        `json:"publishedAt"`
}

// Closing es la entidad raíz que agrupa toda la información de un cierre contable mensual.
type Closing struct {
	ID             string                     `json:"id"`
	ClientID       string                     `json:"clientId"`
	Period         period.Period              `json:"period"`
	Status         Status                     `json:"status"`
	Version        int                        `json:"version"`
	ResultHash     string                     `json:"resultHash"`
	Severity       signal.Severity            `json:"severity"`
	Signals        []signal.Signal            `json:"signals"`
	Description    string                     `json:"description"`
	Mono           *tax.MonotributoAssessment `json:"mono,omitempty"`
	RI             *tax.RIAssessment          `json:"ri,omitempty"`
	Reconciliation Reconciliation             `json:"reconciliation"`
	ComputedAt     time.Time                  `json:"computedAt"`
	Approvals      []Approval                 `json:"approvals"`
	PaidAt         *time.Time                 `json:"paidAt,omitempty"`
	Published      *Published                 `json:"published,omitempty"`
}

// NewClosing inicializa un nuevo cierre en versión 1 con su estado inicial según confirmación de input.
func NewClosing(
	id, clientID string,
	p period.Period,
	confirmedInput bool,
	hash string,
	severity signal.Severity,
	signals []signal.Signal,
	description string,
	mono *tax.MonotributoAssessment,
	ri *tax.RIAssessment,
	rec Reconciliation,
	now time.Time,
) Closing {
	status := StatusPendingClientInput
	if confirmedInput {
		status = StatusReadyForReview
	}

	return Closing{
		ID:             id,
		ClientID:       clientID,
		Period:         p,
		Status:         status,
		Version:        1,
		ResultHash:     hash,
		Severity:       severity,
		Signals:        signals,
		Description:    description,
		Mono:           mono,
		RI:             ri,
		Reconciliation: rec,
		ComputedAt:     now,
		Approvals:      []Approval{},
		PaidAt:         nil,
	}
}

// ApplyResult aplica una nueva liquidación. Si el hash no cambió, devuelve false y no genera versión.
// Si el hash cambió, incrementa Version, actualiza datos, y si estaba APPROVED o NOTIFIED,
// reabre a READY_FOR_REVIEW preservando el historial de aprobaciones.
func (c *Closing) ApplyResult(
	hash string,
	severity signal.Severity,
	signals []signal.Signal,
	description string,
	mono *tax.MonotributoAssessment,
	ri *tax.RIAssessment,
	rec Reconciliation,
	now time.Time,
) bool {
	if c.ResultHash == hash {
		c.ComputedAt = now
		return false
	}

	c.ResultHash = hash
	c.Version++
	c.Severity = severity
	c.Signals = signals
	c.Description = description
	c.Mono = mono
	c.RI = ri
	c.Reconciliation = rec
	c.ComputedAt = now

	if c.Status == StatusApproved || c.Status == StatusNotified {
		c.Status = StatusReadyForReview
	}

	return true
}

// MarkReady avanza el estado de PENDING_CLIENT_INPUT a READY_FOR_REVIEW cuando el cliente confirma su carga.
func (c *Closing) MarkReady() error {
	if c.Status == StatusPendingClientInput {
		c.Status = StatusReadyForReview
		return nil
	}
	if c.Status == StatusReadyForReview {
		return nil
	}
	return fmt.Errorf("%w: no se puede marcar listo para revisión desde estado %s", ErrInvalidTransition, c.Status)
}

// Approve registra la aprobación del contador. Es válida desde READY_FOR_REVIEW o PENDING_CLIENT_INPUT.
func (c *Closing) Approve(a Approval) error {
	if c.Status != StatusReadyForReview && c.Status != StatusPendingClientInput {
		return fmt.Errorf("%w: no se puede aprobar desde estado %s", ErrInvalidTransition, c.Status)
	}

	a.Version = c.Version
	a.ResultHash = c.ResultHash
	c.Status = StatusApproved
	c.Approvals = append(c.Approvals, a)
	return nil
}

// MarkNotified avanza el estado de APPROVED a NOTIFIED una vez despachado el aviso al cliente.
func (c *Closing) MarkNotified(p ...Published) error {
	if c.Status != StatusApproved {
		return fmt.Errorf("%w: no se puede marcar notificado desde estado %s", ErrInvalidTransition, c.Status)
	}
	c.Status = StatusNotified
	if len(p) > 0 {
		c.Published = &p[0]
	}
	return nil
}

// Publish congela el snapshot Published y avanza el estado a NOTIFIED (011 §3).
func (c *Closing) Publish(pub Published) error {
	return c.MarkNotified(pub)
}

// MarkPaid registra la fecha de pago del cliente. Es independiente del estado del cierre.
func (c *Closing) MarkPaid(now time.Time) error {
	c.PaidAt = &now
	return nil
}
