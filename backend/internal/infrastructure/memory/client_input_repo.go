package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

var _ ports.ClientInputRepository = (*ClientInputRepository)(nil)

// ClientInputRepository almacena declaraciones de clientes por período en memoria.
type ClientInputRepository struct {
	mu     sync.RWMutex
	inputs map[string]clientinput.ClientPeriodInput // key: fmt.Sprintf("%s:%s", clientID, p.String())
}

// NewClientInputRepository crea un nuevo repositorio de inputs de cliente en memoria.
func NewClientInputRepository() *ClientInputRepository {
	return &ClientInputRepository{
		inputs: make(map[string]clientinput.ClientPeriodInput),
	}
}

func inputKey(clientID string, p period.Period) string {
	return fmt.Sprintf("%s:%s", clientID, p.String())
}

// Get obtiene el input del cliente para el período, o devuelve domain.ErrNotFound si no existe.
func (r *ClientInputRepository) Get(_ context.Context, clientID string, p period.Period) (clientinput.ClientPeriodInput, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	in, ok := r.inputs[inputKey(clientID, p)]
	if !ok {
		return clientinput.ClientPeriodInput{}, domain.ErrNotFound
	}
	return in, nil
}

// Save almacena o actualiza la declaración del cliente para un período.
func (r *ClientInputRepository) Save(_ context.Context, input clientinput.ClientPeriodInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.inputs[inputKey(input.ClientID, input.Period)] = input
	return nil
}

// SaveAll guarda una colección de inputs.
func (r *ClientInputRepository) SaveAll(_ context.Context, inputs []clientinput.ClientPeriodInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, in := range inputs {
		r.inputs[inputKey(in.ClientID, in.Period)] = in
	}
	return nil
}
