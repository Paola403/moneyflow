package moneyflow_test

// CONCEITO 3: testify/suite. Agrupa testes em uma struct, com ciclo de vida
// (Setup/TearDown) e com os asserts disponíveis direto em `s`.
//
// Ordem de execução:
//   SetupSuite            (1x, antes de tudo)
//     SetupTest           (antes de CADA TestXxx)
//       SetupSubTest      (antes de cada s.Run)
//         ... teste ...
//       TearDownSubTest
//     TearDownTest        (depois de CADA TestXxx)
//   TearDownSuite         (1x, no fim)

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

// Aqui o repositório é o FAKE em memória (resultado de várias operações
// encadeadas); só o notificador é mock.
type ServiceSuite struct {
	suite.Suite // embute os asserts: s.Equal, s.NoError, s.ErrorIs...
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

// Estado novo antes de cada teste e de cada subteste: nada vaza entre eles.
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

// s.Require() = versão que PARA o teste; s.Equal etc. = versão que continua.
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

// InDelta: compara números decimais com tolerância (nunca use Equal com float).
func (s *ServiceSuite) TestUsedPercent() {
	_, err := s.svc.CreateExpense(s.ctx, "Mercado", 3_333, "alimentacao")
	s.Require().NoError(err)

	percent, err := s.svc.UsedPercent(s.ctx)

	s.Require().NoError(err)
	s.InDelta(33.33, percent, 0.001)
}

// Gastar exatamente o saldo é aceito e dispara o aviso (assíncrono).
func (s *ServiceSuite) TestExpenseEqualToBalance_IsAcceptedAndNotifies() {
	_, err := s.svc.CreateExpense(s.ctx, "Tudo", suiteLimit, "geral")
	s.Require().NoError(err)

	balance, err := s.svc.Balance(s.ctx)
	s.Require().NoError(err)
	s.Zero(balance)

	s.Eventually(func() bool { return s.notifications.Load() == 1 },
		time.Second, 10*time.Millisecond, "o aviso deveria ter sido enviado")
}

// Teste de tabela com s.Run: o SetupSubTest zera o estado a cada caso.
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

// Ponto de entrada: o `go test` só enxerga funções Test*(t *testing.T).
func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
