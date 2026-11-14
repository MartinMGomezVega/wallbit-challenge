package ports

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// PageRequest especifica la página y tamaño de página para consultas paginadas.
type PageRequest struct {
	Page     int
	PageSize int
}

// PageResult contiene los resultados paginados de ClosingQuery.Page.
type PageResult struct {
	Items      []closing.Row
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// ClosingQuery resuelve consultas de tablero (paginado, IDs, resumen) sobre cierres de un período.
type ClosingQuery interface {
	Page(ctx context.Context, p period.Period, filter closing.Filter, sort closing.Sort, page PageRequest) (PageResult, error)
	IDs(ctx context.Context, p period.Period, filter closing.Filter, sort closing.Sort) ([]string, error)
	Summary(ctx context.Context, p period.Period, filter closing.Filter) (closing.Summary, error)
}
