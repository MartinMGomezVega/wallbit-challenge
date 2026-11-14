package postgres

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.StatsQuery = (*StatsQuery)(nil)

// StatsQuery implementa ports.StatsQuery usando agregaciones SQL en PostgreSQL.
type StatsQuery struct {
	pool *pgxpool.Pool
}

// NewStatsQuery crea un StatsQuery en Postgres.
func NewStatsQuery(pool *pgxpool.Pool) *StatsQuery {
	return &StatsQuery{pool: pool}
}

func (q *StatsQuery) PeriodStats(ctx context.Context, p period.Period) (closing.PeriodStats, error) {
	db := conn(ctx, q.pool)
	loc := period.Location()
	nextP := p.Next()
	day10Cutoff := time.Date(nextP.Year, time.Month(nextP.Month), 10, 23, 59, 59, 999999999, loc)

	var (
		total, stPendingInput, stReady, stApproved, stNotified int
		pendRed, pendYellow, pendGreen                         int
		sevRed, sevYellow, sevGreen                            int
		billedUSD, billedARS, toPayARS                         int64
		nearKCap, notifiedBeforeDay10, paidOnTime              int
	)

	err := db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE c.status = 'PENDING_CLIENT_INPUT'),
			COUNT(*) FILTER (WHERE c.status = 'READY_FOR_REVIEW'),
			COUNT(*) FILTER (WHERE c.status = 'APPROVED'),
			COUNT(*) FILTER (WHERE c.status = 'NOTIFIED'),
			COUNT(*) FILTER (WHERE c.status IN ('PENDING_CLIENT_INPUT','READY_FOR_REVIEW') AND c.severity = 2),
			COUNT(*) FILTER (WHERE c.status IN ('PENDING_CLIENT_INPUT','READY_FOR_REVIEW') AND c.severity = 1),
			COUNT(*) FILTER (WHERE c.status IN ('PENDING_CLIENT_INPUT','READY_FOR_REVIEW') AND c.severity = 0),
			COUNT(*) FILTER (WHERE c.severity = 2),
			COUNT(*) FILTER (WHERE c.severity = 1),
			COUNT(*) FILTER (WHERE c.severity = 0),
			COALESCE(SUM(c.billed_usd_cents), 0),
			COALESCE(SUM(c.billed_ars_cents), 0),
			COALESCE(SUM(c.to_pay_cents), 0),
			COUNT(*) FILTER (WHERE EXISTS (
				SELECT 1 FROM closing_signals cs
				WHERE cs.closing_id = c.id AND cs.code IN ('MONO_NEAR_K_CAP', 'MONO_EXCLUSION_K_CAP_EXCEEDED')
			)),
			COUNT(*) FILTER (WHERE c.first_published_at IS NOT NULL AND c.first_published_at <= $2),
			COUNT(*) FILTER (
				WHERE c.status = 'NOTIFIED'
				  AND c.paid_at IS NOT NULL
				  AND c.to_pay_due IS NOT NULL
				  AND (c.paid_at AT TIME ZONE 'America/Argentina/Buenos_Aires')::date <= (c.to_pay_due AT TIME ZONE 'America/Argentina/Buenos_Aires')::date
			)
		FROM closings c
		WHERE c.period = $1
	`, p.String(), day10Cutoff).Scan(
		&total, &stPendingInput, &stReady, &stApproved, &stNotified,
		&pendRed, &pendYellow, &pendGreen,
		&sevRed, &sevYellow, &sevGreen,
		&billedUSD, &billedARS, &toPayARS,
		&nearKCap, &notifiedBeforeDay10, &paidOnTime,
	)
	if err != nil {
		return closing.PeriodStats{}, err
	}

	// Due dates
	dueRows, err := db.Query(ctx, `
		SELECT to_char((to_pay_due AT TIME ZONE 'America/Argentina/Buenos_Aires')::date, 'YYYY-MM-DD') AS d,
		       COALESCE(SUM(to_pay_cents), 0),
		       COUNT(*)
		FROM closings
		WHERE period = $1 AND to_pay_due IS NOT NULL
		GROUP BY d
		ORDER BY d ASC
	`, p.String())
	if err != nil {
		return closing.PeriodStats{}, err
	}
	defer dueRows.Close()

	dueDates := make([]closing.DueDateTotal, 0)
	for dueRows.Next() {
		var d string
		var amt int64
		var cnt int
		if err := dueRows.Scan(&d, &amt, &cnt); err != nil {
			return closing.PeriodStats{}, err
		}
		dueDates = append(dueDates, closing.DueDateTotal{
			Date:    d,
			Amount:  money.New(amt, money.ARS),
			Clients: cnt,
		})
	}
	if err := dueRows.Err(); err != nil {
		return closing.PeriodStats{}, err
	}

	// Signals
	sigRows, err := db.Query(ctx, `
		SELECT cs.code, cs.severity, COUNT(*) AS cnt
		FROM closings c
		JOIN closing_signals cs ON cs.closing_id = c.id
		WHERE c.period = $1
		GROUP BY cs.code, cs.severity
		ORDER BY cs.severity DESC, cnt DESC, cs.code ASC
	`, p.String())
	if err != nil {
		return closing.PeriodStats{}, err
	}
	defer sigRows.Close()

	bySignal := make([]closing.SignalStat, 0)
	for sigRows.Next() {
		var code string
		var sev int16
		var cnt int
		if err := sigRows.Scan(&code, &sev, &cnt); err != nil {
			return closing.PeriodStats{}, err
		}
		label := code
		if meta, ok := signal.MetaFor(signal.Code(code)); ok {
			label = meta.Label
		}
		bySignal = append(bySignal, closing.SignalStat{
			Code:     code,
			Label:    label,
			Severity: signal.Severity(sev).String(),
			Count:    cnt,
		})
	}
	if err := sigRows.Err(); err != nil {
		return closing.PeriodStats{}, err
	}

	// Categories
	catCounts := map[string]int{
		"A": 0, "B": 0, "C": 0, "D": 0, "E": 0, "F": 0,
		"G": 0, "H": 0, "I": 0, "J": 0, "K": 0, "RI": 0,
	}
	catRows, err := db.Query(ctx, `
		SELECT CASE WHEN regime = 'RI' THEN 'RI' ELSE category END AS cat, COUNT(*)
		FROM closings
		WHERE period = $1
		GROUP BY cat
	`, p.String())
	if err != nil {
		return closing.PeriodStats{}, err
	}
	defer catRows.Close()
	for catRows.Next() {
		var cat string
		var cnt int
		if err := catRows.Scan(&cat, &cnt); err != nil {
			return closing.PeriodStats{}, err
		}
		if _, ok := catCounts[cat]; ok {
			catCounts[cat] = cnt
		}
	}
	if err := catRows.Err(); err != nil {
		return closing.PeriodStats{}, err
	}

	catOrder := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "RI"}
	byCategory := make([]closing.CategoryStat, 0, len(catOrder))
	for _, cat := range catOrder {
		byCategory = append(byCategory, closing.CategoryStat{
			Category: cat,
			Clients:  catCounts[cat],
		})
	}

	approvedOrNotified := stApproved + stNotified
	ratio := func(num, den int) money.BasisPoints {
		if den <= 0 {
			return 0
		}
		return money.BasisPoints((num * 10000) / den)
	}

	return closing.PeriodStats{
		Period: p.String(),
		Progress: closing.ProgressStats{
			Total: total,
			ByStatus: map[string]int{
				string(closing.StatusPendingClientInput): stPendingInput,
				string(closing.StatusReadyForReview):     stReady,
				string(closing.StatusApproved):           stApproved,
				string(closing.StatusNotified):           stNotified,
			},
			ApprovedPct: ratio(approvedOrNotified, total),
			PendingReview: map[string]int{
				"RED":    pendRed,
				"YELLOW": pendYellow,
				"GREEN":  pendGreen,
			},
			EstimatedReviewMinutes: pendGreen*2 + (pendYellow+pendRed)*10,
		},
		Totals: closing.TotalsStats{
			BilledUSD:      money.New(billedUSD, money.USD),
			BilledARS:      money.New(billedARS, money.ARS),
			ToPayARS:       money.New(toPayARS, money.ARS),
			ToPayByDueDate: dueDates,
		},
		Risk: closing.RiskStats{
			BySeverity: map[string]int{
				"RED":    sevRed,
				"YELLOW": sevYellow,
				"GREEN":  sevGreen,
			},
			BySignal:   bySignal,
			ByCategory: byCategory,
			NearKCap:   nearKCap,
		},
		Success: closing.SuccessStats{
			GreenPct:                 ratio(sevGreen, total),
			GreenCount:               sevGreen,
			Total:                    total,
			NotifiedBeforeDay10Pct:   ratio(notifiedBeforeDay10, total),
			NotifiedBeforeDay10Count: notifiedBeforeDay10,
			NotifiedCount:            stNotified,
			PaidOnTimePct:            ratio(paidOnTime, stNotified),
			PaidOnTimeCount:          paidOnTime,
		},
	}, nil
}
