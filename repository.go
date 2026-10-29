package moneyflow

import (
	"context"
	"slices"
	"sync"
)

// Repository é o contrato de armazenamento. O Service depende só desta
// interface, então nos testes podemos trocá-la por um MOCK.
type Repository interface {
	Save(ctx context.Context, e Expense) error
	List(ctx context.Context) ([]Expense, error)
}

// MemoryRepository guarda as despesas em memória. Em testes é um "fake":
// uma implementação simplificada, mas que funciona de verdade.
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
