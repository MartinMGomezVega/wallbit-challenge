package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ClosingQuery = (*ClosingQuery)(nil)

// ClosingQuery implementa ports.ClosingQuery mediante SQL parametrizado sobre PostgreSQL.
type ClosingQuery struct {
	pool *pgxpool.Pool
}

// NewClosingQuery crea un ClosingQuery en Postgres.
func NewClosingQuery(pool *pgxpool.Pool) *ClosingQuery {
	return &ClosingQuery{pool: pool}
}

func buildWhere(p period.Period, f closing.Filter) (string, []any) {
	conds := []string{"c.period = $1"}
	args := []any{p.String()}
	idx := 2

	if len(f.Severities) > 0 {
		sevs := make([]int16, len(f.Severities))
		for i, s := range f.Severities {
			sevs[i] = int16(s)
		}
		conds = append(conds, fmt.Sprintf("c.severity = ANY($%d)", idx))
		args = append(args, sevs)
		idx++
	}

	if len(f.Statuses) > 0 {
		sts := make([]string, len(f.Statuses))
		for i, st := range f.Statuses {
			sts[i] = string(st)
		}
		conds = append(conds, fmt.Sprintf("c.status = ANY($%d)", idx))
		args = append(args, sts)
		idx++
	}

	if f.Regime != "" {
		conds = append(conds, fmt.Sprintf("c.regime = $%d", idx))
		args = append(args, string(f.Regime))
		idx++
	}

	if len(f.Signals) > 0 {
		codes := make([]string, len(f.Signals))
		for i, c := range f.Signals {
			codes[i] = string(c)
		}
		conds = append(conds, fmt.Sprintf("EXISTS (SELECT 1 FROM closing_signals cs WHERE cs.closing_id = c.id AND cs.code = ANY($%d))", idx))
		args = append(args, codes)
		idx++
	}

	cleanQ := strings.ToLower(strings.TrimSpace(f.Query))
	if cleanQ != "" {
		cleanCUIT := strings.ReplaceAll(strings.ReplaceAll(cleanQ, "-", ""), " ", "")
		conds = append(conds, fmt.Sprintf("(cl.name_search LIKE '%%' || $%d || '%%' OR replace(replace(lower(cl.cuit), '-', ''), ' ', '') LIKE '%%' || $%d || '%%')", idx, idx+1))
		args = append(args, cleanQ, cleanCUIT)
		idx += 2
	}

	return strings.Join(conds, " AND "), args
}

func buildOrderBy(s closing.Sort) string {
	dir := strings.ToLower(s.Dir)
	switch s.By {
	case "name":
		if dir == "desc" {
			return `cl.name_search COLLATE "C" DESC, c.client_id ASC`
		}
		return `cl.name_search COLLATE "C" ASC, c.client_id ASC`
	case "billed":
		if dir == "asc" {
			return `c.billed_ars_cents ASC, c.client_id ASC`
		}
		return `c.billed_ars_cents DESC, c.client_id ASC`
	case "toPay":
		if dir == "asc" {
			return `c.to_pay_cents ASC, c.client_id ASC`
		}
		return `c.to_pay_cents DESC, c.client_id ASC`
	case "pctKCap":
		if dir == "asc" {
			return `c.pct_k_cap_bp ASC, c.client_id ASC`
		}
		return `c.pct_k_cap_bp DESC, c.client_id ASC`
	default:
		if dir == "asc" {
			return `c.severity ASC, c.primary_priority DESC, cl.name_search COLLATE "C" DESC, c.client_id ASC`
		}
		return `c.severity DESC, c.primary_priority ASC, cl.name_search COLLATE "C" ASC, c.client_id ASC`
	}
}

