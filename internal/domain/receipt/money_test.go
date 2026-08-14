package receipt

import "testing"

func TestNewMoney(t *testing.T) {
	money := NewMoney(15000, IDR)

	if money.Amount != 15000 {
		t.Fatalf(
			"expected amount 15000, got %d",
			money.Amount,
		)
	}

	if money.Currency != IDR {
		t.Fatalf(
			"expected currency %q, got %q",
			IDR,
			money.Currency,
		)
	}
}

func TestZeroMoney(t *testing.T) {
	money := ZeroMoney(IDR)

	if !money.IsZero() {
		t.Fatal("expected zero money")
	}

	if money.Amount != 0 {
		t.Fatalf(
			"expected amount 0, got %d",
			money.Amount,
		)
	}

	if money.Currency != IDR {
		t.Fatalf(
			"expected currency %q, got %q",
			IDR,
			money.Currency,
		)
	}
}

func TestMoneyAdd(t *testing.T) {
	a := NewMoney(15000, IDR)
	b := NewMoney(30000, IDR)

	got := a.Add(b)

	if got.Amount != 45000 {
		t.Fatalf(
			"expected 45000, got %d",
			got.Amount,
		)
	}

	if got.Currency != IDR {
		t.Fatalf(
			"expected currency %q, got %q",
			IDR,
			got.Currency,
		)
	}
}

func TestMoneySub(t *testing.T) {
	a := NewMoney(50000, IDR)
	b := NewMoney(30000, IDR)

	got := a.Sub(b)

	if got.Amount != 20000 {
		t.Fatalf(
			"expected 20000, got %d",
			got.Amount,
		)
	}
}

func TestMoneyMul(t *testing.T) {
	money := NewMoney(15000, IDR)

	got := money.Mul(3)

	if got.Amount != 45000 {
		t.Fatalf(
			"expected 45000, got %d",
			got.Amount,
		)
	}
}

func TestMoneyCompare(t *testing.T) {
	tests := []struct {
		name string
		a    Money
		b    Money
		want int
	}{
		{
			name: "less",
			a:    NewMoney(10000, IDR),
			b:    NewMoney(20000, IDR),
			want: -1,
		},
		{
			name: "equal",
			a:    NewMoney(20000, IDR),
			b:    NewMoney(20000, IDR),
			want: 0,
		},
		{
			name: "greater",
			a:    NewMoney(30000, IDR),
			b:    NewMoney(20000, IDR),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Compare(tt.b)

			if got != tt.want {
				t.Fatalf(
					"Compare() = %d, want %d",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestMoneyEquality(t *testing.T) {
	a := NewMoney(15000, IDR)
	b := NewMoney(15000, IDR)
	c := NewMoney(20000, IDR)

	if !a.Equal(b) {
		t.Fatal("expected equal money values")
	}

	if a.Equal(c) {
		t.Fatal("expected different money values")
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{
			name:  "zero",
			money: NewMoney(0, IDR),
			want:  "Rp0",
		},
		{
			name:  "thousand",
			money: NewMoney(1000, IDR),
			want:  "Rp1.000",
		},
		{
			name:  "ten thousand",
			money: NewMoney(15000, IDR),
			want:  "Rp15.000",
		},
		{
			name:  "million",
			money: NewMoney(1500000, IDR),
			want:  "Rp1.500.000",
		},
		{
			name:  "negative",
			money: NewMoney(-15000, IDR),
			want:  "-Rp15.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.money.String()

			if got != tt.want {
				t.Fatalf(
					"String() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestMoneyComparisonHelpers(t *testing.T) {
	a := NewMoney(15000, IDR)
	b := NewMoney(10000, IDR)

	if !a.GreaterThan(b) {
		t.Fatal("expected a > b")
	}

	if !a.GreaterThanOrEqual(b) {
		t.Fatal("expected a >= b")
	}

	if !b.LessThan(a) {
		t.Fatal("expected b < a")
	}

	if !b.LessThanOrEqual(a) {
		t.Fatal("expected b <= a")
	}

	if !a.GreaterThanOrEqual(a) {
		t.Fatal("expected a >= a")
	}

	if !a.LessThanOrEqual(a) {
		t.Fatal("expected a <= a")
	}
}

func TestMoneyNegative(t *testing.T) {
	money := NewMoney(-5000, IDR)

	if !money.IsNegative() {
		t.Fatal("expected money to be negative")
	}

	zero := NewMoney(0, IDR)

	if zero.IsNegative() {
		t.Fatal("zero should not be negative")
	}
}
