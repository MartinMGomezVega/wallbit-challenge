package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ClientInputRepository = (*ClientInputRepository)(nil)

// ClientInputRepository implementa ports.ClientInputRepository sobre PostgreSQL.
type ClientInputRepository struct {
	pool *pgxpool.Pool
}

// NewClientInputRepository crea un ClientInputRepository en Postgres.
func NewClientInputRepository(pool *pgxpool.Pool) *ClientInputRepository {
	return &ClientInputRepository{pool: pool}
}

func (r *ClientInputRepository) Get(ctx context.Context, clientID string, p period.Period) (clientinput.ClientPeriodInput, error) {
	db := conn(ctx, r.pool)
	var out clientinput.ClientPeriodInput
	out.ClientID = clientID
	out.Period = p
	var confirmedAt *time.Time
	err := db.QueryRow(ctx, `
		SELECT declared_none, confirmed_at
		FROM client_period_inputs WHERE client_id = $1 AND period = $2
	`, clientID, p.String()).Scan(&out.DeclaredNoExternalIncome, &confirmedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return clientinput.ClientPeriodInput{}, domain.ErrNotFound
		}
		return clientinput.ClientPeriodInput{}, err
	}
	if confirmedAt != nil {
		t := confirmedAt.In(period.Location())
		out.ConfirmedAt = &t
	}

	rows, err := db.Query(ctx, `
		SELECT amount_cents, description
		FROM external_income_items
		WHERE client_id = $1 AND period = $2
		ORDER BY seq ASC
	`, clientID, p.String())
	if err != nil {
		return clientinput.ClientPeriodInput{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var amt int64
		var desc string
		if err := rows.Scan(&amt, &desc); err != nil {
			return clientinput.ClientPeriodInput{}, err
		}
		out.ExternalIncome = append(out.ExternalIncome, clientinput.ExternalIncomeItem{
			Amount:      money.New(amt, money.USD),
			Description: desc,
		})
	}
	return out, rows.Err()
}

func (r *ClientInputRepository) Save(ctx context.Context, input clientinput.ClientPeriodInput) error {
	db := conn(ctx, r.pool)
	per := input.Period.String()
	if _, err := db.Exec(ctx, `
		INSERT INTO client_period_inputs (client_id, period, declared_none, confirmed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (client_id, period) DO UPDATE SET
			declared_none = EXCLUDED.declared_none,
			confirmed_at = EXCLUDED.confirmed_at
	`, input.ClientID, per, input.DeclaredNoExternalIncome, input.ConfirmedAt); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `DELETE FROM external_income_items WHERE client_id = $1 AND period = $2`, input.ClientID, per); err != nil {
		return err
	}
	for seq, item := range input.ExternalIncome {
		if _, err := db.Exec(ctx, `
			INSERT INTO external_income_items (client_id, period, seq, amount_cents, description)
			VALUES ($1, $2, $3, $4, $5)
		`, input.ClientID, per, seq, item.Amount.Amount, item.Description); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientInputRepository) SaveAll(ctx context.Context, inputs []clientinput.ClientPeriodInput) error {
	for _, in := range inputs {
		if err := r.Save(ctx, in); err != nil {
			return err
		}
	}
	return nil
}
