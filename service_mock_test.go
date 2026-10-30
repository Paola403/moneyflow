package moneyflow_test

// Testes do Service com dependências simuladas por testify/mock.
// Os mocks permitem controlar retornos e verificar as interações entre componentes.

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"moneyflow"
)

// Erro utilizado para simular uma falha na camada de persistência.
var errDB = errors.New("db: connection refused")

type fixture struct {
	svc      *moneyflow.Service
	repo     *MockRepository
	notifier *MockNotifier
}

// newFixture configura o Service com mocks e valida as expectativas ao final
// do teste, garantindo que as interações configuradas foram executadas.
func newFixture(t *testing.T, limit int64) *fixture {
	t.Helper()
	repo, notifier := new(MockRepository), new(MockNotifier)
	t.Cleanup(func() {
		repo.AssertExpectations(t)
		notifier.AssertExpectations(t)
	})
	return &fixture{svc: moneyflow.NewService(repo, notifier, limit), repo: repo, notifier: notifier}
}

// Fluxo do teste: configurar as respostas esperadas, executar a operação e validar o resultado.
func TestCreateExpense_Success(t *testing.T) {
	f := newFixture(t, 10_000)
	ctx := context.Background()
	history := []moneyflow.Expense{{ID: "old", Description: "antiga", AmountCents: 3_000}}
	want := moneyflow.Expense{ID: "exp-1", Description: "Almoço", AmountCents: 1_990, Category: "alimentacao"}

	// Configura as chamadas esperadas e os respectivos retornos.
	f.repo.On("List", ctx).Return(history, nil).Once()
	f.repo.On("Save", ctx, want).Return(nil).Once()

	// Executa a regra de negócio sob teste.
	got, err := f.svc.CreateExpense(ctx, "Almoço", 1_990, "alimentacao")

	// Confere o resultado e as interações realizadas.
	require.NoError(t, err)
	assert.Equal(t, want, got)
	f.repo.AssertNumberOfCalls(t, "Save", 1)
	f.repo.AssertCalled(t, "Save", ctx, want)
}

// Usa um matcher para validar propriedades específicas do objeto enviado ao repositório.
func TestCreateExpense_NormalizesInputBeforeSaving(t *testing.T) {
	f := newFixture(t, 10_000)
	f.repo.On("List", mock.Anything).Return(nil, nil)
	f.repo.On("Save", mock.Anything, mock.MatchedBy(func(e moneyflow.Expense) bool {
		return e.Description == "Café" && e.Category == "alimentacao"
	})).Return(nil)

	_, err := f.svc.CreateExpense(context.Background(), "  Café ", 850, "  Alimentacao ")

	require.NoError(t, err)
}

// Garante que entradas inválidas são rejeitadas antes de acessar a persistência.
func TestCreateExpense_InvalidInput_DoesNotTouchRepository(t *testing.T) {
	tests := []struct {
		name        string
		description string
		amount      int64
		wantErr     error
	}{
		{"valor negativo", "x", -1, moneyflow.ErrInvalidAmount},
		{"valor zero", "x", 0, moneyflow.ErrInvalidAmount},
		{"descrição vazia", "", 100, moneyflow.ErrEmptyDescription},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t, 10_000) // sem expectativas configuradas, chamadas inesperadas são reportadas pelo mock

			got, err := f.svc.CreateExpense(context.Background(), tc.description, tc.amount, "geral")

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Zero(t, got)
			f.repo.AssertNotCalled(t, "List", mock.Anything)
			f.repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
		})
	}
}

