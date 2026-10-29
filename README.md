# MoneyFlow — Demonstração do Testify

Projeto pequeno só para mostrar a biblioteca **testify**. Controla despesas
dentro de um orçamento: sem banco de dados, sem HTTP.

## O que é o testify?

Uma biblioteca que deixa os testes em Go mais curtos e legíveis. Sem ela:

```go
if got != want {
    t.Errorf("esperado %v, obtido %v", want, got)
}
```

Com ela:

```go
assert.Equal(t, want, got)
```

Ela tem quatro partes, todas usadas aqui:

| Pacote | Para quê | Onde ver |
|---|---|---|
| `assert` | Verificações que, se falham, marcam o teste como falho e **continuam** | `expense_test.go` |
| `require` | As mesmas verificações, mas **interrompem** o teste na falha | `expense_test.go` |
| `mock` | Objetos falsos que registram chamadas e devolvem o que o teste programou | `mocks_test.go`, `service_mock_test.go` |
| `suite` | Agrupa testes numa struct com `Setup`/`TearDown` | `service_suite_test.go` |

## Ver funcionando

```bash
go run ./cmd/demo
```

Abre um terminal interativo (orçamento de R$ 100,00):

```
> add 19,90 alimentacao Almoço
> add 70 casa Aluguel
> add 20 lazer Cinema        (rejeitada: passa do saldo)
> add 10,10 lazer Pipoca     (gasta exatamente o saldo e dispara o aviso)
> saldo
> lista
```

## Estrutura

```
cmd/demo/main.go       programa interativo para mostrar na prática
expense.go             Expense, validação e erros
repository.go          interface Repository + MemoryRepository (fake em memória)
service.go             regras: validar, conferir orçamento, salvar, avisar
mocks_test.go          MockRepository e MockNotifier (testify/mock)
expense_test.go        assert e require
service_mock_test.go   mock
service_suite_test.go  suite
```

## Como rodar

```bash
go test ./... -v                 # todos os testes
go test -run TestMustNewExpense -v  # um teste só
go test -run TestServiceSuite -testify.m TestUsedPercent -v  # um método da suite
go test -cover ./...             # com cobertura
```

## Conceitos do testify e onde aparecem

**assert × require** (`expense_test.go`)
- `require.NoError` como pré-condição; `assert.Equal` para o resultado.
- `assert.New(t)` / `require.New(t)`: o mesmo sem repetir o `t`.
- Mensagem extra opcional no último argumento.

**Funções de verificação** (`expense_test.go`, `service_suite_test.go`)
- Igualdade e vazio: `Equal`, `Zero`, `Empty`
- Erros: `Error`, `NoError`, `ErrorIs`, `NotErrorIs`, `ErrorAs`, `ErrorContains`
- Coleções: `Len`, `Contains`, `NotContains`, `ElementsMatch`
- Outros: `JSONEq`, `InDelta`, `Panics`, `NotPanics`
- Tempo: `Eventually` (espera algo assíncrono) e `Never` (garante que não acontece)

**mock** (`service_mock_test.go`)
- Programar: `.On("Método", args...).Return(valores...)`
- Argumentos: valor exato, `mock.Anything`, `mock.AnythingOfType`, `mock.MatchedBy`
- Quantidade de chamadas: `.Once()`, `.Times(n)`, `.Maybe()`
- Reagir à chamada: `.Run(func(args mock.Arguments) {...})`
- Verificar: `AssertExpectations`, `AssertCalled`, `AssertNotCalled`, `AssertNumberOfCalls`
- Simular falhas: `Return(errDB)`

**suite** (`service_suite_test.go`)
- `SetupSuite`, `SetupTest`, `SetupSubTest`, `TearDownTest`, `TearDownSuite`
- Asserts direto em `s` (`s.Equal`) e `s.Require()` para a versão que interrompe
- `s.Run` para subtestes; `suite.Run(t, new(MinhaSuite))` como ponto de entrada

**Mock × fake**: nos testes com mock o `Service` conversa com um repositório
falso que o teste controla (dá para simular erro de banco). Na suite usamos o
`MemoryRepository`, um fake que funciona de verdade, para ver o resultado de
várias operações em sequência.

## Sugestão de ordem para explicar

1. O problema que o testify resolve (`t.Errorf` × `assert.Equal`).
2. `expense_test.go`: assert, require e as funções de verificação.
3. `service_mock_test.go`: por que mockar e como (programar → executar → verificar).
4. `service_suite_test.go`: organizar testes com ciclo de vida.
