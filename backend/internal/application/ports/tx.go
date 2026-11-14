package ports

import "context"

// TxManager ejecuta una función dentro de una unidad de trabajo transaccional.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
