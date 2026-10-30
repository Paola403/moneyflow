package moneyflow_test

// Testes organizados com testify/suite. A estrutura compartilha dependências
// e disponibiliza métodos de preparação e finalização ao longo do ciclo de vida.
// SetupSuite/TearDownSuite executam uma vez; SetupTest/TearDownTest executam
// para cada teste; SetupSubTest prepara cada cenário executado com s.Run.

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"moneyflow"
)

const suiteLimit = 10_000 // R$ 100,00

// O repositório em memória atua como fake; o notificador permanece mockado
// para permitir validar o envio assíncrono de notificações.
type ServiceSuite struct {
	suite.Suite   // embute os asserts: s.Equal, s.NoError, s.ErrorIs...
	ctx           context.Context
	repo          *moneyflow.MemoryRepository
	notifier      *MockNotifier
	notifications *atomic.Int32
	svc           *moneyflow.Service
}

func (s *ServiceSuite) SetupSuite() {
	s.ctx = context.Background()
	s.T().Log("SetupSuite: roda uma vez, antes de todos os testes")
}

func (s *ServiceSuite) TearDownSuite() {
	s.T().Log("TearDownSuite: roda uma vez, depois de todos os testes")
}

// Reinicializa as dependências para manter os cenários isolados entre si.
func (s *ServiceSuite) SetupTest()    { s.reset() }
func (s *ServiceSuite) SetupSubTest() { s.reset() }

func (s *ServiceSuite) TearDownTest() {
	s.notifier.AssertExpectations(s.T())
}

func (s *ServiceSuite) reset() {
	s.repo = moneyflow.NewMemoryRepository()
	s.notifier = new(MockNotifier)
	s.notifications = new(atomic.Int32)
	n := s.notifications
	s.notifier.On("Notify", mock.Anything).
		Run(func(mock.Arguments) { n.Add(1) }).
		Return(nil).
		Maybe()
	s.svc = moneyflow.NewService(s.repo, s.notifier, suiteLimit)
}

// Require interrompe o teste em falhas essenciais; os demais asserts permitem continuar.
func (s *ServiceSuite) TestCreateExpense_UpdatesBalanceAndHistory() {
	_, err := s.svc.CreateExpense(s.ctx, "Almoço", 1_990, "alimentacao")
	s.Require().NoError(err)
	_, err = s.svc.CreateExpense(s.ctx, "Ônibus", 500, "transporte")
	s.Require().NoError(err)

	balance, err := s.svc.Balance(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(7_510), balance)

	history, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Len(history, 2)
}

func (s *ServiceSuite) TestRejectedExpense_DoesNotChangeBalanceOrHistory() {
	_, err := s.svc.CreateExpense(s.ctx, "Notebook", suiteLimit+1, "tecnologia")
	s.ErrorIs(err, moneyflow.ErrBudgetExceeded)

	balance, err := s.svc.Balance(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(suiteLimit), balance)

	history, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Empty(history)
}

// Compara o percentual com uma tolerância para considerar a precisão de float.
func (s *ServiceSuite) TestUsedPercent() {
	_, err := s.svc.CreateExpense(s.ctx, "Mercado", 3_333, "alimentacao")
	s.Require().NoError(err)

	percent, err := s.svc.UsedPercent(s.ctx)

	s.Require().NoError(err)
	s.InDelta(33.33, percent, 0.001)
}

// Confirma que consumir todo o saldo é permitido e aciona a notificação.
func (s *ServiceSuite) TestExpenseEqualToBalance_IsAcceptedAndNotifies() {
	_, err := s.svc.CreateExpense(s.ctx, "Tudo", suiteLimit, "geral")
	s.Require().NoError(err)

	balance, err := s.svc.Balance(s.ctx)
	s.Require().NoError(err)
	s.Zero(balance)

	s.Eventually(func() bool { return s.notifications.Load() == 1 },
		time.Second, 10*time.Millisecond, "o aviso deveria ter sido enviado")
}

// Executa diferentes entradas isoladamente; SetupSubTest prepara um estado novo para cada caso.
func (s *ServiceSuite) TestCreateExpense_Cases() {
	cases := []struct {
		name        string
		description string
		amount      int64
		wantErr     error // nil = deve ser aceita
	}{
		{"válida", "Almoço", 1_990, nil},
		{"valor zero", "Almoço", 0, moneyflow.ErrInvalidAmount},
		{"descrição vazia", "", 100, moneyflow.ErrEmptyDescription},
		{"estoura por 1 centavo", "TV", suiteLimit + 1, moneyflow.ErrBudgetExceeded},
		{"gasta exatamente tudo", "Tudo", suiteLimit, nil},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.svc.CreateExpense(s.ctx, tc.description, tc.amount, "geral")

			if tc.wantErr == nil {
				s.NoError(err)
				return
			}
			s.ErrorIs(err, tc.wantErr)
		})
	}
}

// Integra a suite ao runner padrão de testes do Go.
func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