// Valida a rejeição por orçamento excedido e confirma que nenhuma gravação foi realizada.
func TestCreateExpense_BudgetExceeded(t *testing.T) {
	f := newFixture(t, 10_000)
	f.repo.On("List", mock.Anything).Return([]moneyflow.Expense{{AmountCents: 9_500}}, nil)

	_, err := f.svc.CreateExpense(context.Background(), "TV", 501, "casa")

	require.Error(t, err)
	assert.ErrorIs(t, err, moneyflow.ErrBudgetExceeded)

	var be *moneyflow.BudgetExceededError
	require.ErrorAs(t, err, &be) // garante que be foi preenchido antes da comparação
	assert.Equal(t, &moneyflow.BudgetExceededError{RequestedCents: 501, AvailableCents: 500}, be)

	f.repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// Simula falhas de consulta e gravação, verificando a propagação do erro original.
func TestCreateExpense_RepositoryFailures(t *testing.T) {
	t.Run("falha ao consultar", func(t *testing.T) {
		f := newFixture(t, 10_000)
		f.repo.On("List", mock.Anything).Return(nil, errDB)

		_, err := f.svc.CreateExpense(context.Background(), "Almoço", 1_990, "alimentacao")

		assert.ErrorIs(t, err, errDB)
		f.repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	})

	t.Run("falha ao salvar", func(t *testing.T) {
		f := newFixture(t, 10_000)
		f.repo.On("List", mock.Anything).Return(nil, nil)
		f.repo.On("Save", mock.Anything, mock.Anything).Return(errDB)

		_, err := f.svc.CreateExpense(context.Background(), "Almoço", 1_990, "alimentacao")

		assert.ErrorIs(t, err, errDB)
		assert.ErrorContains(t, err, "salvar despesa")
	})
}

// Verifica a geração sequencial de identificadores e a quantidade de chamadas ao repositório.
func TestCreateExpense_TwoExpenses_GetSequentialIDs(t *testing.T) {
	f := newFixture(t, 10_000)
	f.repo.On("List", mock.Anything).Return(nil, nil).Times(2)
	f.repo.On("Save", mock.Anything, mock.AnythingOfType("moneyflow.Expense")).Return(nil).Times(2)

	first, err := f.svc.CreateExpense(context.Background(), "A", 100, "x")
	require.NoError(t, err)
	second, err := f.svc.CreateExpense(context.Background(), "B", 100, "x")
	require.NoError(t, err)

	assert.Equal(t, "exp-1", first.ID)
	assert.Equal(t, "exp-2", second.ID)
}

// O callback Run registra a chamada do notificador. Eventually aguarda a execução
// da goroutine dentro de um intervalo definido.
func TestCreateExpense_NotifiesWhenBudgetIsExhausted(t *testing.T) {
	f := newFixture(t, 10_000)
	f.repo.On("List", mock.Anything).Return([]moneyflow.Expense{{AmountCents: 9_000}}, nil)
	f.repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	var notified atomic.Int32
	f.notifier.On("Notify", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) { notified.Add(1) }).
		Return(nil).
		Once()

	_, err := f.svc.CreateExpense(context.Background(), "Aluguel", 1_000, "casa") // consome o saldo restante

	require.NoError(t, err)
	assert.Eventually(t, func() bool { return notified.Load() == 1 },
		time.Second, 10*time.Millisecond, "o aviso deveria ter sido enviado")
}

// Maybe torna a chamada opcional. Never verifica que a notificação não ocorre
// enquanto ainda existe saldo disponível.
func TestCreateExpense_DoesNotNotifyWhileThereIsBalance(t *testing.T) {
	f := newFixture(t, 10_000)
	f.repo.On("List", mock.Anything).Return(nil, nil)
	f.repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	var notified atomic.Int32
	f.notifier.On("Notify", mock.Anything).
		Run(func(args mock.Arguments) { notified.Add(1) }).
		Return(nil).
		Maybe()

	_, err := f.svc.CreateExpense(context.Background(), "Café", 500, "alimentacao")

	require.NoError(t, err)
	assert.Never(t, func() bool { return notified.Load() > 0 },
		100*time.Millisecond, 10*time.Millisecond, "não deveria avisar com saldo sobrando")
}
