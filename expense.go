// Package moneyflow concentra as regras de negócio do controle de despesas.
// A implementação usa dependências em memória para manter o exemplo independente de infraestrutura.
package moneyflow

import (
	"errors"
	"fmt"
	"strings"
)

// Erros sentinela usados para identificar falhas específicas com errors.Is.
var (
	ErrInvalidAmount    = errors.New("o valor deve ser maior que zero")
	ErrEmptyDescription = errors.New("a descrição não pode ser vazia")
	ErrBudgetExceeded   = errors.New("a despesa ultrapassa o orçamento disponível")
)

// BudgetExceededError detalha a tentativa de gasto acima do saldo disponível.
// Unwrap mantém a compatibilidade com errors.Is e permite obter os dados via errors.As.
type BudgetExceededError struct {
	RequestedCents int64
	AvailableCents int64
}

func (e *BudgetExceededError) Error() string {
	return fmt.Sprintf("%s (solicitado %d, disponível %d centavos)",
		ErrBudgetExceeded, e.RequestedCents, e.AvailableCents)
}

func (e *BudgetExceededError) Unwrap() error { return ErrBudgetExceeded }

// Expense representa uma despesa. O valor é armazenado em centavos para
// evitar imprecisões de ponto flutuante em operações monetárias.
type Expense struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	Category    string `json:"category"`
}

// Validate verifica os campos obrigatórios e reúne todas as inconsistências
// encontradas em um único erro.
func (e Expense) Validate() error {
	var errs []error
	if strings.TrimSpace(e.Description) == "" {
		errs = append(errs, ErrEmptyDescription)
	}
	if e.AmountCents <= 0 {
		errs = append(errs, ErrInvalidAmount)
	}
	return errors.Join(errs...)
}

// NewExpense normaliza os dados de entrada e valida a despesa antes de retorná-la.
func NewExpense(description string, amountCents int64, category string) (Expense, error) {
	e := Expense{
		Description: strings.TrimSpace(description),
		AmountCents: amountCents,
		Category:    strings.ToLower(strings.TrimSpace(category)),
	}
	if err := e.Validate(); err != nil {
		return Expense{}, err
	}
	return e, nil
}

// MustNewExpense cria uma despesa válida ou provoca panic quando os dados são inválidos.
func MustNewExpense(description string, amountCents int64, category string) Expense {
	e, err := NewExpense(description, amountCents, category)
	if err != nil {
		panic(err)
	}
	return e
}
