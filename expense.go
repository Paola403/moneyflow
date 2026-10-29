// Package moneyflow é uma versão ENXUTA do MoneyFlow, feita só para demonstrar
// a biblioteca testify: sem SQL, sem HTTP, apenas regras de negócio simples.
package moneyflow

import (
	"errors"
	"fmt"
	"strings"
)

// Erros sentinela: quem chama usa errors.Is para identificá-los.
var (
	ErrInvalidAmount    = errors.New("o valor deve ser maior que zero")
	ErrEmptyDescription = errors.New("a descrição não pode ser vazia")
	ErrBudgetExceeded   = errors.New("a despesa ultrapassa o orçamento disponível")
)

// BudgetExceededError carrega os números da rejeição e, ao mesmo tempo,
// "é" um ErrBudgetExceeded (veja Unwrap). Serve para demonstrar errors.As.
type BudgetExceededError struct {
	RequestedCents int64
	AvailableCents int64
}

func (e *BudgetExceededError) Error() string {
	return fmt.Sprintf("%s (solicitado %d, disponível %d centavos)",
		ErrBudgetExceeded, e.RequestedCents, e.AvailableCents)
}

func (e *BudgetExceededError) Unwrap() error { return ErrBudgetExceeded }

// Expense é uma despesa. O valor fica em CENTAVOS (inteiro) para evitar
// erros de ponto flutuante: R$ 19,90 é 1990.
type Expense struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	Category    string `json:"category"`
}

// Validate devolve nil se a despesa é válida, ou todos os problemas juntos.
// errors.Join devolve nil quando não há erros e permite errors.Is em cada um.
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

// NewExpense normaliza os campos (sem espaços nas pontas, categoria em
// minúsculas) e valida o resultado.
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

// MustNewExpense é como NewExpense, mas entra em pânico se algo for inválido.
// Útil em testes e em valores fixos; serve para demonstrar assert.Panics.
func MustNewExpense(description string, amountCents int64, category string) Expense {
	e, err := NewExpense(description, amountCents, category)
	if err != nil {
		panic(err)
	}
	return e
}
