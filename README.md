# MoneyFlow — Demonstração da Biblioteca Testify

## 1. Sobre o projeto

O **MoneyFlow** é uma aplicação desenvolvida em Go com o objetivo de demonstrar, na prática, a utilização da biblioteca [Testify](https://github.com/stretchr/testify) para a criação de testes automatizados.

A aplicação permite registrar e acompanhar despesas dentro de um orçamento definido, realizando validações, verificando o saldo disponível e emitindo notificações quando o orçamento é totalmente utilizado.

O projeto possui uma implementação simplificada, sem utilização de banco de dados ou comunicação HTTP. As informações são armazenadas em memória.

## 2. O que é o Testify?

O **Testify** é uma biblioteca para a linguagem Go que oferece recursos para facilitar a criação, organização e execução de testes automatizados.

Seu principal objetivo é tornar os testes mais simples de escrever, compreender e manter.

Por exemplo, sem utilizar o Testify, podemos realizar uma comparação da seguinte maneira:

```go
if got != want {
    t.Errorf("esperado %v, obtido %v", want, got)
}
```

Utilizando o Testify, podemos realizar a mesma verificação de maneira mais simples:

```go
assert.Equal(t, want, got)
```

A biblioteca disponibiliza quatro recursos principais, todos demonstrados neste projeto.

| Recurso | Finalidade | Arquivos |
|---|---|---|
| `assert` | Realiza verificações e permite que o teste continue mesmo quando uma delas falha. | `expense_test.go` |
| `require` | Realiza verificações e interrompe imediatamente o teste quando uma delas falha. | `expense_test.go` |
| `mock` | Permite criar objetos simulados para representar dependências e controlar suas respostas durante os testes. | `mocks_test.go` e `service_mock_test.go` |
| `suite` | Permite agrupar testes em uma estrutura, compartilhando configurações e organizando sua execução. | `service_suite_test.go` |

## 3. Funcionalidades da aplicação

O MoneyFlow possui as seguintes funcionalidades:

- Registro de despesas.
- Validação da descrição, categoria e valor das despesas.
- Controle de orçamento disponível.
- Consulta do saldo restante.
- Listagem das despesas registradas.
- Identificação de tentativas de gastos acima do orçamento.
- Notificação quando o orçamento é totalmente utilizado.

### Exemplo de utilização

A aplicação inicia com um orçamento de R$ 100,00.

```text
> add 19,90 alimentacao Almoço
Despesa registrada.

> add 70 casa Aluguel
Despesa registrada.

> add 20 lazer Cinema
Despesa rejeitada: orçamento insuficiente.

> add 10,10 lazer Pipoca
Despesa registrada.
Notificação: orçamento totalmente utilizado!

> saldo

> lista
```

Nesse exemplo, as duas primeiras despesas totalizam R$ 89,90. A tentativa de adicionar R$ 20,00 é rejeitada porque ultrapassaria o orçamento disponível.

Posteriormente, uma despesa de R$ 10,10 é registrada, utilizando exatamente o saldo restante e acionando a notificação.

## 4. Estrutura do projeto

O projeto está organizado da seguinte maneira:

```text
moneyflow/
│
├── cmd/
│   └── demo/
│       └── main.go
│
├── expense.go
├── repository.go
├── service.go
│
├── mocks_test.go
├── expense_test.go
├── service_mock_test.go
├── service_suite_test.go
│
├── go.mod
├── go.sum
└── README.md
```

### Responsabilidade dos arquivos

| Arquivo | Descrição |
|---|---|
| `cmd/demo/main.go` | Contém o programa interativo executado pelo terminal, permitindo experimentar as funcionalidades da aplicação. |
| `expense.go` | Define a estrutura `Expense`, suas validações e os erros relacionados às despesas. |
| `repository.go` | Define a interface `Repository` e sua implementação `MemoryRepository`, responsável pelo armazenamento das despesas em memória. |
| `service.go` | Implementa as regras de negócio, incluindo validações, controle do orçamento, registro de despesas e notificações. |
| `mocks_test.go` | Contém os objetos simulados `MockRepository` e `MockNotifier`, desenvolvidos com o recurso `mock` do Testify. |
| `expense_test.go` | Apresenta testes de validação utilizando `assert` e `require`. |
| `service_mock_test.go` | Demonstra a utilização de Mock para simular dependências e testar diferentes comportamentos do Service. |
| `service_suite_test.go` | Organiza os testes utilizando o recurso Suite, incluindo preparação e finalização dos ambientes de teste. |

## 5. Recursos do Testify utilizados no projeto

### 5.1. Assert e Require

Os recursos `assert` e `require` são utilizados para verificar se os resultados obtidos durante os testes correspondem aos resultados esperados.

A principal diferença entre eles está no comportamento quando uma verificação falha:

- **Assert:** registra a falha, mas permite que as próximas instruções do teste sejam executadas.
- **Require:** registra a falha e interrompe imediatamente a execução do teste atual.

Exemplos utilizados no projeto:

```go
require.NoError(t, err)
assert.Equal(t, want, got)
```

Também são demonstradas as versões que permitem reutilizar o objeto de verificação:

```go
assertions := assert.New(t)
assertions.Equal(want, got)
```

```go
requirements := require.New(t)
requirements.NoError(err)
```

### 5.2. Funções de verificação

O projeto apresenta diferentes funções disponibilizadas pelo Testify para verificar valores, erros, coleções e comportamentos.

**Comparação e valores:**

- `Equal`: verifica se dois valores são iguais.
- `Zero`: verifica se um valor corresponde ao valor zero de seu tipo.
- `Empty`: verifica se um valor está vazio.

**Tratamento de erros:**

- `Error`: verifica se ocorreu um erro.
- `NoError`: verifica se não ocorreu um erro.
- `ErrorIs`: verifica se um erro corresponde a outro erro.
- `NotErrorIs`: verifica se um erro não corresponde a outro.
- `ErrorAs`: verifica se um erro pode ser convertido para determinado tipo.
- `ErrorContains`: verifica se a mensagem de erro contém determinado texto.

**Coleções:**

- `Len`: verifica o tamanho de uma coleção.
- `Contains`: verifica se um elemento está presente.
- `NotContains`: verifica se um elemento não está presente.
- `ElementsMatch`: verifica se duas coleções possuem os mesmos elementos, independentemente da ordem.

**Outras verificações:**

- `JSONEq`: compara estruturas JSON.
- `InDelta`: compara valores numéricos considerando uma margem de diferença.
- `Panics`: verifica se uma função provoca um panic.
- `NotPanics`: verifica se uma função não provoca um panic.

**Verificações relacionadas ao tempo:**

- `Eventually`: verifica se uma condição se torna verdadeira dentro de um intervalo de tempo.
- `Never`: verifica se uma condição permanece falsa durante determinado intervalo.

### 5.3. Mock

O recurso `mock` permite criar objetos simulados que representam dependências utilizadas pela aplicação, como repositórios e serviços externos.

No MoneyFlow, são utilizados dois objetos:

- `MockRepository`: simula o repositório responsável pelo armazenamento e consulta das despesas.
- `MockNotifier`: simula o serviço responsável pelo envio de notificações.

Com os Mocks, é possível controlar as respostas recebidas pelo Service sem depender de implementações reais dessas funcionalidades.

**Principais métodos utilizados:**

| Método | Finalidade |
|---|---|
| `On()` | Define uma chamada esperada ao método simulado. |
| `Return()` | Define os valores que serão retornados pelo Mock. |
| `Once()` | Define que uma chamada deve ocorrer uma única vez. |
| `Times(n)` | Define a quantidade esperada de chamadas. |
| `Maybe()` | Permite que uma chamada seja opcional. |
| `Run()` | Executa uma função quando o método simulado é chamado. |
| `AssertExpectations()` | Verifica se as expectativas configuradas foram atendidas. |
| `AssertCalled()` | Verifica se determinado método foi chamado. |
| `AssertNotCalled()` | Verifica se determinado método não foi chamado. |
| `AssertNumberOfCalls()` | Verifica quantas vezes determinado método foi chamado. |

O Testify também disponibiliza recursos para controlar os argumentos recebidos:

- `mock.Anything`: aceita qualquer argumento.
- `mock.AnythingOfType()`: aceita argumentos de determinado tipo.
- `mock.MatchedBy()`: permite definir condições personalizadas para os argumentos.

Além disso, os Mocks permitem simular situações de sucesso e falha, como erros durante a consulta ou o armazenamento de despesas.

**Exemplo de configuração:**

```go
mockRepo.On("List", ctx).
    Return(history, nil).Once()
```

Nesse exemplo, estamos configurando o Mock para retornar um histórico de despesas quando o método `List` for chamado com o contexto informado.

Dessa maneira, podemos testar as regras de negócio do Service sem precisar utilizar um banco de dados real.

### 5.4. Suite

O recurso `suite` permite agrupar testes relacionados em uma estrutura, facilitando a organização e o compartilhamento de configurações.

No projeto, a estrutura `ServiceSuite` reúne os recursos necessários para testar as operações do serviço de despesas.

Entre os campos utilizados estão o contexto, o repositório em memória, o Mock de notificações e a instância do Service.

**Principais métodos utilizados:**

| Método | Finalidade |
|---|---|
| `SetupSuite()` | Executa uma preparação antes de todos os testes da suíte. |
| `SetupTest()` | Executa uma preparação antes de cada teste. |
| `SetupSubTest()` | Executa uma preparação antes de cada subteste. |
| `TearDownTest()` | Executa uma finalização depois de cada teste. |
| `TearDownSuite()` | Executa uma finalização depois de todos os testes da suíte. |
| `s.Run()` | Executa subtestes dentro da suíte. |

O Suite também permite utilizar diretamente as verificações do Testify por meio de `s.Equal()`, `s.NoError()` e outros métodos.

Para interromper um teste quando uma verificação falha, também é possível utilizar `s.Require()`.

A execução da suíte é iniciada por meio de:

```go
func TestServiceSuite(t *testing.T) {
    suite.Run(t, new(ServiceSuite))
}
```

### 5.5. Diferença entre Mock e Fake

O projeto também demonstra a diferença entre objetos simulados e implementações simplificadas.

**Mock:** permite programar respostas, simular erros e verificar se determinados métodos foram chamados.

**Fake:** possui uma implementação funcional simplificada de uma dependência real.

No MoneyFlow:

- Nos testes de `service_mock_test.go`, o Service utiliza um `MockRepository`, permitindo controlar as respostas do repositório.
- Nos testes de `service_suite_test.go`, o Service utiliza `MemoryRepository`, que realmente armazena e consulta despesas em memória.

Essa abordagem permite testar tanto situações controladas individualmente quanto o funcionamento de várias operações em sequência.

## 6. Como executar o projeto

### Requisitos

- Go instalado na máquina.
- Acesso ao terminal.
- Dependências do projeto disponíveis para instalação.

### Executar a aplicação

Na raiz do projeto, execute:

```bash
go run ./cmd/demo
```

O programa abrirá uma interface interativa no terminal, permitindo registrar despesas e consultar o orçamento disponível.

### Executar todos os testes

```bash
go test ./... -v
```

O comando executa os testes de todos os pacotes e apresenta os resultados detalhados.

### Executar um teste específico

```bash
go test -run TestMustNewExpense -v
```

Permite executar somente o teste cujo nome corresponde ao padrão informado.

### Executar um método específico de uma Suite

```bash
go test -run TestServiceSuite -testify.m TestUsedPercent -v
```

Permite executar especificamente o método `TestUsedPercent` dentro da suíte `TestServiceSuite`.

### Verificar a cobertura dos testes

```bash
go test -cover ./...
```

Apresenta a porcentagem de cobertura de código dos testes em cada pacote.

## 7. Considerações finais

O MoneyFlow demonstra como a biblioteca Testify pode auxiliar na criação e organização de testes automatizados em Go.

Por meio dos recursos apresentados, é possível realizar verificações de resultados, interromper testes quando necessário, simular dependências e organizar conjuntos de testes.

A utilização de Mock permite testar o comportamento da aplicação sem depender de componentes externos, enquanto o Suite facilita a preparação e a organização dos cenários de teste.

Dessa forma, o projeto reúne diferentes recursos do Testify em uma aplicação simples, demonstrando suas possibilidades de utilização em situações práticas de desenvolvimento de software.
