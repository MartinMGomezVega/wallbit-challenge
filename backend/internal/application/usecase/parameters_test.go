package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
)

type fakeParamsRepo struct {
	params parameters.AccountantParameters
	err    error
}

func (f *fakeParamsRepo) Get(_ context.Context) (parameters.AccountantParameters, error) {
	if f.err != nil {
		return parameters.AccountantParameters{}, f.err
	}
	return f.params, nil
}

func (f *fakeParamsRepo) Save(_ context.Context, p parameters.AccountantParameters) error {
	if f.err != nil {
		return f.err
	}
	f.params = p
	return nil
}

func TestParametersUsecases(t *testing.T) {
	ctx := context.Background()
	initial := parameters.Defaults()
	repo := &fakeParamsRepo{params: initial}
	clk := fakes.Clock{Current: time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)}

	getUC := &GetParameters{Repo: repo}
	updateUC := &UpdateParameters{Repo: repo, Clock: clk}

	// 1. Get
	got, err := getUC.Execute(ctx)
	if err != nil || got.CategoryCapWarningBp != initial.CategoryCapWarningBp {
		t.Fatalf("GetParameters failed: %v, got %+v", err, got)
	}

	// 2. Update success
	modified := initial
	modified.CategoryCapWarningBp = 8500
	updated, err := updateUC.Execute(ctx, modified)
	if err != nil {
		t.Fatalf("UpdateParameters failed: %v", err)
	}
	if updated.Parameters.CategoryCapWarningBp != 8500 || updated.Parameters.UpdatedBy != "accountant" || !updated.Parameters.UpdatedAt.Equal(clk.Current) {
		t.Fatalf("unexpected updated params: %+v", updated)
	}

	// 3. Update validation error
	invalid := initial
	invalid.CategoryCapWarningBp = -10
	_, err = updateUC.Execute(ctx, invalid)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}

	// 4. Repo save error
	repo.err = errors.New("db error")
	_, err = updateUC.Execute(ctx, modified)
	if err == nil {
		t.Fatal("expected save error, got nil")
	}

	// 5. Repo get error
	_, err = getUC.Execute(ctx)
	if err == nil {
		t.Fatal("expected get error, got nil")
	}
}
