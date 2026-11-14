package closing_test

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestComputeStats_EmptyPeriod(t *testing.T) {
	p := period.MustParse("2026-10")
	st := closing.ComputeStats(nil, p)

	if st.Period != "2026-10" || st.Progress.Total != 0 {
		t.Fatalf("unexpected empty stats: %+v", st)
	}
	if st.Progress.ApprovedPct != 0 || st.Success.GreenPct != 0 || st.Success.NotifiedBeforeDay10Pct != 0 || st.Success.PaidOnTimePct != 0 {
		t.Fatalf("expected 0 percentages for empty period, got %+v", st.Success)
	}
	if len(st.Risk.ByCategory) != 12 {
		t.Fatalf("expected 12 categories (A-K + RI), got %d", len(st.Risk.ByCategory))
	}
	if st.Risk.ByCategory[0].Category != "A" || st.Risk.ByCategory[11].Category != "RI" {
		t.Fatalf("unexpected category ordering: %+v", st.Risk.ByCategory)
	}
}

func TestComputeStats_AllDefinitionsAndEdgeCases(t *testing.T) {
	loc := period.Location()
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, loc)
	due20 := time.Date(2026, 11, 20, 0, 0, 0, 0, loc)
	due10 := time.Date(2026, 11, 10, 0, 0, 0, 0, loc)

	// 1. RED pending review, near K cap + excluded
	c1 := closing.NewClosing("c1", "cli1", p, true, "h1", signal.SeverityRed,
		[]signal.Signal{
			{Code: signal.CodeMonoExclusionKCapExceeded, Severity: signal.SeverityRed, Priority: 1},
			{Code: signal.CodeMonoExclusionKCapExceeded, Severity: signal.SeverityRed, Priority: 1}, // duplicada en mismo cierre: cuenta 1 vez
		}, "Excluido",
		&tax.MonotributoAssessment{CurrentCategory: tax.CatK, Fee: money.New(500000, money.ARS), FeeDueDate: due20, BilledInPeriodUSD: money.New(100000, money.USD), BilledInPeriodARS: money.New(100000000, money.ARS)},
		nil, closing.Reconciliation{}, now)

	// 2. YELLOW pending client input, near K cap
	c2 := closing.NewClosing("c2", "cli2", p, false, "h2", signal.SeverityYellow,
		[]signal.Signal{{Code: signal.CodeMonoNearKCap, Severity: signal.SeverityYellow, Priority: 5}}, "Cerca K",
		&tax.MonotributoAssessment{CurrentCategory: tax.CatI, Fee: money.New(300000, money.ARS), FeeDueDate: due20, BilledInPeriodUSD: money.New(50000, money.USD), BilledInPeriodARS: money.New(50000000, money.ARS)},
		nil, closing.Reconciliation{}, now)

	// 3. GREEN notified on day 8 (<= 10), paid on time (day 19 <= 20)
	c3 := closing.NewClosing("c3", "cli3", p, true, "h3", signal.SeverityGreen, nil, "OK",
		&tax.MonotributoAssessment{CurrentCategory: tax.CatA, Fee: money.New(100000, money.ARS), FeeDueDate: due20, BilledInPeriodUSD: money.New(10000, money.USD), BilledInPeriodARS: money.New(10000000, money.ARS)},
		nil, closing.Reconciliation{}, now)
	_ = c3.Approve(closing.Approval{ID: "a3", Version: 1, ResultHash: "h3", ApprovedBy: "acc", ApprovedAt: time.Date(2026, 11, 8, 12, 0, 0, 0, loc)})
	_ = c3.Publish(closing.Published{Version: 1, Title: "t", Body: "b", PublishedAt: time.Date(2026, 11, 8, 12, 0, 0, 0, loc)})
	_ = c3.MarkPaid(time.Date(2026, 11, 19, 15, 0, 0, 0, loc))

	// 4. GREEN RI notified on day 12 (> 10), not paid on time (paid day 15 > due10)
	c4 := closing.NewClosing("c4", "cli4", p, true, "h4", signal.SeverityGreen, nil, "OK RI",
		nil, &tax.RIAssessment{TotalToPay: money.New(800000, money.ARS), VatDueDate: due20, SelfEmployedDueDate: due10, BilledInPeriodUSD: money.New(200000, money.USD), BilledInPeriodARS: money.New(200000000, money.ARS)},
		closing.Reconciliation{}, now)
	_ = c4.Approve(closing.Approval{ID: "a4", Version: 1, ResultHash: "h4", ApprovedBy: "acc", ApprovedAt: time.Date(2026, 11, 12, 12, 0, 0, 0, loc)})
	_ = c4.Publish(closing.Published{Version: 1, Title: "t", Body: "b", PublishedAt: time.Date(2026, 11, 12, 12, 0, 0, 0, loc)})
	_ = c4.MarkPaid(time.Date(2026, 11, 15, 10, 0, 0, 0, loc))

	st := closing.ComputeStats([]closing.Closing{c1, c2, c3, c4}, p)

	if st.Progress.Total != 4 || st.Progress.ApprovedPct != 5000 {
		t.Fatalf("progress = %+v", st.Progress)
	}
	// 1 RED (10) + 1 YELLOW (10) + 0 GREEN (2) = 20 min
	if st.Progress.EstimatedReviewMinutes != 20 {
		t.Fatalf("estimatedReviewMinutes = %d, want 20", st.Progress.EstimatedReviewMinutes)
	}
	if st.Totals.BilledUSD.Amount != 360000 || st.Totals.ToPayARS.Amount != 1700000 {
		t.Fatalf("totals = %+v", st.Totals)
	}
	if len(st.Totals.ToPayByDueDate) != 2 || st.Totals.ToPayByDueDate[0].Date != "2026-11-10" || st.Totals.ToPayByDueDate[1].Date != "2026-11-20" {
		t.Fatalf("toPayByDueDate = %+v", st.Totals.ToPayByDueDate)
	}
	if st.Risk.NearKCap != 2 {
		t.Fatalf("nearKCap = %d, want 2", st.Risk.NearKCap)
	}
	if len(st.Risk.BySignal) != 2 || st.Risk.BySignal[0].Code != string(signal.CodeMonoExclusionKCapExceeded) || st.Risk.BySignal[0].Count != 1 {
		t.Fatalf("bySignal = %+v", st.Risk.BySignal)
	}
	if st.Success.GreenPct != 5000 || st.Success.NotifiedBeforeDay10Pct != 2500 || st.Success.PaidOnTimePct != 5000 {
		t.Fatalf("success = %+v", st.Success)
	}
}
