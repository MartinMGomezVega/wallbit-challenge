package postgres

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.OutboxRepository = (*OutboxRepository)(nil)

// OutboxRepository implementa ports.OutboxRepository sobre PostgreSQL.
type OutboxRepository struct {
	pool *pgxpool.Pool
}

// NewOutboxRepository crea un OutboxRepository en Postgres.
func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) Enqueue(ctx context.Context, m notification.OutboundMessage) error {
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO outbox (id, channel, client_id, subject, body, attempts, sent_at, last_error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, m.ID, string(m.Channel), m.ClientID, m.Subject, m.Body, m.Attempts, m.SentAt, m.LastError, m.CreatedAt)
	return err
}

func (r *OutboxRepository) Pending(ctx context.Context, limit int) ([]notification.OutboundMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, channel, client_id, subject, body, attempts, sent_at, last_error, created_at
		FROM outbox
		WHERE sent_at IS NULL AND attempts < 3
		ORDER BY created_at ASC, id ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []notification.OutboundMessage
	for rows.Next() {
		var m notification.OutboundMessage
		var ch string
		var sentAt *time.Time
		if err := rows.Scan(&m.ID, &ch, &m.ClientID, &m.Subject, &m.Body, &m.Attempts, &sentAt, &m.LastError, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Channel = notification.Channel(ch)
		m.CreatedAt = m.CreatedAt.In(period.Location())
		if sentAt != nil {
			t := sentAt.In(period.Location())
			m.SentAt = &t
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *OutboxRepository) Update(ctx context.Context, m notification.OutboundMessage) error {
	tag, err := conn(ctx, r.pool).Exec(ctx, `
		UPDATE outbox SET attempts = $2, sent_at = $3, last_error = $4 WHERE id = $1
	`, m.ID, m.Attempts, m.SentAt, m.LastError)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
