package memory_test

import (
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/repotest"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestMemoryContract(t *testing.T) {
	newBundle := func(t *testing.T) repotest.Bundle {
		t.Helper()
		clients := memory.NewClientRepository()
		closings := memory.NewClosingRepository()
		return repotest.Bundle{
			TxManager:        memory.NewTxManager(),
			Clients:          clients,
			Movements:        memory.NewMovementRepository(),
			Classifications:  memory.NewClassificationRepository(),
			Invoices:         memory.NewInvoiceRepository(),
			PurchaseInvoices: memory.NewPurchaseInvoiceRepository(),
			ClientInputs:     memory.NewClientInputRepository(),
			Closings:         closings,
			ClosingQuery:     memory.NewClosingQuery(closings, clients),
			StatsQuery:       memory.NewStatsQuery(closings),
			Notifications:    memory.NewNotificationRepository(),
			Outbox:           memory.NewOutboxRepository(),
			Parameters:       memory.NewParametersRepository(),
		}
	}

	repotest.RunAll(t, newBundle, false)
}

func TestMemoryClosingQueryCompare(t *testing.T) {
	clients := memory.NewClientRepository()
	closings := memory.NewClosingRepository()
	q := memory.NewClosingQuery(closings, clients)
	repotest.CompareClosingQueries(t, q, q, period.MustParse("2026-10"))
}
