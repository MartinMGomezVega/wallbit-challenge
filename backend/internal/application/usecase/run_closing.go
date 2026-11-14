package usecase

import (
	"context"
	"sort"
	"sync"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// FailedItem registra el error de procesamiento de un cliente individual.
type FailedItem struct {
	ClientID string `json:"clientId"`
	Error    string `json:"error"`
}

// RunSummary resume el resultado de procesar el cierre masivo del período (spec 007 criterio 5).
type RunSummary struct {
	Processed int          `json:"processed"`
	Created   int          `json:"created"`
	Updated   int          `json:"updated"`
	Unchanged int          `json:"unchanged"`
	Failed    []FailedItem `json:"failed"`
}

// RunMonthlyClosing coordina la ejecución paralela del cierre mensual de todos los clientes
// utilizando un worker pool concurrente e idempotente.
type RunMonthlyClosing struct {
	Clients       ports.ClientRepository
	Compute       *ComputeClosing
	Workers       int
	ActionNotices *CreateActionRequiredNotices
}

// Execute procesa a todos los clientes para el período especificado.
func (uc *RunMonthlyClosing) Execute(ctx context.Context, p period.Period) (RunSummary, error) {
	clients, err := uc.Clients.List(ctx)
	if err != nil {
		return RunSummary{}, err
	}

	summary := RunSummary{
		Failed: []FailedItem{},
	}

	if len(clients) == 0 {
		return summary, nil
	}

	numWorkers := uc.Workers
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if numWorkers > len(clients) {
		numWorkers = len(clients)
	}

	jobs := make(chan string, len(clients))
	for _, c := range clients {
		jobs <- c.ID
	}
	close(jobs)

	type workerResult struct {
		clientID string
		status   ComputeStatus
		err      error
	}

	results := make(chan workerResult, len(clients))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for clientID := range jobs {
				select {
				case <-ctx.Done():
					results <- workerResult{clientID: clientID, err: ctx.Err()}
				default:
					res, err := uc.Compute.Execute(ctx, clientID, p)
					if err != nil {
						results <- workerResult{clientID: clientID, err: err}
					} else {
						results <- workerResult{clientID: clientID, status: res.Status}
					}
				}
			}
		}()
	}

	wg.Wait()
	close(results)

	for res := range results {
		summary.Processed++
		if res.err != nil {
			summary.Failed = append(summary.Failed, FailedItem{
				ClientID: res.clientID,
				Error:    res.err.Error(),
			})
			continue
		}

		switch res.status {
		case ComputeStatusCreated:
			summary.Created++
		case ComputeStatusUpdated:
			summary.Updated++
		case ComputeStatusUnchanged:
			summary.Unchanged++
		}
	}

	sort.Slice(summary.Failed, func(i, j int) bool {
		return summary.Failed[i].ClientID < summary.Failed[j].ClientID
	})

	if uc.ActionNotices != nil {
		_, _ = uc.ActionNotices.Execute(ctx, p)
	}

	return summary, nil
}
