package moneyflow_test

// CONCEITO 1: testify/assert e testify/require (sem mocks, sem suite).

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"moneyflow"
)

// require = pré-condição (se falhar, PARA o teste).
// assert  = verificação do resultado (se falhar, marca e CONTINUA).
func TestNewExpense_NormalizesFields(t *testing.T) {
	got, err := moneyflow.NewExpense("  Café ", 850, "  Alimentacao ")

	require.NoError(t, err) // sem sucesso, olhar `got` não faria sentido

	// Vários asserts: se mais de um falhar, você vê TODAS as falhas de uma vez.
	assert.Equal(t, "Café", got.Description, "espaços das pontas devem sair") // mensagem opcional
	assert.Equal(t, "alimentacao", got.Category)
	assert.Equal(t, int64(850), got.AmountCents)
	assert.Empty(t, got.ID, "o ID só é gerado pelo service")
}

// Teste de tabela + subtestes (t.Run): cada caso vira uma linha de dados.
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
				assert.ErrorIs(t, err, want) // compara o erro, não o texto
			}
		})
	}
}

func TestExpense_Validate_DoesNotReportUnrelatedErrors(t *testing.T) {
	err := moneyflow.Expense{Description: "Almoço", AmountCents: 0}.Validate()

	require.Error(t, err)
	assert.ErrorIs(t, err, moneyflow.ErrInvalidAmount)
	assert.NotErrorIs(t, err, moneyflow.ErrEmptyDescription)
	assert.ErrorContains(t, err, "maior que zero") // checa um trecho da mensagem
}

// assert.New(t) / require.New(t): o mesmo, sem repetir `t` em cada linha.
func TestExpense_Validate_UsingAssertNew(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	err := moneyflow.Expense{}.Validate()

	r.Error(err)
	a.ErrorIs(err, moneyflow.ErrInvalidAmount)
	a.ErrorIs(err, moneyflow.ErrEmptyDescription)
}

// Panics / NotPanics: verificam se uma função entra (ou não) em pânico.
func TestMustNewExpense(t *testing.T) {
	assert.NotPanics(t, func() { moneyflow.MustNewExpense("Almoço", 1990, "alimentacao") })
	assert.Panics(t, func() { moneyflow.MustNewExpense("", -1, "x") })
}

// JSONEq: compara JSON ignorando espaços e a ordem das chaves.
func TestExpense_JSON(t *testing.T) {
	b, err := json.Marshal(moneyflow.Expense{ID: "exp-1", Description: "Almoço", AmountCents: 1990, Category: "alimentacao"})

	require.NoError(t, err)
	assert.JSONEq(t, `{"category":"alimentacao","amount_cents":1990,"description":"Almoço","id":"exp-1"}`, string(b))
}

// Asserts de coleções: Len, Contains, NotContains, ElementsMatch, Empty.
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
