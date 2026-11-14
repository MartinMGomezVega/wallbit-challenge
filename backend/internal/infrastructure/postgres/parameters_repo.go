package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ParametersRepository = (*ParametersRepository)(nil)

// ParametersRepository implementa ports.ParametersRepository sobre PostgreSQL.
type ParametersRepository struct {
	pool *pgxpool.Pool
}

// NewParametersRepository crea un ParametersRepository en Postgres.
func NewParametersRepository(pool *pgxpool.Pool) *ParametersRepository {
	return &ParametersRepository{pool: pool}
}

func (r *ParametersRepository) Get(ctx context.Context) (parameters.AccountantParameters, error) {
	var data []byte
	err := conn(ctx, r.pool).QueryRow(ctx, `SELECT data FROM parameters WHERE id = 1`).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return parameters.Defaults(), nil
		}
		return parameters.AccountantParameters{}, err
	}
	var p parameters.AccountantParameters
	if err := json.Unmarshal(data, &p); err != nil {
		return parameters.AccountantParameters{}, err
	}
	return p, nil
}

func (r *ParametersRepository) Save(ctx context.Context, p parameters.AccountantParameters) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO parameters (id, data, updated_at, updated_by)
		VALUES (1, $1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			data = EXCLUDED.data,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`, data, p.UpdatedAt, p.UpdatedBy)
	return err
}
