package closing

import (
	"sort"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ProgressStats resume el avance operativo del cierre mensual.
type ProgressStats struct {
	Total                  int               `json:"total"`
	ByStatus               map[string]int    `json:"byStatus"`
	ApprovedPct            money.BasisPoints `json:"approvedPct"`
	PendingReview          map[string]int    `json:"pendingReview"`
	EstimatedReviewMinutes int               `json:"estimatedReviewMinutes"`
}

// DueDateTotal agrupa montos a pagar por fecha de vencimiento.
type DueDateTotal struct {
	Date    string      `json:"date"`
	Amount  money.Money `json:"amount"`
	Clients int         `json:"clients"`
}

// TotalsStats consolida los montos facturados y a pagar del período.
type TotalsStats struct {
	BilledUSD      money.Money    `json:"billedUSD"`
	BilledARS      money.Money    `json:"billedARS"`
	ToPayARS       money.Money    `json:"toPayARS"`
	ToPayByDueDate []DueDateTotal `json:"toPayByDueDate"`
}

// SignalStat cuenta cuántos cierres tienen una señal dada.
type SignalStat struct {
	Code     string `json:"code"`
	Label    string `json:"label"`
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}

// CategoryStat cuenta cuántos clientes pertenecen a cada categoría o régimen.
type CategoryStat struct {
	Category string `json:"category"`
	Clients  int    `json:"clients"`
}

// RiskStats agrupa las métricas de riesgo del período.
type RiskStats struct {
	BySeverity map[string]int `json:"bySeverity"`
	BySignal   []SignalStat   `json:"bySignal"`
	ByCategory []CategoryStat `json:"byCategory"`
	NearKCap   int            `json:"nearKCap"`
}

// SuccessStats mide los KPIs de éxito del servicio contable.
type SuccessStats struct {
	GreenPct                 money.BasisPoints `json:"greenPct"`
	GreenCount               int               `json:"greenCount"`
	Total                    int               `json:"total"`
	NotifiedBeforeDay10Pct   money.BasisPoints `json:"notifiedBeforeDay10Pct"`
	NotifiedBeforeDay10Count int               `json:"notifiedBeforeDay10Count"`
	NotifiedCount            int               `json:"notifiedCount"`
	PaidOnTimePct            money.BasisPoints `json:"paidOnTimePct"`
	PaidOnTimeCount          int               `json:"paidOnTimeCount"`
}

// PeriodStats es la respuesta completa del dashboard de resumen (spec 014).
type PeriodStats struct {
	Period   string        `json:"period"`
	Progress ProgressStats `json:"progress"`
	Totals   TotalsStats   `json:"totals"`
	Risk     RiskStats     `json:"risk"`
	Success  SuccessStats  `json:"success"`
}

var allMonoCategories = []tax.Category{
	tax.CatA, tax.CatB, tax.CatC, tax.CatD, tax.CatE, tax.CatF,
	tax.CatG, tax.CatH, tax.CatI, tax.CatJ, tax.CatK,
}

// ComputeStats calcula de forma pura las estadísticas agregadas de un período a partir de sus cierres.
func ComputeStats(closings []Closing, p period.Period) PeriodStats {
	loc := period.Location()
	byStatus := map[string]int{
		string(StatusPendingClientInput): 0,
		string(StatusReadyForReview):     0,
		string(StatusApproved):           0,
		string(StatusNotified):           0,
	}
	pendingReview := map[string]int{
		"RED":    0,
		"YELLOW": 0,
		"GREEN":  0,
	}
	bySeverity := map[string]int{
		"RED":    0,
		"YELLOW": 0,
		"GREEN":  0,
	}
	catCounts := make(map[string]int)
	for _, cat := range allMonoCategories {
		catCounts[string(cat)] = 0
	}
	catCounts["RI"] = 0

	type dueAccum struct {
		amount  int64
		clients int
	}
	dueMap := make(map[string]*dueAccum)

	type sigAccum struct {
		code     string
		severity signal.Severity
		count    int
	}
	sigMap := make(map[string]*sigAccum)

	var totalBilledUSD, totalBilledARS, totalToPayARS int64
	nearKCap := 0
	notifiedCount := 0
	notifiedBeforeDay10 := 0
	paidOnTime := 0

	nextP := p.Next()
	day10Cutoff := time.Date(nextP.Year, time.Month(nextP.Month), 10, 23, 59, 59, 999999999, loc)

	for _, c := range closings {
		byStatus[string(c.Status)]++
		sevStr := c.Severity.String()
		bySeverity[sevStr]++

		if c.Status == StatusPendingClientInput || c.Status == StatusReadyForReview {
			pendingReview[sevStr]++
		}

		var nearestDue time.Time
		if c.Mono != nil {
			totalBilledUSD += c.Mono.BilledInPeriodUSD.Amount
			totalBilledARS += c.Mono.BilledInPeriodARS.Amount
			totalToPayARS += c.Mono.Fee.Amount
			if c.Mono.CurrentCategory != "" {
				catCounts[string(c.Mono.CurrentCategory)]++
			}
			if !c.Mono.FeeDueDate.IsZero() {
				nearestDue = c.Mono.FeeDueDate.In(loc)
				dKey := nearestDue.Format("2006-01-02")
				da, ok := dueMap[dKey]
				if !ok {
					da = &dueAccum{}
					dueMap[dKey] = da
				}
				da.amount += c.Mono.Fee.Amount
				da.clients++
			}
		} else if c.RI != nil {
			totalBilledUSD += c.RI.BilledInPeriodUSD.Amount
			totalBilledARS += c.RI.BilledInPeriodARS.Amount
			totalToPayARS += c.RI.TotalToPay.Amount
			catCounts["RI"]++
			if !c.RI.VatDueDate.IsZero() {
				nearestDue = c.RI.VatDueDate.In(loc)
				if !c.RI.SelfEmployedDueDate.IsZero() && c.RI.SelfEmployedDueDate.In(loc).Before(nearestDue) {
					nearestDue = c.RI.SelfEmployedDueDate.In(loc)
				}
				dKey := nearestDue.Format("2006-01-02")
				da, ok := dueMap[dKey]
				if !ok {
					da = &dueAccum{}
					dueMap[dKey] = da
				}
				da.amount += c.RI.TotalToPay.Amount
				da.clients++
			}
		}

		seenSig := make(map[signal.Code]bool)
		hasNearK := false
		for _, s := range c.Signals {
			if seenSig[s.Code] {
				continue
			}
			seenSig[s.Code] = true
			if s.Code == signal.CodeMonoNearKCap || s.Code == signal.CodeMonoExclusionKCapExceeded {
				hasNearK = true
			}
			sa, ok := sigMap[string(s.Code)]
			if !ok {
				sa = &sigAccum{code: string(s.Code), severity: s.Severity}
				sigMap[string(s.Code)] = sa
			}
			sa.count++
		}
		if hasNearK {
			nearKCap++
		}

		if c.FirstPublishedAt != nil && !c.FirstPublishedAt.In(loc).After(day10Cutoff) {
			notifiedBeforeDay10++
		}

		if c.Status == StatusNotified {
			notifiedCount++
			if c.PaidAt != nil && !nearestDue.IsZero() {
				paidDay := c.PaidAt.In(loc).Format("2006-01-02")
				dueDay := nearestDue.Format("2006-01-02")
				if paidDay <= dueDay {
					paidOnTime++
				}
			}
		}
	}

	total := len(closings)
	approvedOrNotified := byStatus[string(StatusApproved)] + byStatus[string(StatusNotified)]
	estMinutes := pendingReview["GREEN"]*2 + (pendingReview["YELLOW"]+pendingReview["RED"])*10

	dueDates := make([]DueDateTotal, 0, len(dueMap))
	for dKey, da := range dueMap {
		dueDates = append(dueDates, DueDateTotal{
			Date:    dKey,
			Amount:  money.New(da.amount, money.ARS),
			Clients: da.clients,
		})
	}
	sort.Slice(dueDates, func(i, j int) bool {
		return dueDates[i].Date < dueDates[j].Date
	})

	bySignal := make([]SignalStat, 0, len(sigMap))
	for _, sa := range sigMap {
		label := sa.code
		if meta, ok := signal.MetaFor(signal.Code(sa.code)); ok {
			label = meta.Label
		}
		bySignal = append(bySignal, SignalStat{
			Code:     sa.code,
			Label:    label,
			Severity: sa.severity.String(),
			Count:    sa.count,
		})
	}
	sort.Slice(bySignal, func(i, j int) bool {
		si := parseSeverityRank(bySignal[i].Severity)
		sj := parseSeverityRank(bySignal[j].Severity)
		if si != sj {
			return si > sj
		}
		if bySignal[i].Count != bySignal[j].Count {
			return bySignal[i].Count > bySignal[j].Count
		}
		return bySignal[i].Code < bySignal[j].Code
	})

	byCategory := make([]CategoryStat, 0, len(allMonoCategories)+1)
	for _, cat := range allMonoCategories {
		byCategory = append(byCategory, CategoryStat{
			Category: string(cat),
			Clients:  catCounts[string(cat)],
		})
	}
	byCategory = append(byCategory, CategoryStat{
		Category: "RI",
		Clients:  catCounts["RI"],
	})

	return PeriodStats{
		Period: p.String(),
		Progress: ProgressStats{
			Total:                  total,
			ByStatus:               byStatus,
			ApprovedPct:            ratioBp(approvedOrNotified, total),
			PendingReview:          pendingReview,
			EstimatedReviewMinutes: estMinutes,
		},
		Totals: TotalsStats{
			BilledUSD:      money.New(totalBilledUSD, money.USD),
			BilledARS:      money.New(totalBilledARS, money.ARS),
			ToPayARS:       money.New(totalToPayARS, money.ARS),
			ToPayByDueDate: dueDates,
		},
		Risk: RiskStats{
			BySeverity: bySeverity,
			BySignal:   bySignal,
			ByCategory: byCategory,
			NearKCap:   nearKCap,
		},
		Success: SuccessStats{
			GreenPct:                 ratioBp(bySeverity["GREEN"], total),
			GreenCount:               bySeverity["GREEN"],
			Total:                    total,
			NotifiedBeforeDay10Pct:   ratioBp(notifiedBeforeDay10, total),
			NotifiedBeforeDay10Count: notifiedBeforeDay10,
			NotifiedCount:            notifiedCount,
			PaidOnTimePct:            ratioBp(paidOnTime, notifiedCount),
			PaidOnTimeCount:          paidOnTime,
		},
	}
}

func ratioBp(num, den int) money.BasisPoints {
	if den <= 0 {
		return 0
	}
	return money.BasisPoints((num * 10000) / den)
}

func parseSeverityRank(s string) int {
	switch s {
	case "RED":
		return 2
	case "YELLOW":
		return 1
	default:
		return 0
	}
}
