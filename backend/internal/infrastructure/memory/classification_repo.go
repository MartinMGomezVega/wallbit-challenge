package memory

import (
	"context"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
)

var _ ports.ClassificationRepository = (*ClassificationRepository)(nil)

// ClassificationRepository almacena clasificaciones de movimientos en memoria.
type ClassificationRepository struct {
	mu              sync.RWMutex
	classifications map[string]movement.Classification
}

// NewClassificationRepository crea una nueva instancia de ClassificationRepository.
func NewClassificationRepository() *ClassificationRepository {
	return &ClassificationRepository{
		classifications: make(map[string]movement.Classification),
	}
}

// Get obtiene la clasificación de un movimiento por su ID.
func (r *ClassificationRepository) Get(_ context.Context, movementID string) (movement.Classification, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.classifications[movementID]
	return c, ok, nil
}

// Save almacena o actualiza la clasificación de un movimiento.
func (r *ClassificationRepository) Save(_ context.Context, c movement.Classification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.classifications[c.MovementID] = c
	return nil
}

// SaveAll guarda una colección de clasificaciones.
func (r *ClassificationRepository) SaveAll(_ context.Context, cs []movement.Classification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range cs {
		r.classifications[c.MovementID] = c
	}
	return nil
}

// ByMovementIDs devuelve un mapa con las clasificaciones existentes para los IDs solicitados.
func (r *ClassificationRepository) ByMovementIDs(_ context.Context, ids []string) (map[string]movement.Classification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make(map[string]movement.Classification, len(ids))
	for _, id := range ids {
		if c, ok := r.classifications[id]; ok {
			res[id] = c
		}
	}
	return res, nil
}
