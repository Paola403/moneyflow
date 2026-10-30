package moneyflow

import (
	"context"
	"slices"
	"sync"
)

// Repository define as operações de persistência utilizadas pelo Service.
// A interface permite substituir a implementação real por doubles nos testes.
type Repository interface {
	Save(ctx context.Context, e Expense) error
	List(ctx context.Context) ([]Expense, error)
}

// MemoryRepository implementa Repository usando armazenamento em memória.
// Nos testes, funciona como fake: mantém estado e executa as operações solicitadas.
type MemoryRepository struct {
	mu       sync.Mutex
	expenses []Expense
}

var _ Repository = (*MemoryRepository)(nil)

func NewMemoryRepository() *MemoryRepository { return &MemoryRepository{} }

func (r *MemoryRepository) Save(_ context.Context, e Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expenses = append(r.expenses, e)
	return nil
}

func (r *MemoryRepository) List(_ context.Context) ([]Expense, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.expenses), nil
}
