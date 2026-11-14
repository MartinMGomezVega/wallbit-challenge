package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
)

var _ ports.OutboxRepository = (*OutboxRepository)(nil)

// OutboxRepository almacena mensajes salientes en memoria de forma segura para concurrencia.
type OutboxRepository struct {
	mu       sync.RWMutex
	messages map[string]notification.OutboundMessage
}

// NewOutboxRepository crea un nuevo repositorio de outbox en memoria.
func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{
		messages: make(map[string]notification.OutboundMessage),
	}
}

// Enqueue agrega un mensaje al outbox.
func (r *OutboxRepository) Enqueue(_ context.Context, m notification.OutboundMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages[m.ID] = m
	return nil
}

// Pending devuelve los mensajes no enviados ordenados ascendentemente por fecha de creación.
func (r *OutboxRepository) Pending(_ context.Context, limit int) ([]notification.OutboundMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]notification.OutboundMessage, 0)
	for _, m := range r.messages {
		if m.SentAt == nil && m.Attempts < 3 {
			out = append(out, m)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// Update actualiza el estado de un mensaje saliente (intentos, SentAt, LastError).
func (r *OutboxRepository) Update(_ context.Context, m notification.OutboundMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[m.ID]; !ok {
		return domain.ErrNotFound
	}
	r.messages[m.ID] = m
	return nil
}
