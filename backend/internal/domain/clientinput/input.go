package clientinput

import (
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ConfirmedBy define quién confirmó el período (spec 016).
type ConfirmedBy string

const (
	ConfirmedByClient     ConfirmedBy = "CLIENT"
	ConfirmedByAccountant ConfirmedBy = "ACCOUNTANT"
	ConfirmedBySeed       ConfirmedBy = "SEED"
)

// Errores de inputs del cliente.
var (
	ErrExternalIncomeNotAnswered = domain.NewError(domain.ErrValidation, "EXTERNAL_INCOME_NOT_ANSWERED", "Se requiere responder sobre ingresos externos antes de confirmar el período")
	ErrAlreadyConfirmed          = domain.NewError(domain.ErrConflict, "ALREADY_CONFIRMED", "El período ya fue confirmado")
	ErrNoteRequired              = domain.NewError(domain.ErrValidation, "NOTE_REQUIRED", "La nota es obligatoria cuando el contador confirma el período (1 a 300 caracteres)")
)

// ExternalIncomeItem representa un cobro percibido fuera de Wallbit declarado por el cliente.
type ExternalIncomeItem struct {
	Amount      money.Money `json:"amount"` // USD
	Description string      `json:"description"`
}

// ClientPeriodInput agrupa las declaraciones del cliente para un período mensual (spec 003 criterio 3).
type ClientPeriodInput struct {
	ClientID                 string               `json:"clientId"`
	Period                   period.Period        `json:"period"`
	ExternalIncome           []ExternalIncomeItem `json:"externalIncome"`
	DeclaredNoExternalIncome bool                 `json:"declaredNoExternalIncome"`
	ConfirmedAt              *time.Time           `json:"confirmedAt,omitempty"`
	ConfirmedBy              ConfirmedBy          `json:"confirmedBy,omitempty"`
	ConfirmationNote         string               `json:"confirmationNote,omitempty"`
	UpdatedAt                time.Time            `json:"updatedAt"`
}

// Answered indica si el cliente ya respondió sobre ingresos externos (cargando ítems o declarando que no tuvo).
func (c ClientPeriodInput) Answered() bool {
	return c.DeclaredNoExternalIncome || len(c.ExternalIncome) > 0
}

// TotalExternalIncomeUSD suma los ingresos externos declarados en USD.
func (c ClientPeriodInput) TotalExternalIncomeUSD() money.Money {
	total := money.Zero(money.USD)
	for _, item := range c.ExternalIncome {
		if item.Amount.Currency == money.USD {
			total, _ = total.Add(item.Amount)
		}
	}
	return total
}

// IsConfirmed reporta si el cliente confirmó el período.
func (c ClientPeriodInput) IsConfirmed() bool {
	return c.ConfirmedAt != nil
}

// ConfirmBy sella la confirmación con el actor indicado (CLIENT o ACCOUNTANT) y registra la nota si aplica.
func (c *ClientPeriodInput) ConfirmBy(by ConfirmedBy, note string, now time.Time) error {
	if c.IsConfirmed() {
		return ErrAlreadyConfirmed
	}
	if !c.Answered() {
		return ErrExternalIncomeNotAnswered
	}
	if by == ConfirmedByAccountant {
		trimmed := strings.TrimSpace(note)
		if len(trimmed) < 1 || len(trimmed) > 300 {
			return ErrNoteRequired
		}
		c.ConfirmationNote = trimmed
	}
	c.ConfirmedBy = by
	c.ConfirmedAt = &now
	c.UpdatedAt = now
	return nil
}

// Confirm sella la confirmación del cliente si ya respondió sobre ingresos externos (criterio 14).
func (c *ClientPeriodInput) Confirm(now time.Time) error {
	return c.ConfirmBy(ConfirmedByClient, "", now)
}
