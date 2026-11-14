package memory

import (
	"context"
	"sync"
)

type txContextKey struct{}

// TxManager implementa ports.TxManager en memoria mediante un lock global.
type TxManager struct {
	mu sync.Mutex
}

// NewTxManager crea un TxManager en memoria.
func NewTxManager() *TxManager {
	return &TxManager{}
}

// WithinTx ejecuta fn bajo el lock global de transacciones en memoria.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if ctx.Value(txContextKey{}) != nil {
		return fn(ctx)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(context.WithValue(ctx, txContextKey{}, true))
}