func (q *ClosingQuery) Page(ctx context.Context, p period.Period, filter closing.Filter, s closing.Sort, page ports.PageRequest) (ports.PageResult, error) {
	db := conn(ctx, q.pool)
	whereSQL, args := buildWhere(p, filter)

	var total int
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM closings c JOIN clients cl ON cl.id = c.client_id WHERE %s`, whereSQL)
	if err := db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return ports.PageResult{}, err
	}

	pNum := page.Page
	if pNum < 1 {
		pNum = 1
	}
	pSize := page.PageSize
	if pSize < 1 {
		pSize = 50
	}
	if pSize > 500 {
		pSize = 500
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pSize - 1) / pSize
	}
	offset := (pNum - 1) * pSize
	if offset >= total {
		return ports.PageResult{
			Items:      []closing.Row{},
			Total:      total,
			Page:       pNum,
			PageSize:   pSize,
			TotalPages: totalPages,
		}, nil
	}

	argIdx := len(args) + 1
	orderSQL := buildOrderBy(s)
	pageSQL := fmt.Sprintf(`
		SELECT c.id, c.client_id, cl.name, cl.cuit, c.regime, COALESCE(c.category, ''), c.status, c.version, c.severity, c.description, c.paid_at, c.payload
		FROM closings c
		JOIN clients cl ON cl.id = c.client_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderSQL, argIdx, argIdx+1)
	pageArgs := append(append([]any{}, args...), pSize, offset)

	rows, err := db.Query(ctx, pageSQL, pageArgs...)
	if err != nil {
		return ports.PageResult{}, err
	}
	defer rows.Close()

	items := make([]closing.Row, 0)
	for rows.Next() {
		var c closing.Closing
		var cliName, cliCUIT, regimeStr, catStr, statusStr string
		var sev int16
		var paidAt *time.Time
		var payloadBytes []byte
		if err := rows.Scan(&c.ID, &c.ClientID, &cliName, &cliCUIT, &regimeStr, &catStr, &statusStr, &c.Version, &sev, &c.Description, &paidAt, &payloadBytes); err != nil {
			return ports.PageResult{}, err
		}
		c.Period = p
		c.Status = closing.Status(statusStr)
		c.Severity = signal.Severity(sev)
		if paidAt != nil {
			t := paidAt.In(period.Location())
			c.PaidAt = &t
		}
		var pl closingPayload
		if err := json.Unmarshal(payloadBytes, &pl); err != nil {
			return ports.PageResult{}, err
		}
		c.Mono = pl.Mono
		if c.Mono != nil {
			c.Mono.FeeDueDate = c.Mono.FeeDueDate.In(period.Location())
		}
		c.RI = pl.RI
		if c.RI != nil {
			c.RI.VatDueDate = c.RI.VatDueDate.In(period.Location())
			c.RI.SelfEmployedDueDate = c.RI.SelfEmployedDueDate.In(period.Location())
		}
		c.Signals = pl.Signals
		c.Reconciliation = pl.Reconciliation

		cli := client.Client{
			ID:                  c.ClientID,
			Name:                cliName,
			CUIT:                client.CUIT(cliCUIT),
			Regime:              client.Regime(regimeStr),
			MonotributoCategory: tax.Category(catStr),
		}
		items = append(items, closing.BuildRow(c, cli))
	}
	if err := rows.Err(); err != nil {
		return ports.PageResult{}, err
	}
	return ports.PageResult{
		Items:      items,
		Total:      total,
		Page:       pNum,
		PageSize:   pSize,
		TotalPages: totalPages,
	}, nil
}

func (q *ClosingQuery) IDs(ctx context.Context, p period.Period, filter closing.Filter, s closing.Sort) ([]string, error) {
	db := conn(ctx, q.pool)
	whereSQL, args := buildWhere(p, filter)
	orderSQL := buildOrderBy(s)

	querySQL := fmt.Sprintf(`
		SELECT c.id
		FROM closings c
		JOIN clients cl ON cl.id = c.client_id
		WHERE %s
		ORDER BY %s
	`, whereSQL, orderSQL)

	rows, err := db.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (q *ClosingQuery) Summary(ctx context.Context, p period.Period, filter closing.Filter) (closing.Summary, error) {
	db := conn(ctx, q.pool)
	noSevFilter := filter
	noSevFilter.Severities = nil
	whereSQL, args := buildWhere(p, noSevFilter)

	sumSQL := fmt.Sprintf(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE c.severity = 2),
			COUNT(*) FILTER (WHERE c.severity = 1),
			COUNT(*) FILTER (WHERE c.severity = 0),
			COUNT(*) FILTER (WHERE c.status = 'PENDING_CLIENT_INPUT'),
			COUNT(*) FILTER (WHERE c.status = 'READY_FOR_REVIEW'),
			COUNT(*) FILTER (WHERE c.status = 'APPROVED'),
			COUNT(*) FILTER (WHERE c.status = 'NOTIFIED')
		FROM closings c
		JOIN clients cl ON cl.id = c.client_id
		WHERE %s
	`, whereSQL)

	var total, red, yellow, green int
	var pendingInput, readyReview, approved, notified int
	if err := db.QueryRow(ctx, sumSQL, args...).Scan(
		&total, &red, &yellow, &green,
		&pendingInput, &readyReview, &approved, &notified,
	); err != nil {
		return closing.Summary{}, err
	}

	out := closing.Summary{
		Total: total,
		BySeverity: map[string]int{
			"RED":    red,
			"YELLOW": yellow,
			"GREEN":  green,
		},
		BySignal: make(map[string]int),
		ByStatus: map[closing.Status]int{
			closing.StatusPendingClientInput: pendingInput,
			closing.StatusReadyForReview:     readyReview,
			closing.StatusApproved:           approved,
			closing.StatusNotified:           notified,
		},
	}

	sigSQL := fmt.Sprintf(`
		SELECT cs.code, COUNT(*)
		FROM closings c
		JOIN clients cl ON cl.id = c.client_id
		JOIN closing_signals cs ON cs.closing_id = c.id
		WHERE %s
		GROUP BY cs.code
	`, whereSQL)
	rows, err := db.Query(ctx, sigSQL, args...)
	if err != nil {
		return closing.Summary{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		var cnt int
		if err := rows.Scan(&code, &cnt); err != nil {
			return closing.Summary{}, err
		}
		out.BySignal[code] = cnt
	}
	return out, rows.Err()
}
