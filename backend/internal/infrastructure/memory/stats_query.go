package memory

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

var _ ports.StatsQuery = (*StatsQuery)(nil)

// StatsQuery implementa ports.StatsQuery en memoria usando closing.ComputeStats.
type StatsQuery struct {
	closings *ClosingRepository
}

// NewStatsQuery crea un StatsQuery en memoria.
func NewStatsQuery(closings *ClosingRepository) *StatsQuery {
	return &StatsQuery{closings: closings}
}

func (q *StatsQuery) PeriodStats(ctx context.Context, p period.Period) (closing.PeriodStats, error) {
	list, err := q.closings.ListByPeriod(ctx, p)
	if err != nil {
		return closing.PeriodStats{}, err
	}
	return closing.ComputeStats(list, p), nil
}
