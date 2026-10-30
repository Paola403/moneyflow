// Programa de demonstração que disponibiliza as operações do MoneyFlow
// por meio de uma interface simples de terminal.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"moneyflow"
)

// consoleNotifier implementa Notifier exibindo as mensagens diretamente no terminal.
type consoleNotifier struct{}

func (consoleNotifier) Notify(message string) error {
	fmt.Println("\n🔔", message)
	return nil
}

func reais(cents int64) string { return fmt.Sprintf("R$ %.2f", float64(cents)/100) }

func main() {
	const limit = 10_000 // R$ 100,00

	repo := moneyflow.NewMemoryRepository()
	svc := moneyflow.NewService(repo, consoleNotifier{}, limit)
	ctx := context.Background()

	fmt.Printf("MoneyFlow — orçamento de %s. Digite 'ajuda' para ver os comandos.\n", reais(limit))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "add":
			add(ctx, svc, fields[1:])
		case "saldo":
			balance, err := svc.Balance(ctx)
			if err != nil {
				fmt.Println("✗", err)
				continue
			}
			percent, _ := svc.UsedPercent(ctx)
			fmt.Printf("Saldo: %s (%.2f%% do orçamento usado)\n", reais(balance), percent)
		case "lista":
			expenses, err := repo.List(ctx)
			if err != nil {
				fmt.Println("✗", err)
				continue
			}
			if len(expenses) == 0 {
				fmt.Println("Nenhuma despesa ainda.")
			}
			for _, e := range expenses {
				fmt.Printf("%s  %-20s %10s  [%s]\n", e.ID, e.Description, reais(e.AmountCents), e.Category)
			}
		case "ajuda":
			fmt.Println("  add <valor> <categoria> <descrição>   ex.: add 19,90 alimentacao Almoço")
			fmt.Println("  saldo                                 mostra quanto ainda pode gastar")
			fmt.Println("  lista                                 lista as despesas")
			fmt.Println("  sair")
		case "sair":
			return
		default:
			fmt.Println("Comando desconhecido. Digite 'ajuda'.")
		}
	}
}

func add(ctx context.Context, svc *moneyflow.Service, args []string) {
	if len(args) < 3 {
		fmt.Println("Uso: add <valor> <categoria> <descrição>")
		return
	}
	value, err := strconv.ParseFloat(strings.Replace(args[0], ",", ".", 1), 64)
	if err != nil {
		fmt.Println("✗ valor inválido:", args[0])
		return
	}
	cents := int64(math.Round(value * 100))

	expense, err := svc.CreateExpense(ctx, strings.Join(args[2:], " "), cents, args[1])

	var exceeded *moneyflow.BudgetExceededError
	switch {
	case errors.As(err, &exceeded):
		fmt.Printf("✗ Orçamento insuficiente: pediu %s, disponível %s\n",
			reais(exceeded.RequestedCents), reais(exceeded.AvailableCents))
	case err != nil:
		fmt.Println("✗", err)
	default:
		fmt.Printf("✓ %s registrada: %s (%s)\n", expense.ID, reais(expense.AmountCents), expense.Description)
	}
}
