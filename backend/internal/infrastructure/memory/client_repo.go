package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
)

var _ ports.ClientRepository = (*ClientRepository)(nil)

// ClientRepository almacena clientes en memoria de forma concurrente.
type ClientRepository struct {
	mu      sync.RWMutex
	clients map[string]client.Client
}

// NewClientRepository crea un nuevo repositorio de clientes en memoria.
func NewClientRepository() *ClientRepository {
	return &ClientRepository{
		clients: make(map[string]client.Client),
	}
}

// List devuelve todos los clientes ordenados por ID.
func (r *ClientRepository) List(_ context.Context) ([]client.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]client.Client, 0, len(r.clients))
	for _, c := range r.clients {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// Get obtiene un cliente por su ID.
func (r *ClientRepository) Get(_ context.Context, id string) (client.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.clients[id]
	if !ok {
		return client.Client{}, domain.ErrNotFound
	}
	return c, nil
}

// Save guarda o actualiza un cliente.
func (r *ClientRepository) Save(_ context.Context, c client.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[c.ID] = c
	return nil
}

// SaveAll guarda un conjunto de clientes.
func (r *ClientRepository) SaveAll(_ context.Context, cs []client.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range cs {
		r.clients[c.ID] = c
	}
	return nil
}
