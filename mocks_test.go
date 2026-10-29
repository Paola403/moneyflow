package moneyflow_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"moneyflow"
)

// Os mocks apenas REGISTRAM a chamada (m.Called) e devolvem o que o teste
// programou com .On(...).Return(...). Nenhuma regra de negócio mora aqui.

type MockRepository struct{ mock.Mock }

func (m *MockRepository) Save(ctx context.Context, e moneyflow.Expense) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]moneyflow.Expense, error) {
	args := m.Called(ctx)
	list, _ := args.Get(0).([]moneyflow.Expense) // comma-ok: Return(nil, err) não causa panic
	return list, args.Error(1)
}

type MockNotifier struct{ mock.Mock }

func (m *MockNotifier) Notify(message string) error {
	return m.Called(message).Error(0)
}

// Se a interface mudar e o mock ficar para trás, o projeto deixa de compilar.
var (
	_ moneyflow.Repository = (*MockRepository)(nil)
	_ moneyflow.Notifier   = (*MockNotifier)(nil)
)
