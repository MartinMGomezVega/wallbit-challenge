package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
)

var _ ports.InvoiceRepository = (*InvoiceRepository)(nil)

// InvoiceRepository almacena facturas emitidas en memoria indexadas por ID, por cliente y por clave única.
type InvoiceRepository struct {
	mu       sync.RWMutex
	invoices map[string]invoice.Invoice
	byClient map[string][]string
	byKey    map[string]string // key.String() -> invoice ID
}

// NewInvoiceRepository crea una nueva instancia de InvoiceRepository.
func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{
		invoices: make(map[string]invoice.Invoice),
		byClient: make(map[string][]string),
		byKey:    make(map[string]string),
	}
}

// Get obtiene una factura por ID o devuelve domain.ErrNotFound.
func (r *InvoiceRepository) Get(_ context.Context, id string) (invoice.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, ok := r.invoices[id]
	if !ok {
		return invoice.Invoice{}, domain.ErrNotFound
	}
	return inv, nil
}

// Exists verifica si ya existe una factura para el cliente con la misma clave natural.
func (r *InvoiceRepository) Exists(_ context.Context, clientID string, key invoice.Key) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key.ClientID = clientID
	_, exists := r.byKey[key.String()]
	return exists, nil
}

// Save almacena o actualiza una factura.
func (r *InvoiceRepository) Save(ctx context.Context, inv invoice.Invoice) error {
	return r.SaveAll(ctx, []invoice.Invoice{inv})
}

// SaveAll guarda una colección de facturas y actualiza índices.
func (r *InvoiceRepository) SaveAll(_ context.Context, invs []invoice.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, inv := range invs {
		_, exists := r.invoices[inv.ID]
		r.invoices[inv.ID] = inv
		r.byKey[inv.Key().String()] = inv.ID
		if !exists {
			r.byClient[inv.ClientID] = append(r.byClient[inv.ClientID], inv.ID)
		}
	}
	return nil
}

// ByClient devuelve las facturas de un cliente ordenadas por fecha de emisión ascendente.
func (r *InvoiceRepository) ByClient(_ context.Context, clientID string) ([]invoice.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byClient[clientID]
	out := make([]invoice.Invoice, 0, len(ids))
	for _, id := range ids {
		if inv, ok := r.invoices[id]; ok {
			out = append(out, inv)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].IssueDate.Before(out[j].IssueDate)
	})
	return out, nil
}

// ByClientBetween devuelve las facturas de un cliente con fecha en [from, to] (inclusive).
func (r *InvoiceRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.Invoice, error) {
	all, err := r.ByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	out := make([]invoice.Invoice, 0, len(all))
	for _, inv := range all {
		t := inv.IssueDate
		if (t.After(from) || t.Equal(from)) && (t.Before(to) || t.Equal(to)) {
			out = append(out, inv)
		}
	}
	return out, nil
}
