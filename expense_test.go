package moneyflow_test

// Testes unitários das regras de validação e da estrutura Expense.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"moneyflow"
)

// require interrompe o teste quando uma condição essencial falha; assert
// registra a falha e permite executar as verificações seguintes.
func TestNewExpense_NormalizesFields(t *testing.T) {
	got, err := moneyflow.NewExpense("  Café ", 850, "  Alimentacao ")

	require.NoError(t, err) // sem sucesso, olhar `got` não faria sentido

	// As verificações independentes são executadas para reportar possíveis falhas em conjunto.
	assert.Equal(t, "Café", got.Description, "espaços das pontas devem sair") // mensagem opcional
	assert.Equal(t, "alimentacao", got.Category)
	assert.Equal(t, int64(850), got.AmountCents)
	assert.Empty(t, got.ID, "o ID só é gerado pelo service")
}

// Teste orientado a tabela: cada cenário é executado como um subteste independente.
func TestExpense_Validate(t *testing.T) {
	tests := []struct {
		name     string
		expense  moneyflow.Expense
		wantErrs []error // vazio = deve ser válida
	}{
		{"válida", moneyflow.Expense{Description: "Almoço", AmountCents: 1990}, nil},
		{"valor zero", moneyflow.Expense{Description: "Almoço", AmountCents: 0}, []error{moneyflow.ErrInvalidAmount}},
		{"valor negativo", moneyflow.Expense{Description: "Almoço", AmountCents: -5}, []error{moneyflow.ErrInvalidAmount}},
		{"descrição só espaços", moneyflow.Expense{Description: "   ", AmountCents: 100}, []error{moneyflow.ErrEmptyDescription}},
		{"dois problemas juntos", moneyflow.Expense{}, []error{moneyflow.ErrEmptyDescription, moneyflow.ErrInvalidAmount}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.expense.Validate()

			if len(tc.wantErrs) == 0 {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, want := range tc.wantErrs {
				assert.ErrorIs(t, err, want) // verifica a identidade do erro na cadeia
			}
		})
	}
}

func TestExpense_Validate_DoesNotReportUnrelatedErrors(t *testing.T) {
	err := moneyflow.Expense{Description: "Almoço", AmountCents: 0}.Validate()

	require.Error(t, err)
	assert.ErrorIs(t, err, moneyflow.ErrInvalidAmount)
	assert.NotErrorIs(t, err, moneyflow.ErrEmptyDescription)
	assert.ErrorContains(t, err, "maior que zero") // valida parte da mensagem retornada
}

// Cria instâncias dos helpers para reutilizar as verificações sem passar t em cada chamada.
func TestExpense_Validate_UsingAssertNew(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	err := moneyflow.Expense{}.Validate()

	r.Error(err)
	a.ErrorIs(err, moneyflow.ErrInvalidAmount)
	a.ErrorIs(err, moneyflow.ErrEmptyDescription)
}

// Verifica o comportamento de MustNewExpense para entradas válidas e inválidas.
func TestMustNewExpense(t *testing.T) {
	assert.NotPanics(t, func() { moneyflow.MustNewExpense("Almoço", 1990, "alimentacao") })
	assert.Panics(t, func() { moneyflow.MustNewExpense("", -1, "x") })
}

// Compara a representação JSON sem depender da ordem das propriedades ou da formatação.
func TestExpense_JSON(t *testing.T) {
	b, err := json.Marshal(moneyflow.Expense{ID: "exp-1", Description: "Almoço", AmountCents: 1990, Category: "alimentacao"})

	require.NoError(t, err)
	assert.JSONEq(t, `{"category":"alimentacao","amount_cents":1990,"description":"Almoço","id":"exp-1"}`, string(b))
}

// Valida o conteúdo retornado pelo repositório e as operações sobre a coleção.
func TestMemoryRepository_SaveAndList(t *testing.T) {
	repo := moneyflow.NewMemoryRepository()
	ctx := context.Background()
	a := moneyflow.Expense{ID: "exp-1", Description: "A", AmountCents: 100}
	b := moneyflow.Expense{ID: "exp-2", Description: "B", AmountCents: 200}

	empty, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, empty)

	require.NoError(t, repo.Save(ctx, a))
	require.NoError(t, repo.Save(ctx, b))

	got, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Contains(t, got, a)
	assert.NotContains(t, got, moneyflow.Expense{ID: "exp-3"})
	assert.ElementsMatch(t, []moneyflow.Expense{b, a}, got) // mesma lista, qualquer ordem
}
