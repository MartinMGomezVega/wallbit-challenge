package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.NotificationRepository = (*NotificationRepository)(nil)

// NotificationRepository implementa ports.NotificationRepository sobre PostgreSQL.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository crea un NotificationRepository en Postgres.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Get(ctx context.Context, id string) (notification.Notification, error) {
	var n notification.Notification
	var kind string
	var readAt *time.Time
	err := conn(ctx, r.pool).QueryRow(ctx, `
		SELECT id, client_id, closing_id, kind, dedup_key, title, body, created_at, read_at
		FROM notifications WHERE id = $1
	`, id).Scan(&n.ID, &n.ClientID, &n.ClosingID, &kind, &n.Key, &n.Title, &n.Body, &n.CreatedAt, &readAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, domain.ErrNotFound
		}
		return notification.Notification{}, err
	}
	n.Kind = notification.Kind(kind)
	n.CreatedAt = n.CreatedAt.In(period.Location())
	if readAt != nil {
		t := readAt.In(period.Location())
		n.ReadAt = &t
	}
	return n, nil
}

func (r *NotificationRepository) Save(ctx context.Context, n notification.Notification) error {
	_, err := conn(ctx, r.pool).Exec(ctx, `
		INSERT INTO notifications (id, client_id, closing_id, kind, dedup_key, title, body, created_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (client_id, closing_id, kind, dedup_key) DO UPDATE SET
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			read_at = COALESCE(notifications.read_at, EXCLUDED.read_at)
	`, n.ID, n.ClientID, n.ClosingID, string(n.Kind), n.Key, n.Title, n.Body, n.CreatedAt, n.ReadAt)
	return err
}

func (r *NotificationRepository) ByClient(ctx context.Context, clientID string) ([]notification.Notification, error) {
	rows, err := conn(ctx, r.pool).Query(ctx, `
		SELECT id, client_id, closing_id, kind, dedup_key, title, body, created_at, read_at
		FROM notifications WHERE client_id = $1
		ORDER BY created_at DESC, id DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []notification.Notification
	for rows.Next() {
		var n notification.Notification
		var kind string
		var readAt *time.Time
		if err := rows.Scan(&n.ID, &n.ClientID, &n.ClosingID, &kind, &n.Key, &n.Title, &n.Body, &n.CreatedAt, &readAt); err != nil {
			return nil, err
		}
		n.Kind = notification.Kind(kind)
		n.CreatedAt = n.CreatedAt.In(period.Location())
		if readAt != nil {
			t := readAt.In(period.Location())
			n.ReadAt = &t
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id string, at time.Time) error {
	tag, err := conn(ctx, r.pool).Exec(ctx, `
		UPDATE notifications SET read_at = COALESCE(read_at, $2) WHERE id = $1
	`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) ExistsFor(ctx context.Context, clientID, closingID string, kind notification.Kind, key string) (bool, error) {
	var exists bool
	err := conn(ctx, r.pool).QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM notifications WHERE client_id = $1 AND closing_id = $2 AND kind = $3 AND dedup_key = $4
		)
	`, clientID, closingID, string(kind), key).Scan(&exists)
	return exists, err
}
