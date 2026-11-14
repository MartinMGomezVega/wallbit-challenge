package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
)

var _ ports.MovementRepository = (*MovementRepository)(nil)

// MovementRepository almacena movimientos en memoria indexados por ID y por cliente.
type MovementRepository struct {
	mu        sync.RWMutex
	movements map[string]movement.Movement
	byClient  map[string][]string // clientID -> slice de movement IDs
}

// NewMovementRepository crea una nueva instancia de MovementRepository.
func NewMovementRepository() *MovementRepository {
	return &MovementRepository{
		movements: make(map[string]movement.Movement),
		byClient:  make(map[string][]string),
	}
}

// Get devuelve un movimiento por ID o ErrNotFound si no existe.
func (r *MovementRepository) Get(_ context.Context, id string) (movement.Movement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.movements[id]
	if !ok {
		return movement.Movement{}, domain.ErrNotFound
	}
	return m, nil
}

// Save almacena o actualiza un movimiento.
func (r *MovementRepository) Save(ctx context.Context, m movement.Movement) error {
	return r.SaveAll(ctx, []movement.Movement{m})
}

// SaveAll guarda una colección de movimientos y actualiza los índices.
func (r *MovementRepository) SaveAll(_ context.Context, ms []movement.Movement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range ms {
		_, exists := r.movements[m.ID]
		r.movements[m.ID] = m
		if !exists {
			r.byClient[m.ClientID] = append(r.byClient[m.ClientID], m.ID)
		}
	}
	return nil
}

// ByClient devuelve todos los movimientos de un cliente ordenados por fecha ascendente.
func (r *MovementRepository) ByClient(_ context.Context, clientID string) ([]movement.Movement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byClient[clientID]
	out := make([]movement.Movement, 0, len(ids))
	for _, id := range ids {
		if m, ok := r.movements[id]; ok {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}

// ByClientBetween devuelve los movimientos de un cliente en el rango [from, to] (inclusive).
func (r *MovementRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]movement.Movement, error) {
	all, err := r.ByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	out := make([]movement.Movement, 0, len(all))
	for _, m := range all {
		t := m.OccurredAt
		if (t.After(from) || t.Equal(from)) && (t.Before(to) || t.Equal(to)) {
			out = append(out, m)
		}
	}
	return out, nil
}

// ListAll devuelve todos los movimientos en memoria.
func (r *MovementRepository) ListAll(_ context.Context) ([]movement.Movement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]movement.Movement, 0, len(r.movements))
	for _, m := range r.movements {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}
