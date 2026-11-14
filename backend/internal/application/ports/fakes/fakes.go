package fakes

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// Clock es un reloj fijo para tests en application.
type Clock struct {
	Current time.Time
}

// Now devuelve el tiempo fijo.
func (c Clock) Now() time.Time {
	return c.Current
}

// Location devuelve la ubicación del tiempo actual.
func (c Clock) Location() *time.Location {
	if c.Current.Location() != nil {
		return c.Current.Location()
	}
	return time.UTC
}

// IDGenerator genera IDs secuenciales para tests.
type IDGenerator struct {
	mu   sync.Mutex
	next map[string]int
}

// NewIDGenerator crea un generador de IDs para tests.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{next: make(map[string]int)}
}

// New genera un ID con prefijo.
func (g *IDGenerator) New(prefix string) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next[prefix]++
	return fmt.Sprintf("%s_%06d", prefix, g.next[prefix])
}

// ClientRepository es un fake para tests unitarios.
type ClientRepository struct {
	Clients map[string]client.Client
}

// NewClientRepository crea un fake de clientes.
func NewClientRepository() *ClientRepository {
	return &ClientRepository{Clients: make(map[string]client.Client)}
}

func (f *ClientRepository) List(_ context.Context) ([]client.Client, error) {
	out := make([]client.Client, 0, len(f.Clients))
	for _, c := range f.Clients {
		out = append(out, c)
	}
	return out, nil
}

func (f *ClientRepository) Get(_ context.Context, id string) (client.Client, error) {
	c, ok := f.Clients[id]
	if !ok {
		return client.Client{}, domain.ErrNotFound
	}
	return c, nil
}

func (f *ClientRepository) Save(_ context.Context, c client.Client) error {
	f.Clients[c.ID] = c
	return nil
}

func (f *ClientRepository) SaveAll(_ context.Context, cs []client.Client) error {
	for _, c := range cs {
		f.Clients[c.ID] = c
	}
	return nil
}

// MovementRepository es un fake para tests unitarios.
type MovementRepository struct {
	Movements map[string]movement.Movement
}

// NewMovementRepository crea un fake vacío.
func NewMovementRepository() *MovementRepository {
	return &MovementRepository{Movements: make(map[string]movement.Movement)}
}

func (f *MovementRepository) Get(_ context.Context, id string) (movement.Movement, error) {
	m, ok := f.Movements[id]
	if !ok {
		return movement.Movement{}, domain.ErrNotFound
	}
	return m, nil
}

func (f *MovementRepository) ByClientBetween(_ context.Context, clientID string, from, to time.Time) ([]movement.Movement, error) {
	var out []movement.Movement
	for _, m := range f.Movements {
		if m.ClientID == clientID {
			if (m.OccurredAt.After(from) || m.OccurredAt.Equal(from)) && (m.OccurredAt.Before(to) || m.OccurredAt.Equal(to)) {
				out = append(out, m)
			}
		}
	}
	return out, nil
}

