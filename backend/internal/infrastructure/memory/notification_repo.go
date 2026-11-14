package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
)

var _ ports.NotificationRepository = (*NotificationRepository)(nil)

// NotificationRepository almacena notificaciones en memoria de forma segura para concurrencia.
type NotificationRepository struct {
	mu            sync.RWMutex
	notifications map[string]notification.Notification
}

// NewNotificationRepository crea un nuevo repositorio de notificaciones en memoria.
func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		notifications: make(map[string]notification.Notification),
	}
}

// Get obtiene una notificación por su ID o devuelve domain.ErrNotFound.
func (r *NotificationRepository) Get(_ context.Context, id string) (notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.notifications[id]
	if !ok {
		return notification.Notification{}, domain.ErrNotFound
	}
	return n, nil
}

// Save guarda o actualiza una notificación.
func (r *NotificationRepository) Save(_ context.Context, n notification.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[n.ID] = n
	return nil
}

// ByClient devuelve las notificaciones de un cliente ordenadas descendentemente por fecha de creación.
func (r *NotificationRepository) ByClient(_ context.Context, clientID string) ([]notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]notification.Notification, 0)
	for _, n := range r.notifications {
		if n.ClientID == clientID {
			out = append(out, n)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// MarkRead marca una notificación como leída con la fecha especificada.
func (r *NotificationRepository) MarkRead(_ context.Context, id string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.ReadAt = &at
	r.notifications[id] = n
	return nil
}

// ExistsFor verifica si ya existe una notificación que coincida con cliente, cierre, tipo y clave de deduplicación.
func (r *NotificationRepository) ExistsFor(_ context.Context, clientID, closingID string, kind notification.Kind, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, n := range r.notifications {
		if n.ClientID != clientID || n.Kind != kind {
			continue
		}
		if closingID != "" && n.ClosingID != closingID {
			continue
		}
		if key != "" && n.Key != key {
			continue
		}
		return true, nil
	}
	return false, nil
}
