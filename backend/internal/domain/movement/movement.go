// Package movement modela las operaciones financieras dentro de Wallbit, su clasificación
// con reglas determinísticas y la normalización de contrapartes (domain-context.md §4, §6).
package movement

import (
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

// Tipos de movimiento.
const (
	TypeInboundTransfer  MovementType = "INBOUND_TRANSFER"
	TypeOutboundTransfer MovementType = "OUTBOUND_TRANSFER"
	TypeCardSpend        MovementType = "CARD_SPEND"
	TypeCardRefund       MovementType = "CARD_REFUND"
	TypeStablecoinBuy    MovementType = "STABLECOIN_BUY"
	TypeStablecoinSell   MovementType = "STABLECOIN_SELL"
	TypeFee              MovementType = "FEE"
)

// MovementType describe el tipo de operación en Wallbit.
type MovementType string

// Tipos de contraparte.
const (
	CounterpartyCompany  CounterpartyKind = "COMPANY"
	CounterpartyPerson   CounterpartyKind = "PERSON"
	CounterpartyPlatform CounterpartyKind = "PLATFORM"
)

// CounterpartyKind indica si la contraparte es empresa, persona física o plataforma intermediaria.
type CounterpartyKind string

// Counterparty describe quién originó o recibió los fondos.
type Counterparty struct {
	Name         string           `json:"name"`
	Country      string           `json:"country"` // ISO-2 (ej. "US", "DE", "AR")
	Kind         CounterpartyKind `json:"kind"`
	IsOwnAccount bool             `json:"isOwnAccount"`
}

// Movement representa un movimiento dentro de Wallbit.
type Movement struct {
	ID           string       `json:"id"`
	ClientID     string       `json:"clientId"`
	Type         MovementType `json:"type"`
	Amount       money.Money  `json:"amount"` // Monto en USD (centavos)
	FXRate       money.Rate   `json:"fxRate"` // TC de la operación registrado por Wallbit
	OccurredAt   time.Time    `json:"occurredAt"`
	Counterparty Counterparty `json:"counterparty"`
	Description  string       `json:"description"`
}

// AmountARS devuelve el importe pesificado usando el tipo de cambio propio de la operación
// (domain-context.md §4 nota: "Pesificación de movimientos: usar el tipo de cambio de la operación").
func (m Movement) AmountARS() (money.Money, error) {
	if m.Amount.Currency != money.USD {
		return money.Money{}, fmt.Errorf("el movimiento debe estar en USD, got %s", m.Amount.Currency)
	}
	return money.ConvertUSDToARS(m.Amount, m.FXRate)
}
