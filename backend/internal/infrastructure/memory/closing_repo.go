package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

var _ ports.ClosingRepository = (*ClosingRepository)(nil)
var _ ports.LatestExclusionDateReader = (*ClosingRepository)(nil)

// ClosingRepository almacena cierres contables mensuales en memoria de forma segura para concurrencia.
type ClosingRepository struct {
	mu             sync.RWMutex
	byID           map[string]closing.Closing
	byClientPeriod map[string]string // clave clientID:period -> id
}

// NewClosingRepository crea una nueva instancia del repositorio en memoria.
func NewClosingRepository() *ClosingRepository {
	return &ClosingRepository{
		byID:           make(map[string]closing.Closing),
		byClientPeriod: make(map[string]string),
	}
}

// Get obtiene un cierre por su ID. Devuelve domain.ErrNotFound si no existe.
func (r *ClosingRepository) Get(_ context.Context, id string) (closing.Closing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.byID[id]
	if !ok {
		return closing.Closing{}, domain.ErrNotFound
	}
	return c, nil
}

// ByClientPeriod busca el cierre de un cliente para un período específico.
func (r *ClosingRepository) ByClientPeriod(_ context.Context, clientID string, p period.Period) (closing.Closing, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", clientID, p.String())
	id, ok := r.byClientPeriod[key]
	if !ok {
		return closing.Closing{}, false, nil
	}
	c, ok := r.byID[id]
	if !ok {
		return closing.Closing{}, false, nil
	}
	return c, true, nil
}

// ListByPeriod devuelve todos los cierres de un período, ordenados por ClientID.
func (r *ClosingRepository) ListByPeriod(_ context.Context, p period.Period) ([]closing.Closing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []closing.Closing
	for _, c := range r.byID {
		if c.Period == p {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ClientID < out[j].ClientID
	})
	return out, nil
}

// ByClient devuelve todos los cierres de un cliente ordenados por período descendente.
func (r *ClosingRepository) ByClient(_ context.Context, clientID string) ([]closing.Closing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []closing.Closing
	for _, c := range r.byID {
		if c.ClientID == clientID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Period.String() > out[j].Period.String()
	})
	return out, nil
}

// Save persiste un cierre haciendo upsert por (clientID, period).
func (r *ClosingRepository) Save(_ context.Context, c closing.Closing) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", c.ClientID, c.Period.String())
	if oldID, ok := r.byClientPeriod[key]; ok && oldID != c.ID {
		delete(r.byID, oldID)
	}

	r.byID[c.ID] = c
	r.byClientPeriod[key] = c.ID
	return nil
}

// SaveIfVersion persiste un cierre solo si la versión actual coincide con expectedVersion y no fue ya aprobado/notificado en esa versión.
func (r *ClosingRepository) SaveIfVersion(_ context.Context, c closing.Closing, expectedVersion int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[c.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if existing.Version != expectedVersion {
		return domain.ErrVersionConflict
	}
	if (c.Status == closing.StatusApproved || c.Status == closing.StatusNotified) &&
		(existing.Status == closing.StatusApproved || existing.Status == closing.StatusNotified) {
		return domain.ErrVersionConflict
	}

	key := fmt.Sprintf("%s:%s", c.ClientID, c.Period.String())
	if oldID, ok := r.byClientPeriod[key]; ok && oldID != c.ID {
		delete(r.byID, oldID)
	}
	r.byID[c.ID] = c
	r.byClientPeriod[key] = c.ID
	return nil
}

// LatestExclusionDate devuelve la fecha de exclusión más reciente detectada en los cierres del cliente.
func (r *ClosingRepository) LatestExclusionDate(_ context.Context, clientID string) (*time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *time.Time
	for _, c := range r.byID {
		if c.ClientID == clientID && c.Mono != nil && c.Mono.Excluded && c.Mono.ExclusionDate != nil {
			if latest == nil || c.Mono.ExclusionDate.After(*latest) {
				d := *c.Mono.ExclusionDate
				latest = &d
			}
		}
	}
	return latest, nil
}

// Periods devuelve todos los períodos distintos registrados, ordenados descendente.
func (r *ClosingRepository) Periods(_ context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var out []string
	for _, c := range r.byID {
		pStr := c.Period.String()
		if !seen[pStr] {
			seen[pStr] = true
			out = append(out, pStr)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i] > out[j]
	})
	if out == nil {
		out = []string{}
	}
	return out, nil
}