func (f *MovementRepository) ByClient(_ context.Context, clientID string) ([]movement.Movement, error) {
	var out []movement.Movement
	for _, m := range f.Movements {
		if m.ClientID == clientID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *MovementRepository) Save(_ context.Context, m movement.Movement) error {
	f.Movements[m.ID] = m
	return nil
}

func (f *MovementRepository) SaveAll(_ context.Context, ms []movement.Movement) error {
	for _, m := range ms {
		f.Movements[m.ID] = m
	}
	return nil
}

func (f *MovementRepository) ListAll(_ context.Context) ([]movement.Movement, error) {
	out := make([]movement.Movement, 0, len(f.Movements))
	for _, m := range f.Movements {
		out = append(out, m)
	}
	return out, nil
}

// ClassificationRepository es un fake para tests unitarios.
type ClassificationRepository struct {
	Classifications map[string]movement.Classification
}

// NewClassificationRepository crea un fake de clasificaciones.
func NewClassificationRepository() *ClassificationRepository {
	return &ClassificationRepository{Classifications: make(map[string]movement.Classification)}
}

func (f *ClassificationRepository) Get(_ context.Context, id string) (movement.Classification, bool, error) {
	c, ok := f.Classifications[id]
	return c, ok, nil
}

func (f *ClassificationRepository) Save(_ context.Context, c movement.Classification) error {
	f.Classifications[c.MovementID] = c
	return nil
}

func (f *ClassificationRepository) SaveAll(_ context.Context, cs []movement.Classification) error {
	for _, c := range cs {
		f.Classifications[c.MovementID] = c
	}
	return nil
}

func (f *ClassificationRepository) ByMovementIDs(_ context.Context, ids []string) (map[string]movement.Classification, error) {
	res := make(map[string]movement.Classification)
	for _, id := range ids {
		if c, ok := f.Classifications[id]; ok {
			res[id] = c
		}
	}
	return res, nil
}

// InvoiceRepository fake para tests.
type InvoiceRepository struct {
	Invoices map[string]invoice.Invoice
}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{Invoices: make(map[string]invoice.Invoice)}
}

func (f *InvoiceRepository) Get(_ context.Context, id string) (invoice.Invoice, error) {
	inv, ok := f.Invoices[id]
	if !ok {
		return invoice.Invoice{}, domain.ErrNotFound
	}
	return inv, nil
}

func (f *InvoiceRepository) Exists(_ context.Context, clientID string, key invoice.Key) (bool, error) {
	key.ClientID = clientID
	for _, inv := range f.Invoices {
		if inv.Key() == key {
			return true, nil
		}
	}
	return false, nil
}

func (f *InvoiceRepository) Save(_ context.Context, inv invoice.Invoice) error {
	f.Invoices[inv.ID] = inv
	return nil
}

func (f *InvoiceRepository) SaveAll(_ context.Context, invs []invoice.Invoice) error {
	for _, inv := range invs {
		f.Invoices[inv.ID] = inv
	}
	return nil
}

