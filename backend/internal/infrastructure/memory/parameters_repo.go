// Package memory implementa repositorios en memoria protegidos con sync.RWMutex
// (architecture.md §2.2, §2.3).
package memory

import (
	"context"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
)

var _ ports.ParametersRepository = (*ParametersRepository)(nil)

// ParametersRepository almacena los parámetros en memoria.
type ParametersRepository struct {
	mu     sync.RWMutex
	params parameters.AccountantParameters
}

// NewParametersRepository crea un repositorio inicializado con Defaults().
func NewParametersRepository() *ParametersRepository {
	return &ParametersRepository{
		params: parameters.Defaults(),
	}
}

// Get devuelve una copia de los parámetros actuales.
func (r *ParametersRepository) Get(_ context.Context) (parameters.AccountantParameters, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.params, nil
}

// Save actualiza los parámetros en memoria.
func (r *ParametersRepository) Save(_ context.Context, p parameters.AccountantParameters) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.params = p
	return nil
}
