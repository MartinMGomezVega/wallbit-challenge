package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.MovementRepository = (*MovementRepository)(nil)

// MovementRepository implementa ports.MovementRepository sobre PostgreSQL.
type MovementRepository struct {
	pool *pgxpool.Pool
}

// NewMovementRepository crea un MovementRepository en Postgres.
func NewMovementRepository(pool *pgxpool.Pool) *MovementRepository {
	return &MovementRepository{pool: pool}
}

func scanMovement(row pgx.Row) (movement.Movement, error) {
	var m movement.Movement
	var typ, cur, cpKind string
	var amt, fx int64
	err := row.Scan(
		&m.ID, &m.ClientID, &typ, &amt, &cur, &fx, &m.OccurredAt,
		&m.Counterparty.Name, &m.Counterparty.Country, &cpKind, &m.Counterparty.IsOwnAccount, &m.Description,
	)
	if err != nil {
		return movement.Movement{}, err
	}
	m.Type = movement.MovementType(typ)
	m.Amount = money.New(amt, money.Currency(cur))
	m.FXRate = money.Rate(fx)
	m.OccurredAt = m.OccurredAt.In(period.Location())
	m.Counterparty.Kind = movement.CounterpartyKind(cpKind)
	return m, nil
}

func (r *MovementRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]movement.Movement, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, type, amount_cents, currency, fx_rate_scaled, occurred_at, cp_name, cp_country, cp_kind, cp_own, description
		FROM movements
		WHERE client_id = $1 AND occurred_at >= $2 AND occurred_at <= $3
		ORDER BY occurred_at ASC, id ASC
	`, clientID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []movement.Movement
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MovementRepository) ByClient(ctx context.Context, clientID string) ([]movement.Movement, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, type, amount_cents, currency, fx_rate_scaled, occurred_at, cp_name, cp_country, cp_kind, cp_own, description
		FROM movements
		WHERE client_id = $1
		ORDER BY occurred_at ASC, id ASC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []movement.Movement
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MovementRepository) ListAll(ctx context.Context) ([]movement.Movement, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, type, amount_cents, currency, fx_rate_scaled, occurred_at, cp_name, cp_country, cp_kind, cp_own, description
		FROM movements
		ORDER BY occurred_at ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []movement.Movement
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MovementRepository) Get(ctx context.Context, id string) (movement.Movement, error) {
	m, err := scanMovement(conn(ctx, r.pool).QueryRow(ctx, `
		SELECT id, client_id, type, amount_cents, currency, fx_rate_scaled, occurred_at, cp_name, cp_country, cp_kind, cp_own, description
		FROM movements WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return movement.Movement{}, domain.ErrNotFound
		}
		return movement.Movement{}, err
	}
	return m, nil
}

func (r *MovementRepository) Save(ctx context.Context, m movement.Movement) error {
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO movements (id, client_id, type, amount_cents, currency, fx_rate_scaled, occurred_at, cp_name, cp_country, cp_kind, cp_own, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			client_id = EXCLUDED.client_id,
			type = EXCLUDED.type,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			fx_rate_scaled = EXCLUDED.fx_rate_scaled,
			occurred_at = EXCLUDED.occurred_at,
			cp_name = EXCLUDED.cp_name,
			cp_country = EXCLUDED.cp_country,
			cp_kind = EXCLUDED.cp_kind,
			cp_own = EXCLUDED.cp_own,
			description = EXCLUDED.description
	`, m.ID, m.ClientID, string(m.Type), m.Amount.Amount, string(m.Amount.Currency), int64(m.FXRate), m.OccurredAt,
		m.Counterparty.Name, m.Counterparty.Country, string(m.Counterparty.Kind), m.Counterparty.IsOwnAccount, m.Description)
	return err
}

func (r *MovementRepository) SaveAll(ctx context.Context, ms []movement.Movement) error {
	for _, m := range ms {
		if err := r.Save(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
