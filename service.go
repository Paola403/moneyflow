package moneyflow

import (
	"context"
	"fmt"
	"sync"
)

// Notifier abstrai o envio de notificações, permitindo trocar o mecanismo
// de comunicação sem alterar as regras do Service.
type Notifier interface {
	Notify(message string) error
}

// Service coordena a criação de despesas, valida o orçamento e aciona
// as dependências de persistência e notificação.
type Service struct {
	repo     Repository
	notifier Notifier // dependência opcional para envio de notificações
	limit    int64    // limite total do orçamento, em centavos

	mu     sync.Mutex // sincroniza a consulta de saldo e a gravação da despesa
	lastID int
}

func NewService(repo Repository, notifier Notifier, limitCents int64) *Service {
	return &Service{repo: repo, notifier: notifier, limit: limitCents}
}

// CreateExpense valida os dados, verifica o saldo e persiste a despesa.
// A gravação só ocorre quando todas as regras de negócio são atendidas.
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
	if expense.AmountCents > available { // permite consumir o saldo exato
		return Expense{}, &BudgetExceededError{
			RequestedCents: expense.AmountCents,
			AvailableCents: available,
		}
	}

	expense.ID = fmt.Sprintf("exp-%d", s.lastID+1)
	if err := s.repo.Save(ctx, expense); err != nil {
		return Expense{}, fmt.Errorf("salvar despesa: %w", err) // mantém o erro original acessível via errors.Is
	}
	s.lastID++ // atualiza o identificador somente após a gravação

	if s.notifier != nil && spent+expense.AmountCents == s.limit {
		go s.notifier.Notify("Orçamento esgotado!") // envia a notificação sem bloquear a operação principal
	}
	return expense, nil
}

// Balance retorna o valor disponível do orçamento, em centavos.
func (s *Service) Balance(ctx context.Context) (int64, error) {
	spent, err := s.spent(ctx)
	if err != nil {
		return 0, err
	}
	return s.limit - spent, nil
}

// UsedPercent calcula a porcentagem do orçamento que já foi utilizada.
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
