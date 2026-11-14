package memory

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ClosingQuery implementa ports.ClosingQuery en memoria usando closing.Apply y closing.ComputeSummary.
type ClosingQuery struct {
	closings ports.ClosingRepository
	clients  ports.ClientRepository
}

// NewClosingQuery crea una instancia de ClosingQuery en memoria.
func NewClosingQuery(closings ports.ClosingRepository, clients ports.ClientRepository) *ClosingQuery {
	return &ClosingQuery{closings: closings, clients: clients}
}

func (q *ClosingQuery) buildRows(ctx context.Context, p period.Period) ([]closing.Row, error) {
	closingsList, err := q.closings.ListByPeriod(ctx, p)
	if err != nil {
		return nil, err
	}

	clientsList, err := q.clients.List(ctx)
	if err != nil {
		return nil, err
	}

	clientMap := make(map[string]client.Client, len(clientsList))
	for _, c := range clientsList {
		clientMap[c.ID] = c
	}

	rows := make([]closing.Row, 0, len(closingsList))
	for _, c := range closingsList {
		cli := clientMap[c.ClientID]
		rows = append(rows, closing.BuildRow(c, cli))
	}
	return rows, nil
}

// Page devuelve una página de filas de cierres filtradas y ordenadas.
func (q *ClosingQuery) Page(ctx context.Context, p period.Period, filter closing.Filter, sort closing.Sort, pageReq ports.PageRequest) (ports.PageResult, error) {
	rows, err := q.buildRows(ctx, p)
	if err != nil {
		return ports.PageResult{}, err
	}

	filtered := closing.Apply(rows, filter, sort)
	total := len(filtered)

	pageSize := pageReq.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}

	page := pageReq.Page
	if page <= 0 {
		page = 1
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	items := filtered[start:end]
	if items == nil {
		items = []closing.Row{}
	}

	return ports.PageResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// IDs devuelve los IDs de los cierres filtrados y ordenados, sin paginación.
func (q *ClosingQuery) IDs(ctx context.Context, p period.Period, filter closing.Filter, sort closing.Sort) ([]string, error) {
	rows, err := q.buildRows(ctx, p)
	if err != nil {
		return nil, err
	}

	filtered := closing.Apply(rows, filter, sort)
	ids := make([]string, len(filtered))
	for i, r := range filtered {
		ids[i] = r.ClosingID
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}

// Summary calcula los contadores de severidad, señal y estado ignorando el filtro de severidad.
func (q *ClosingQuery) Summary(ctx context.Context, p period.Period, filter closing.Filter) (closing.Summary, error) {
	rows, err := q.buildRows(ctx, p)
	if err != nil {
		return closing.Summary{}, err
	}

	noSevFilter := filter
	noSevFilter.Severities = nil

	filtered := closing.Apply(rows, noSevFilter, closing.Sort{})
	return closing.ComputeSummary(filtered), nil
}
