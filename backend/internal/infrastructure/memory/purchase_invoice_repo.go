package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
)

var _ ports.PurchaseInvoiceRepository = (*PurchaseInvoiceRepository)(nil)

// PurchaseInvoiceRepository almacena facturas de compra en memoria.
type PurchaseInvoiceRepository struct {
	mu       sync.RWMutex
	invoices map[string]invoice.PurchaseInvoice
	byClient map[string][]string
}

// NewPurchaseInvoiceRepository crea una nueva instancia de PurchaseInvoiceRepository.
func NewPurchaseInvoiceRepository() *PurchaseInvoiceRepository {
	return &PurchaseInvoiceRepository{
		invoices: make(map[string]invoice.PurchaseInvoice),
		byClient: make(map[string][]string),
	}
}

// Save almacena o actualiza una factura de compra.
func (r *PurchaseInvoiceRepository) Save(ctx context.Context, pi invoice.PurchaseInvoice) error {
	return r.SaveAll(ctx, []invoice.PurchaseInvoice{pi})
}

// SaveAll guarda una colección de facturas de compra.
func (r *PurchaseInvoiceRepository) SaveAll(_ context.Context, pis []invoice.PurchaseInvoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, pi := range pis {
		_, exists := r.invoices[pi.ID]
		r.invoices[pi.ID] = pi
		if !exists {
			r.byClient[pi.ClientID] = append(r.byClient[pi.ClientID], pi.ID)
		}
	}
	return nil
}

// ByClient devuelve todas las facturas de compra de un cliente ordenadas por fecha.
func (r *PurchaseInvoiceRepository) ByClient(_ context.Context, clientID string) ([]invoice.PurchaseInvoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byClient[clientID]
	out := make([]invoice.PurchaseInvoice, 0, len(ids))
	for _, id := range ids {
		if pi, ok := r.invoices[id]; ok {
			out = append(out, pi)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].IssueDate.Before(out[j].IssueDate)
	})
	return out, nil
}

// ByClientBetween devuelve las facturas de compra de un cliente en [from, to] (inclusive).
func (r *PurchaseInvoiceRepository) ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.PurchaseInvoice, error) {
	all, err := r.ByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	out := make([]invoice.PurchaseInvoice, 0, len(all))
	for _, pi := range all {
		t := pi.IssueDate
		if (t.After(from) || t.Equal(from)) && (t.Before(to) || t.Equal(to)) {
			out = append(out, pi)
		}
	}
	return out, nil
}

