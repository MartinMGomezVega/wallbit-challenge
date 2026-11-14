// Package signal implementa las señales de riesgo, el semáforo y los templates de cierre
// según domain-context.md §8 y architecture.md §2.2.
package signal

import (
	"fmt"
	"sort"
)

// Code identifica unívocamente una señal de riesgo del catálogo oficial.
type Code string

const (
	CodeMonoExclusionKCapExceeded     Code = "MONO_EXCLUSION_K_CAP_EXCEEDED"
	CodePostExclusionCInvoices        Code = "POST_EXCLUSION_C_INVOICES"
	CodeMissingInvoices               Code = "MISSING_INVOICES"
	CodeDepositsExceedKCap            Code = "DEPOSITS_EXCEED_K_CAP"
	CodeMonoNearKCap                  Code = "MONO_NEAR_K_CAP"
	CodeMonoCategoryChangeRequired    Code = "MONO_CATEGORY_CHANGE_REQUIRED"
	CodeMonoNearCategoryCap           Code = "MONO_NEAR_CATEGORY_CAP"
	CodeAtypicalIncome                Code = "ATYPICAL_INCOME"
	CodeUnclassifiedMovements         Code = "UNCLASSIFIED_MOVEMENTS"
	CodeExternalIncomeDeclared        Code = "EXTERNAL_INCOME_DECLARED"
	CodeInvoiceWithoutPayment         Code = "INVOICE_WITHOUT_PAYMENT"
	CodeHighCardSpending              Code = "HIGH_CARD_SPENDING"
	CodeMonoCategoryDowngradePossible Code = "MONO_CATEGORY_DOWNGRADE_POSSIBLE"
)

// Severity representa la gravedad de una señal o del cierre en su conjunto.
type Severity int

const (
	SeverityGreen  Severity = 0
	SeverityYellow Severity = 1
	SeverityRed    Severity = 2
)

func (s Severity) String() string {
	switch s {
	case SeverityGreen:
		return "GREEN"
	case SeverityYellow:
		return "YELLOW"
	case SeverityRed:
		return "RED"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON serializa Severity como string ("GREEN", "YELLOW", "RED").
func (s Severity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// UnmarshalJSON deserializa Severity desde string ("GREEN", "YELLOW", "RED").
func (s *Severity) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case `"GREEN"`:
		*s = SeverityGreen
	case `"YELLOW"`:
		*s = SeverityYellow
	case `"RED"`:
		*s = SeverityRed
	default:
		return fmt.Errorf("unknown severity: %s", str)
	}
	return nil
}

// Meta define los metadatos fijos de una señal en el catálogo.
type Meta struct {
	Code     Code
	Severity Severity
	Priority int
	Label    string
}

// CatalogItem representa un elemento exportable del catálogo de señales para la API.
type CatalogItem struct {
	Code     Code   `json:"code"`
	Severity string `json:"severity"`
	Priority int    `json:"priority"`
	Label    string `json:"label"`
}

var catalog = map[Code]Meta{
	CodeMonoExclusionKCapExceeded: {
		Code:     CodeMonoExclusionKCapExceeded,
		Severity: SeverityRed,
		Priority: 1,
		Label:    "Excede tope K (pasa a RI)",
	},
	CodePostExclusionCInvoices: {
		Code:     CodePostExclusionCInvoices,
		Severity: SeverityRed,
		Priority: 2,
		Label:    "Facturas C post-exclusión",
	},
	CodeMissingInvoices: {
		Code:     CodeMissingInvoices,
		Severity: SeverityRed,
		Priority: 3,
		Label:    "Faltan facturas",
	},
	CodeDepositsExceedKCap: {
		Code:     CodeDepositsExceedKCap,
		Severity: SeverityRed,
		Priority: 4,
		Label:    "Depósitos sobre tope K",
	},
	CodeMonoNearKCap: {
		Code:     CodeMonoNearKCap,
		Severity: SeverityYellow,
		Priority: 5,
		Label:    "Cerca del tope K",
	},
	CodeMonoCategoryChangeRequired: {
		Code:     CodeMonoCategoryChangeRequired,
		Severity: SeverityYellow,
		Priority: 6,
		Label:    "Debe recategorizar",
	},
	CodeMonoNearCategoryCap: {
		Code:     CodeMonoNearCategoryCap,
		Severity: SeverityYellow,
		Priority: 7,
		Label:    "Cerca del tope de su categoría",
	},
	CodeAtypicalIncome: {
		Code:     CodeAtypicalIncome,
		Severity: SeverityYellow,
		Priority: 8,
		Label:    "Cobro fuera de lo habitual",
	},
	CodeUnclassifiedMovements: {
		Code:     CodeUnclassifiedMovements,
		Severity: SeverityYellow,
		Priority: 9,
		Label:    "Movimientos sin clasificar",
	},
	CodeExternalIncomeDeclared: {
		Code:     CodeExternalIncomeDeclared,
		Severity: SeverityYellow,
		Priority: 10,
		Label:    "Cobros fuera de Wallbit",
	},
	CodeInvoiceWithoutPayment: {
		Code:     CodeInvoiceWithoutPayment,
		Severity: SeverityYellow,
		Priority: 11,
		Label:    "Facturas sin cobro",
	},
	CodeHighCardSpending: {
		Code:     CodeHighCardSpending,
		Severity: SeverityYellow,
		Priority: 12,
		Label:    "Gastos con tarjeta altos",
	},
	CodeMonoCategoryDowngradePossible: {
		Code:     CodeMonoCategoryDowngradePossible,
		Severity: SeverityYellow,
		Priority: 13,
		Label:    "Puede bajar de categoría",
	},
}

// Catalog devuelve todos los elementos del catálogo ordenados por prioridad ascendente.
func Catalog() []CatalogItem {
	items := make([]CatalogItem, 0, len(catalog))
	for _, m := range catalog {
		items = append(items, CatalogItem{
			Code:     m.Code,
			Severity: m.Severity.String(),
			Priority: m.Priority,
			Label:    m.Label,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority < items[j].Priority
	})
	return items
}

// MetaFor devuelve los metadatos de un código de señal.
func MetaFor(code Code) (Meta, bool) {
	m, ok := catalog[code]
	return m, ok
}
