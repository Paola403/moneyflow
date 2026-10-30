package moneyflow_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"moneyflow"
)

// Os mocks registram as chamadas recebidas e retornam os valores configurados
// pelo teste. Eles isolam o Service das implementações concretas.

type MockRepository struct{ mock.Mock }

func (m *MockRepository) Save(ctx context.Context, e moneyflow.Expense) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]moneyflow.Expense, error) {
	args := m.Called(ctx)
	list, _ := args.Get(0).([]moneyflow.Expense) // evita panic quando o retorno configurado é nil
	return list, args.Error(1)
}

type MockNotifier struct{ mock.Mock }

func (m *MockNotifier) Notify(message string) error {
	return m.Called(message).Error(0)
}

// Verifica em tempo de compilação se os mocks implementam as interfaces esperadas.
var (
	_ moneyflow.Repository = (*MockRepository)(nil)
	_ moneyflow.Notifier   = (*MockNotifier)(nil)
)