func (f *InvoiceRepository) ByClient(_ context.Context, clientID string) ([]invoice.Invoice, error) {
	var out []invoice.Invoice
	for _, inv := range f.Invoices {
		if inv.ClientID == clientID {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (f *InvoiceRepository) ByClientBetween(_ context.Context, clientID string, from, to time.Time) ([]invoice.Invoice, error) {
	var out []invoice.Invoice
	for _, inv := range f.Invoices {
		if inv.ClientID == clientID {
			t := inv.IssueDate
			if (t.After(from) || t.Equal(from)) && (t.Before(to) || t.Equal(to)) {
				out = append(out, inv)
			}
		}
	}
	return out, nil
}

// PurchaseInvoiceRepository fake para tests.
type PurchaseInvoiceRepository struct {
	Invoices map[string]invoice.PurchaseInvoice
}

func NewPurchaseInvoiceRepository() *PurchaseInvoiceRepository {
	return &PurchaseInvoiceRepository{Invoices: make(map[string]invoice.PurchaseInvoice)}
}

func (f *PurchaseInvoiceRepository) Save(_ context.Context, pi invoice.PurchaseInvoice) error {
	f.Invoices[pi.ID] = pi
	return nil
}

func (f *PurchaseInvoiceRepository) SaveAll(_ context.Context, pis []invoice.PurchaseInvoice) error {
	for _, pi := range pis {
		f.Invoices[pi.ID] = pi
	}
	return nil
}

func (f *PurchaseInvoiceRepository) ByClient(_ context.Context, clientID string) ([]invoice.PurchaseInvoice, error) {
	var out []invoice.PurchaseInvoice
	for _, pi := range f.Invoices {
		if pi.ClientID == clientID {
			out = append(out, pi)
		}
	}
	return out, nil
}

func (f *PurchaseInvoiceRepository) ByClientBetween(_ context.Context, clientID string, from, to time.Time) ([]invoice.PurchaseInvoice, error) {
	var out []invoice.PurchaseInvoice
	for _, pi := range f.Invoices {
		if pi.ClientID == clientID {
			t := pi.IssueDate
			if (t.After(from) || t.Equal(from)) && (t.Before(to) || t.Equal(to)) {
				out = append(out, pi)
			}
		}
	}
	return out, nil
}

// ClientInputRepository fake para tests.
type ClientInputRepository struct {
	Inputs map[string]clientinput.ClientPeriodInput
}

func NewClientInputRepository() *ClientInputRepository {
	return &ClientInputRepository{Inputs: make(map[string]clientinput.ClientPeriodInput)}
}

func (f *ClientInputRepository) Get(_ context.Context, clientID string, p period.Period) (clientinput.ClientPeriodInput, error) {
	key := fmt.Sprintf("%s:%s", clientID, p.String())
	in, ok := f.Inputs[key]
	if !ok {
		return clientinput.ClientPeriodInput{}, domain.ErrNotFound
	}
	return in, nil
}

func (f *ClientInputRepository) Save(_ context.Context, in clientinput.ClientPeriodInput) error {
	key := fmt.Sprintf("%s:%s", in.ClientID, in.Period.String())
	f.Inputs[key] = in
	return nil
}

func (f *ClientInputRepository) SaveAll(_ context.Context, ins []clientinput.ClientPeriodInput) error {
	for _, in := range ins {
		key := fmt.Sprintf("%s:%s", in.ClientID, in.Period.String())
		f.Inputs[key] = in
	}
	return nil
}

// LatestExclusionDateReader es un fake para tests unitarios.
type LatestExclusionDateReader struct {
	Dates map[string]*time.Time
}

// NewLatestExclusionDateReader crea un fake vacío.
func NewLatestExclusionDateReader() *LatestExclusionDateReader {
	return &LatestExclusionDateReader{Dates: make(map[string]*time.Time)}
}

func (f *LatestExclusionDateReader) LatestExclusionDate(_ context.Context, clientID string) (*time.Time, error) {
	return f.Dates[clientID], nil
}

// ClosingRepository es un fake para tests unitarios.
type ClosingRepository struct {
	mu   sync.RWMutex
	ByID map[string]closing.Closing
	ByCP map[string]string
}

// NewClosingRepository crea un fake de ClosingRepository.
func NewClosingRepository() *ClosingRepository {
	return &ClosingRepository{
		ByID: make(map[string]closing.Closing),
		ByCP: make(map[string]string),
	}
}

func (f *ClosingRepository) Get(_ context.Context, id string) (closing.Closing, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	c, ok := f.ByID[id]
	if !ok {
		return closing.Closing{}, domain.ErrNotFound
	}
	return c, nil
}

func (f *ClosingRepository) ByClientPeriod(_ context.Context, clientID string, p period.Period) (closing.Closing, bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", clientID, p.String())
	id, ok := f.ByCP[key]
	if !ok {
		return closing.Closing{}, false, nil
	}
	c, ok := f.ByID[id]
	return c, ok, nil
}

func (f *ClosingRepository) ListByPeriod(_ context.Context, p period.Period) ([]closing.Closing, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var out []closing.Closing
	for _, c := range f.ByID {
		if c.Period == p {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *ClosingRepository) ByClient(_ context.Context, clientID string) ([]closing.Closing, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var out []closing.Closing
	for _, c := range f.ByID {
		if c.ClientID == clientID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Period.String() > out[j].Period.String()
	})
	return out, nil
}

func (f *ClosingRepository) Save(_ context.Context, c closing.Closing) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fmt.Sprintf("%s:%s", c.ClientID, c.Period.String())
	if oldID, ok := f.ByCP[key]; ok && oldID != c.ID {
		delete(f.ByID, oldID)
	}
	f.ByID[c.ID] = c
	f.ByCP[key] = c.ID
	return nil
}

func (f *ClosingRepository) SaveIfVersion(_ context.Context, c closing.Closing, expectedVersion int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.ByID[c.ID]
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
	if oldID, ok := f.ByCP[key]; ok && oldID != c.ID {
		delete(f.ByID, oldID)
	}
	f.ByID[c.ID] = c
	f.ByCP[key] = c.ID
	return nil
}

func (f *ClosingRepository) LatestExclusionDate(_ context.Context, clientID string) (*time.Time, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var latest *time.Time
	for _, c := range f.ByID {
		if c.ClientID == clientID && c.Mono != nil && c.Mono.Excluded && c.Mono.ExclusionDate != nil {
			if latest == nil || c.Mono.ExclusionDate.After(*latest) {
				d := *c.Mono.ExclusionDate
				latest = &d
			}
		}
	}
	return latest, nil
}

// TaxScaleProvider fake para tests.
type TaxScaleProvider struct {
	Scale tax.Scale
}

func (f *TaxScaleProvider) MonotributoScale(_ context.Context, _ time.Time) (tax.Scale, error) {
	return f.Scale, nil
}

// DueDateProvider fake para tests.
type DueDateProvider struct {
	Dates ports.DueDates
}

func (f *DueDateProvider) DueDates(_ context.Context, _ period.Period, cuit client.CUIT) (ports.DueDates, error) {
	if cuit == "invalid-cuit" {
		return ports.DueDates{}, fmt.Errorf("cuit inválida")
	}
	return f.Dates, nil
}

// ExchangeRateProvider fake para tests.
type ExchangeRateProvider struct {
	Default ports.Quote
	Quotes  map[string]ports.Quote
}

func (f *ExchangeRateProvider) USDQuote(_ context.Context, day time.Time) (ports.Quote, error) {
	key := day.Format("2006-01-02")
	if f.Quotes != nil {
		if q, ok := f.Quotes[key]; ok {
			return q, nil
		}
	}
	if !f.Default.Day.IsZero() {
		return f.Default, nil
	}
	return ports.Quote{Day: day, Buy: 150000, Sell: 155000}, nil
}

// ParametersRepository fake para tests.
type ParametersRepository struct {
	Params parameters.AccountantParameters
}

func (f *ParametersRepository) Get(_ context.Context) (parameters.AccountantParameters, error) {
	return f.Params, nil
}

func (f *ParametersRepository) Save(_ context.Context, p parameters.AccountantParameters) error {
	f.Params = p
	return nil
}

// NotificationRepository fake para tests.
type NotificationRepository struct {
	mu            sync.RWMutex
	Notifications map[string]notification.Notification
}

// NewNotificationRepository crea un fake de NotificationRepository.
func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		Notifications: make(map[string]notification.Notification),
	}
}

func (f *NotificationRepository) Get(_ context.Context, id string) (notification.Notification, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	n, ok := f.Notifications[id]
	if !ok {
		return notification.Notification{}, domain.ErrNotFound
	}
	return n, nil
}

func (f *NotificationRepository) Save(_ context.Context, n notification.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Notifications[n.ID] = n
	return nil
}

func (f *NotificationRepository) ByClient(_ context.Context, clientID string) ([]notification.Notification, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make([]notification.Notification, 0)
	for _, n := range f.Notifications {
		if n.ClientID == clientID {
			out = append(out, n)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (f *NotificationRepository) MarkRead(_ context.Context, id string, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	n, ok := f.Notifications[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.ReadAt = &at
	f.Notifications[id] = n
	return nil
}

func (f *NotificationRepository) ExistsFor(_ context.Context, clientID, closingID string, kind notification.Kind, key string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, n := range f.Notifications {
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

// OutboxRepository fake para tests.
type OutboxRepository struct {
	mu       sync.RWMutex
	Messages map[string]notification.OutboundMessage
}

// NewOutboxRepository crea un fake de OutboxRepository.
func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{
		Messages: make(map[string]notification.OutboundMessage),
	}
}

func (f *OutboxRepository) Enqueue(_ context.Context, m notification.OutboundMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Messages[m.ID] = m
	return nil
}

func (f *OutboxRepository) Pending(_ context.Context, limit int) ([]notification.OutboundMessage, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make([]notification.OutboundMessage, 0)
	for _, m := range f.Messages {
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

func (f *OutboxRepository) Update(_ context.Context, m notification.OutboundMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.Messages[m.ID]; !ok {
		return domain.ErrNotFound
	}
	f.Messages[m.ID] = m
	return nil
}

// NotificationSender fake para tests.
type NotificationSender struct {
	mu     sync.Mutex
	Sent   []notification.OutboundMessage
	FailN  int
	Errors []error
}

// NewNotificationSender crea un fake de NotificationSender.
func NewNotificationSender() *NotificationSender {
	return &NotificationSender{
		Sent: make([]notification.OutboundMessage, 0),
	}
}

func (f *NotificationSender) Send(_ context.Context, m notification.OutboundMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.FailN > 0 {
		f.FailN--
		err := fmt.Errorf("sender error: remaining failures %d", f.FailN)
		f.Errors = append(f.Errors, err)
		return err
	}
	f.Sent = append(f.Sent, m)
	return nil
}

// TxManager es un fake de ports.TxManager para tests unitarios.
type TxManager struct{}

// NewTxManager crea un TxManager fake.
func NewTxManager() *TxManager {
	return &TxManager{}
}

func (f *TxManager) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ClosingQuery es un fake de ports.ClosingQuery sobre ClosingRepository y ClientRepository.
type ClosingQuery struct {
	Closings *ClosingRepository
	Clients  *ClientRepository
}

// NewClosingQuery crea un fake de ClosingQuery.
func NewClosingQuery(closings *ClosingRepository, clients *ClientRepository) *ClosingQuery {
	return &ClosingQuery{Closings: closings, Clients: clients}
}

func (f *ClosingQuery) rows(ctx context.Context, p period.Period) ([]closing.Row, error) {
	list, err := f.Closings.ListByPeriod(ctx, p)
	if err != nil {
		return nil, err
	}
	rows := make([]closing.Row, 0, len(list))
	for _, c := range list {
		cli, err := f.Clients.Get(ctx, c.ClientID)
		if err != nil {
			continue
		}
		rows = append(rows, closing.BuildRow(c, cli))
	}
	return rows, nil
}

func (f *ClosingQuery) Page(ctx context.Context, p period.Period, filter closing.Filter, s closing.Sort, page ports.PageRequest) (ports.PageResult, error) {
	rows, err := f.rows(ctx, p)
	if err != nil {
		return ports.PageResult{}, err
	}
	filtered := closing.Apply(rows, filter, s)
	total := len(filtered)
	pNum := page.Page
	if pNum < 1 {
		pNum = 1
	}
	pSize := page.PageSize
	if pSize < 1 {
		pSize = 50
	}
	start := (pNum - 1) * pSize
	if start >= total {
		return ports.PageResult{Items: []closing.Row{}, Total: total}, nil
	}
	end := start + pSize
	if end > total {
		end = total
	}
	return ports.PageResult{Items: filtered[start:end], Total: total}, nil
}

func (f *ClosingQuery) IDs(ctx context.Context, p period.Period, filter closing.Filter, s closing.Sort) ([]string, error) {
	rows, err := f.rows(ctx, p)
	if err != nil {
		return nil, err
	}
	filtered := closing.Apply(rows, filter, s)
	ids := make([]string, len(filtered))
	for i, r := range filtered {
		ids[i] = r.ClosingID
	}
	return ids, nil
}

func (f *ClosingQuery) Summary(ctx context.Context, p period.Period, filter closing.Filter) (closing.Summary, error) {
	rows, err := f.rows(ctx, p)
	if err != nil {
		return closing.Summary{}, err
	}
	return closing.ComputeSummary(closing.Apply(rows, filter, closing.Sort{})), nil
}
