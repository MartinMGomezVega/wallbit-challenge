package closing

import (
	"sort"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// SignalSummary es una vista resumida de una señal para las listas y tablas de la UI.
type SignalSummary struct {
	Code     signal.Code `json:"code"`
	Label    string      `json:"label"`
	Severity string      `json:"severity,omitempty"`
}

// BilledInPeriod agrupa los montos facturados en el período.
type BilledInPeriod struct {
	USD money.Money `json:"usd"`
	ARS money.Money `json:"ars"`
}

// Accumulated12m agrupa el acumulado de 12 meses móviles y su porcentaje respecto al tope K.
type Accumulated12m struct {
	ARS         money.Money `json:"ars"`
	PctOfKCapBp int64       `json:"pctOfKCapBp"`
}

// ToPay resume el importe total a pagar del período y su fecha de vencimiento más próxima.
type ToPay struct {
	Amount  money.Money `json:"amount"`
	DueDate time.Time   `json:"dueDate"`
}

// Row representa una fila plana en la tabla del contador con datos del cierre y cliente.
type Row struct {
	ClosingID      string          `json:"closingId"`
	ClientID       string          `json:"clientId"`
	ClientName     string          `json:"clientName"`
	CUIT           string          `json:"cuit"`
	Regime         client.Regime   `json:"regime"`
	Category       *tax.Category   `json:"category,omitempty"`
	Severity       signal.Severity `json:"severity"`
	PrimarySignal  *SignalSummary  `json:"primarySignal,omitempty"`
	Description    string          `json:"description"`
	OtherSignals   []SignalSummary `json:"otherSignals"`
	BilledInPeriod BilledInPeriod  `json:"billedInPeriod"`
	Accumulated12m Accumulated12m  `json:"accumulated12m"`
	ToPay          ToPay           `json:"toPay"`
	Status         Status          `json:"status"`
	Version        int             `json:"version"`
	PaidAt         *time.Time      `json:"paidAt,omitempty"`
}

// Filter especifica los criterios de filtrado para las filas de cierres.
type Filter struct {
	Severities []signal.Severity
	Signals    []signal.Code
	Regime     client.Regime
	Statuses   []Status
	Query      string
}

// Sort especifica la columna y dirección de ordenamiento.
type Sort struct {
	By  string // "severity", "name", "regime", "billed", "toPay", "pctKCap", "status"
	Dir string // "asc", "desc"
}

// Apply filtra y ordena las filas dadas de manera pura (sin mutar el slice de entrada).
func Apply(rows []Row, f Filter, s Sort) []Row {
	var filtered []Row

	cleanQ := strings.ToLower(strings.TrimSpace(f.Query))
	cleanCUIT := strings.ReplaceAll(strings.ReplaceAll(cleanQ, "-", ""), " ", "")

	for _, r := range rows {
		// 1. Filtro por severidad
		if len(f.Severities) > 0 {
			match := false
			for _, sev := range f.Severities {
				if r.Severity == sev {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// 2. Filtro por señales (match si tiene al menos una de las indicadas)
		if len(f.Signals) > 0 {
			match := false
			if r.PrimarySignal != nil {
				for _, sc := range f.Signals {
					if r.PrimarySignal.Code == sc {
						match = true
						break
					}
				}
			}
			if !match {
				for _, other := range r.OtherSignals {
					for _, sc := range f.Signals {
						if other.Code == sc {
							match = true
							break
						}
					}
					if match {
						break
					}
				}
			}
			if !match {
				continue
			}
		}

		// 3. Filtro por régimen
		if f.Regime != "" && r.Regime != f.Regime {
			continue
		}

		// 4. Filtro por estado
		if len(f.Statuses) > 0 {
			match := false
			for _, st := range f.Statuses {
				if r.Status == st {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// 5. Filtro por búsqueda (q: nombre o CUIT)
		if cleanQ != "" {
			nameLower := strings.ToLower(r.ClientName)
			cuitNorm := strings.ReplaceAll(strings.ReplaceAll(r.CUIT, "-", ""), " ", "")
			if !strings.Contains(nameLower, cleanQ) && !strings.Contains(cuitNorm, cleanCUIT) {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	out := make([]Row, len(filtered))
	copy(out, filtered)

	sortRows(out, s)
	return out
}

func statusRank(st Status) int {
	switch st {
	case StatusPendingClientInput:
		return 1
	case StatusReadyForReview:
		return 2
	case StatusApproved:
		return 3
	case StatusNotified:
		return 4
	default:
		return 5
	}
}

func sortRows(rows []Row, s Sort) {
	by := s.By
	if by == "" {
		by = "severity"
	}
	dir := strings.ToLower(s.Dir)
	if dir == "" {
		switch by {
		case "name", "regime", "status":
			dir = "asc"
		default:
			dir = "desc"
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		switch by {
		case "name":
			cmp := strings.Compare(strings.ToLower(rows[i].ClientName), strings.ToLower(rows[j].ClientName))
			if cmp != 0 {
				if dir == "desc" {
					return cmp > 0
				}
				return cmp < 0
			}

		case "regime":
			regimeI := string(rows[i].Regime)
			regimeJ := string(rows[j].Regime)
			if regimeI != regimeJ {
				if dir == "desc" {
					return regimeI > regimeJ
				}
				return regimeI < regimeJ
			}
			catI := ""
			if rows[i].Category != nil {
				catI = string(*rows[i].Category)
			}
			catJ := ""
			if rows[j].Category != nil {
				catJ = string(*rows[j].Category)
			}
			if catI != catJ {
				if dir == "desc" {
					return catI > catJ
				}
				return catI < catJ
			}

		case "status":
			rankI := statusRank(rows[i].Status)
			rankJ := statusRank(rows[j].Status)
			diff := rankI - rankJ
			if diff != 0 {
				if dir == "desc" {
					return diff > 0
				}
				return diff < 0
			}

		case "billed":
			diff := rows[i].BilledInPeriod.ARS.Amount - rows[j].BilledInPeriod.ARS.Amount
			if diff != 0 {
				if dir == "asc" {
					return diff < 0
				}
				return diff > 0
			}

		case "toPay":
			diff := rows[i].ToPay.Amount.Amount - rows[j].ToPay.Amount.Amount
			if diff != 0 {
				if dir == "asc" {
					return diff < 0
				}
				return diff > 0
			}

		case "pctKCap":
			diff := rows[i].Accumulated12m.PctOfKCapBp - rows[j].Accumulated12m.PctOfKCapBp
			if diff != 0 {
				if dir == "asc" {
					return diff < 0
				}
				return diff > 0
			}

		case "severity":
			fallthrough
		default:
			// Orden severity = severidad desc (RED > YELLOW > GREEN),
			// prioridad de señal principal asc (1..13), nombre asc, clientId asc.
			sevDiff := int(rows[i].Severity) - int(rows[j].Severity)
			if sevDiff != 0 {
				if dir == "asc" {
					return sevDiff < 0
				}
				return sevDiff > 0
			}

			// Misma severidad: prioridad de señal principal asc
			prioI := 999
			if rows[i].PrimarySignal != nil {
				if meta, ok := signal.MetaFor(rows[i].PrimarySignal.Code); ok {
					prioI = meta.Priority
				}
			}
			prioJ := 999
			if rows[j].PrimarySignal != nil {
				if meta, ok := signal.MetaFor(rows[j].PrimarySignal.Code); ok {
					prioJ = meta.Priority
				}
			}
			if prioI != prioJ {
				if dir == "asc" {
					return prioI > prioJ
				}
				return prioI < prioJ
			}

			// Mismo prioridad: nombre asc
			cmp := strings.Compare(strings.ToLower(rows[i].ClientName), strings.ToLower(rows[j].ClientName))
			if cmp != 0 {
				if dir == "asc" {
					return cmp > 0
				}
				return cmp < 0
			}
		}

		// Desempate estable por ClientID asc
		return rows[i].ClientID < rows[j].ClientID
	})
}

// Summary contiene los contadores para los chips de la UI.
type Summary struct {
	Total      int            `json:"total"`
	BySeverity map[string]int `json:"bySeverity"`
	BySignal   map[string]int `json:"bySignal"`
	ByStatus   map[Status]int `json:"byStatus"`
}

// ComputeSummary calcula los contadores agrupados para una lista de filas.
func ComputeSummary(rows []Row) Summary {
	s := Summary{
		Total: len(rows),
		BySeverity: map[string]int{
			"RED":    0,
			"YELLOW": 0,
			"GREEN":  0,
		},
		BySignal: make(map[string]int),
		ByStatus: map[Status]int{
			StatusPendingClientInput: 0,
			StatusReadyForReview:     0,
			StatusApproved:           0,
			StatusNotified:           0,
		},
	}

	for _, r := range rows {
		s.BySeverity[r.Severity.String()]++
		s.ByStatus[r.Status]++

		seen := make(map[signal.Code]bool)
		if r.PrimarySignal != nil {
			seen[r.PrimarySignal.Code] = true
		}
		for _, sig := range r.OtherSignals {
			seen[sig.Code] = true
		}
		for code := range seen {
			s.BySignal[string(code)]++
		}
	}
	return s
}

// BuildRow construye una fila plana de tablero a partir de un cierre y su cliente.
func BuildRow(c Closing, cli client.Client) Row {
	var cat *tax.Category
	if c.Mono != nil {
		catCopy := c.Mono.CurrentCategory
		cat = &catCopy
	} else if cli.Regime == client.RegimeMonotributo && cli.MonotributoCategory != "" {
		catCopy := cli.MonotributoCategory
		cat = &catCopy
	}

	var primary *SignalSummary
	var others []SignalSummary
	if len(c.Signals) > 0 {
		first := c.Signals[0]
		label := string(first.Code)
		if meta, ok := signal.MetaFor(first.Code); ok {
			label = meta.Label
		}
		primary = &SignalSummary{
			Code:     first.Code,
			Label:    label,
			Severity: first.Severity.String(),
		}
		for _, s := range c.Signals[1:] {
			sLabel := string(s.Code)
			if meta, ok := signal.MetaFor(s.Code); ok {
				sLabel = meta.Label
			}
			others = append(others, SignalSummary{
				Code:     s.Code,
				Label:    sLabel,
				Severity: s.Severity.String(),
			})
		}
	}
	if others == nil {
		others = []SignalSummary{}
	}

	var billed BilledInPeriod
	var accum Accumulated12m
	var toPay ToPay

	if c.Mono != nil {
		billed = BilledInPeriod{
			USD: c.Mono.BilledInPeriodUSD,
			ARS: c.Mono.BilledInPeriodARS,
		}
		accum = Accumulated12m{
			ARS:         c.Mono.Billed12mARS,
			PctOfKCapBp: c.Mono.PctOfKCapBp,
		}
		toPay = ToPay{
			Amount:  c.Mono.Fee,
			DueDate: c.Mono.FeeDueDate,
		}
	} else if c.RI != nil {
		billed = BilledInPeriod{
			USD: c.RI.BilledInPeriodUSD,
			ARS: c.RI.BilledInPeriodARS,
		}
		accum = Accumulated12m{
			ARS:         money.Zero(money.ARS),
			PctOfKCapBp: 0,
		}
		minDue := c.RI.VatDueDate
		if !c.RI.SelfEmployedDueDate.IsZero() && (minDue.IsZero() || c.RI.SelfEmployedDueDate.Before(minDue)) {
			minDue = c.RI.SelfEmployedDueDate
		}
		toPay = ToPay{
			Amount:  c.RI.TotalToPay,
			DueDate: minDue,
		}
	}

	return Row{
		ClosingID:      c.ID,
		ClientID:       c.ClientID,
		ClientName:     cli.Name,
		CUIT:           cli.CUIT.String(),
		Regime:         cli.RegimeAt(c.Period.End()),
		Category:       cat,
		Severity:       c.Severity,
		PrimarySignal:  primary,
		Description:    c.Description,
		OtherSignals:   others,
		BilledInPeriod: billed,
		Accumulated12m: accum,
		ToPay:          toPay,
		Status:         c.Status,
		Version:        c.Version,
		PaidAt:         c.PaidAt,
	}
}
