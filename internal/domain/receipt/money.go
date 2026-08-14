package receipt

import (
	"fmt"
	"strings"
)

type Currency string

const (
	IDR Currency = "IDR"
)

type Money struct {
	Amount   int64
	Currency Currency
}

func NewMoney(amount int64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func ZeroMoney(currency Currency) Money {
	return Money{
		Amount:   0,
		Currency: currency,
	}
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) IsNegative() bool {
	return m.Amount < 0
}

func (m Money) Add(other Money) Money {
	if m.Currency != other.Currency {
		panic("cannot add money with different currencies")
	}

	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}
}

func (m Money) Sub(other Money) Money {
	if m.Currency != other.Currency {
		panic("cannot subtract money with different currencies")
	}

	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}
}

func (m Money) Mul(quantity int64) Money {
	return Money{
		Amount:   m.Amount * quantity,
		Currency: m.Currency,
	}
}

func (m Money) Compare(other Money) int {
	if m.Currency != other.Currency {
		panic("cannot compare money with different currencies")
	}

	switch {
	case m.Amount < other.Amount:
		return -1

	case m.Amount > other.Amount:
		return 1

	default:
		return 0
	}
}

func (m Money) Equal(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}

	return m.Amount == other.Amount
}

func (m Money) GreaterThan(other Money) bool {
	return m.Compare(other) > 0
}

func (m Money) GreaterThanOrEqual(other Money) bool {
	return m.Compare(other) >= 0
}

func (m Money) LessThan(other Money) bool {
	return m.Compare(other) < 0
}

func (m Money) LessThanOrEqual(other Money) bool {
	return m.Compare(other) <= 0
}

func (m Money) String() string {
	return formatMoneyAmount(m.Amount)
}

func (m Money) Format() string {
	return formatMoneyAmount(m.Amount)
}

func formatMoneyAmount(amount int64) string {
	if amount == 0 {
		return "Rp0"
	}

	negative := amount < 0

	if negative {
		amount = -amount
	}

	value := fmt.Sprintf("%d", amount)

	for i := len(value) - 3; i > 0; i -= 3 {
		value = value[:i] + "." + value[i:]
	}

	if negative {
		return "-Rp" + value
	}

	return "Rp" + value
}

func (c Currency) String() string {
	return strings.TrimSpace(string(c))
}
