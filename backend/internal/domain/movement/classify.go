package movement

import (
	"time"
)

// ClassificationKind define el resultado de la clasificación de un movimiento.
type ClassificationKind string

const (
	KindIncome       ClassificationKind = "INCOME"
	KindNotIncome    ClassificationKind = "NOT_INCOME"
	KindUnclassified ClassificationKind = "UNCLASSIFIED"
)

// Confidence define el nivel de confianza de la clasificación.
type Confidence string

const (
	ConfidenceHigh Confidence = "HIGH"
	ConfidenceLow  Confidence = "LOW"
)

// RuleCode representa el código de regla aplicado para clasificar.
type RuleCode string

const (
	RuleOwnAccountTransfer   RuleCode = "OWN_ACCOUNT_TRANSFER"
	RuleStablecoinConversion RuleCode = "STABLECOIN_CONVERSION"
	RuleCardExpense          RuleCode = "CARD_EXPENSE"
	RuleOutbound             RuleCode = "OUTBOUND"
	RuleKnownPayer           RuleCode = "KNOWN_PAYER"
	RuleForeignCompany       RuleCode = "FOREIGN_COMPANY"
	RuleForeignPerson        RuleCode = "FOREIGN_PERSON"
	RuleLocalInbound         RuleCode = "LOCAL_INBOUND"
	RuleFallback             RuleCode = "FALLBACK"
	RuleManualOverride       RuleCode = "MANUAL_OVERRIDE"
)

// Classification representa el resultado determinístico o manual de clasificar un movimiento.
type Classification struct {
	MovementID   string             `json:"movementId"`
	Kind         ClassificationKind `json:"kind"`
	Confidence   Confidence         `json:"confidence"`
	Rule         RuleCode           `json:"rule"`
	OverriddenBy string             `json:"overriddenBy,omitempty"`
	OverriddenAt *time.Time         `json:"overriddenAt,omitempty"`
	Note         string             `json:"note,omitempty"`
}

// Classify aplica las reglas determinísticas R1..R9 en estricto orden (domain-context.md §4).
// knownPayers contiene nombres normalizados de receptores de facturas emitidas por el cliente antes del período.
func Classify(m Movement, knownPayers map[string]bool) Classification {
	res := Classification{
		MovementID: m.ID,
	}

	// R1: OWN_ACCOUNT_TRANSFER: counterparty.isOwnAccount == true -> NOT_INCOME / HIGH
	if m.Counterparty.IsOwnAccount {
		res.Kind = KindNotIncome
		res.Confidence = ConfidenceHigh
		res.Rule = RuleOwnAccountTransfer
		return res
	}

	// R2: STABLECOIN_CONVERSION: tipo STABLECOIN_BUY o STABLECOIN_SELL -> NOT_INCOME / HIGH
	if m.Type == TypeStablecoinBuy || m.Type == TypeStablecoinSell {
		res.Kind = KindNotIncome
		res.Confidence = ConfidenceHigh
		res.Rule = RuleStablecoinConversion
		return res
	}

	// R3: CARD_EXPENSE: tipo CARD_SPEND o CARD_REFUND -> NOT_INCOME / HIGH
	if m.Type == TypeCardSpend || m.Type == TypeCardRefund {
		res.Kind = KindNotIncome
		res.Confidence = ConfidenceHigh
		res.Rule = RuleCardExpense
		return res
	}

	// R4: OUTBOUND: tipo OUTBOUND_TRANSFER o FEE -> NOT_INCOME / HIGH
	if m.Type == TypeOutboundTransfer || m.Type == TypeFee {
		res.Kind = KindNotIncome
		res.Confidence = ConfidenceHigh
		res.Rule = RuleOutbound
		return res
	}

	// R5: KNOWN_PAYER: INBOUND_TRANSFER y el nombre normalizado coincide con receptor histórico
	if m.Type == TypeInboundTransfer && len(knownPayers) > 0 {
		normName := NormalizeName(m.Counterparty.Name)
		if knownPayers[normName] {
			res.Kind = KindIncome
			res.Confidence = ConfidenceHigh
			res.Rule = RuleKnownPayer
			return res
		}
	}

	// R6: FOREIGN_COMPANY: INBOUND_TRANSFER, country != "AR", kind in {COMPANY, PLATFORM} -> INCOME / HIGH
	if m.Type == TypeInboundTransfer && m.Counterparty.Country != "AR" &&
		(m.Counterparty.Kind == CounterpartyCompany || m.Counterparty.Kind == CounterpartyPlatform) {
		res.Kind = KindIncome
		res.Confidence = ConfidenceHigh
		res.Rule = RuleForeignCompany
		return res
	}

	// R7: FOREIGN_PERSON: INBOUND_TRANSFER, country != "AR", kind == PERSON -> INCOME / LOW
	if m.Type == TypeInboundTransfer && m.Counterparty.Country != "AR" && m.Counterparty.Kind == CounterpartyPerson {
		res.Kind = KindIncome
		res.Confidence = ConfidenceLow
		res.Rule = RuleForeignPerson
		return res
	}

	// R8: LOCAL_INBOUND: INBOUND_TRANSFER, country == "AR" -> UNCLASSIFIED / LOW
	if m.Type == TypeInboundTransfer && m.Counterparty.Country == "AR" {
		res.Kind = KindUnclassified
		res.Confidence = ConfidenceLow
		res.Rule = RuleLocalInbound
		return res
	}

	// R9: FALLBACK: cualquier otro -> UNCLASSIFIED / LOW
	res.Kind = KindUnclassified
	res.Confidence = ConfidenceLow
	res.Rule = RuleFallback
	return res
}
