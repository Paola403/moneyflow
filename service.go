package moneyflow

import (
	"context"
	"fmt"
	"sync"
)

// Notifier avisa o usuário de algo (e-mail, push...). É outra dependência
// que os testes substituem por um mock.
type Notifier interface {
	Notify(message string) error
}

// Service aplica as regras de negócio: valida a despesa, confere o orçamento,
// salva e, se o orçamento acabar, avisa em segundo plano (goroutine).
type Service struct {
	repo     Repository
	notifier Notifier // opcional: nil = sem avisos
	limit    int64    // orçamento total, em centavos

	mu     sync.Mutex // torna "conferir saldo + salvar" uma operação só
	lastID int
}

func NewService(repo Repository, notifier Notifier, limitCents int64) *Service {
	return &Service{repo: repo, notifier: notifier, limit: limitCents}
}

// CreateExpense valida, confere o orçamento e persiste a despesa.
// Despesa inválida ou acima do saldo NUNCA chega ao Save.
func (s *Service) CreateExpense(ctx context.Context, description string, amountCents int64, category string) (Expense, error) {
	expense, err := NewExpense(description, amountCents, category)
	if err != nil {
		return Expense{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	spent, err := s.spent(ctx)
	if err != nil {
		return Expense{}, err
	}

	available := s.limit - spent
	if expense.AmountCents > available { // igualdade é aceita: gastar tudo pode
		return Expense{}, &BudgetExceededError{
			RequestedCents: expense.AmountCents,
			AvailableCents: available,
		}
	}

	expense.ID = fmt.Sprintf("exp-%d", s.lastID+1)
	if err := s.repo.Save(ctx, expense); err != nil {
		return Expense{}, fmt.Errorf("salvar despesa: %w", err) // %w preserva errors.Is
	}
	s.lastID++ // só despesas realmente salvas consomem um ID

	if s.notifier != nil && spent+expense.AmountCents == s.limit {
		go s.notifier.Notify("Orçamento esgotado!") // assíncrono; erro de aviso é ignorado
	}
	return expense, nil
}

// Balance devolve quanto ainda pode ser gasto, em centavos.
func (s *Service) Balance(ctx context.Context) (int64, error) {
	spent, err := s.spent(ctx)
	if err != nil {
		return 0, err
	}
	return s.limit - spent, nil
}

// UsedPercent devolve a porcentagem do orçamento já gasta (ex.: 33.33).
func (s *Service) UsedPercent(ctx context.Context) (float64, error) {
	spent, err := s.spent(ctx)
	if err != nil {
		return 0, err
	}
	if s.limit <= 0 {
		return 0, nil
	}
	return float64(spent) * 100 / float64(s.limit), nil
}

func (s *Service) spent(ctx context.Context) (int64, error) {
	expenses, err := s.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("listar despesas: %w", err)
	}
	var total int64
	for _, e := range expenses {
		total += e.AmountCents
	}
	return total, nil
}
