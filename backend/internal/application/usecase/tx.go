package usecase

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
)

func withinTx(ctx context.Context, tx ports.TxManager, fn func(context.Context) error) error {
	if tx == nil {
		return fn(ctx)
	}
	return tx.WithinTx(ctx, fn)
}
