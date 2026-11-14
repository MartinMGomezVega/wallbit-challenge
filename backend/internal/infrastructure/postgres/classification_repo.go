package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.ClassificationRepository = (*ClassificationRepository)(nil)

// ClassificationRepository implementa ports.ClassificationRepository sobre PostgreSQL.
type ClassificationRepository struct {
	pool *pgxpool.Pool
}

// NewClassificationRepository crea un ClassificationRepository en Postgres.
func NewClassificationRepository(pool *pgxpool.Pool) *ClassificationRepository {
	return &ClassificationRepository{pool: pool}
}

func (r *ClassificationRepository) Get(ctx context.Context, movementID string) (movement.Classification, bool, error) {
	var c movement.Classification
	var kind, conf, rule string
	var overriddenAt *time.Time
	err := conn(ctx, r.pool).QueryRow(ctx, `
		SELECT movement_id, kind, confidence, rule, overridden_by, overridden_at, note
		FROM classifications WHERE movement_id = $1
	`, movementID).Scan(&c.MovementID, &kind, &conf, &rule, &c.OverriddenBy, &overriddenAt, &c.Note)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return movement.Classification{}, false, nil
		}
		return movement.Classification{}, false, err
	}
	c.Kind = movement.ClassificationKind(kind)
	c.Confidence = movement.Confidence(conf)
	c.Rule = movement.RuleCode(rule)
	if overriddenAt != nil {
		t := overriddenAt.In(period.Location())
		c.OverriddenAt = &t
	}
	return c, true, nil
}

func (r *ClassificationRepository) Save(ctx context.Context, c movement.Classification) error {
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO classifications (movement_id, kind, confidence, rule, overridden_by, overridden_at, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (movement_id) DO UPDATE SET
			kind = EXCLUDED.kind,
			confidence = EXCLUDED.confidence,
			rule = EXCLUDED.rule,
			overridden_by = EXCLUDED.overridden_by,
			overridden_at = EXCLUDED.overridden_at,
			note = EXCLUDED.note
	`, c.MovementID, string(c.Kind), string(c.Confidence), string(c.Rule), c.OverriddenBy, c.OverriddenAt, c.Note)
	return err
}

func (r *ClassificationRepository) SaveAll(ctx context.Context, cs []movement.Classification) error {
	for _, c := range cs {
		if err := r.Save(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClassificationRepository) ByMovementIDs(ctx context.Context, ids []string) (map[string]movement.Classification, error) {
	out := make(map[string]movement.Classification)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT movement_id, kind, confidence, rule, overridden_by, overridden_at, note
		FROM classifications WHERE movement_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c movement.Classification
		var kind, conf, rule string
		var overriddenAt *time.Time
		if err := rows.Scan(&c.MovementID, &kind, &conf, &rule, &c.OverriddenBy, &overriddenAt, &c.Note); err != nil {
			return nil, err
		}
		c.Kind = movement.ClassificationKind(kind)
		c.Confidence = movement.Confidence(conf)
		c.Rule = movement.RuleCode(rule)
		if overriddenAt != nil {
			t := overriddenAt.In(period.Location())
			c.OverriddenAt = &t
		}
		out[c.MovementID] = c
	}
	return out, rows.Err()
}
