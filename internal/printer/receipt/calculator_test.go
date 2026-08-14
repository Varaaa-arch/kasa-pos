package receipt

import (
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
)

func assertMoney(
	t *testing.T,
	got domainreceipt.Money,
	wantAmount int64,
	wantCurrency domainreceipt.Currency,
) {
	t.Helper()

	if got.Amount != wantAmount {
		t.Fatalf(
			"amount = %d, want %d",
			got.Amount,
			wantAmount,
		)
	}

	if got.Currency != wantCurrency {
		t.Fatalf(
			"currency = %q, want %q",
			got.Currency,
			wantCurrency,
		)
	}
}

func TestCalculator(t *testing.T) {
	calculator := NewCalculator()

	tests := []struct {
		name           string
		input          domainreceipt.Receipt
		wantItemsTotal int64
		wantSubtotal   int64
		wantDiscount   int64
		wantTax        int64
		wantService    int64
		wantTotal      int64
		wantPaid       int64
		wantChange     int64
	}{
		{
			name: "basic receipt",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  2,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
					},
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 42000,
			wantSubtotal:   42000,
			wantDiscount:   0,
			wantTax:        0,
			wantService:    0,
			wantTotal:      42000,
			wantPaid:       50000,
			wantChange:     8000,
		},
		{
			name: "with discount",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  2,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
					},
				},
				Summary: domainreceipt.Summary{
					Discount: domainreceipt.NewMoney(5000, domainreceipt.IDR),
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 42000,
			wantSubtotal:   42000,
			wantDiscount:   5000,
			wantTax:        0,
			wantService:    0,
			wantTotal:      37000,
			wantPaid:       50000,
			wantChange:     13000,
		},
		{
			name: "with tax",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  2,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
					},
				},
				Summary: domainreceipt.Summary{
					Tax: domainreceipt.NewMoney(4200, domainreceipt.IDR),
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 42000,
			wantSubtotal:   42000,
			wantDiscount:   0,
			wantTax:        4200,
			wantService:    0,
			wantTotal:      46200,
			wantPaid:       50000,
			wantChange:     3800,
		},
		{
			name: "with service charge",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  2,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
				},
				Summary: domainreceipt.Summary{
					ServiceCharge: domainreceipt.NewMoney(3000, domainreceipt.IDR),
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(35000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 30000,
			wantSubtotal:   30000,
			wantDiscount:   0,
			wantTax:        0,
			wantService:    3000,
			wantTotal:      33000,
			wantPaid:       35000,
			wantChange:     2000,
		},
		{
			name: "discount tax and service charge",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  2,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
				},
				Summary: domainreceipt.Summary{
					Discount:      domainreceipt.NewMoney(2000, domainreceipt.IDR),
					Tax:           domainreceipt.NewMoney(2800, domainreceipt.IDR),
					ServiceCharge: domainreceipt.NewMoney(1000, domainreceipt.IDR),
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(40000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 30000,
			wantSubtotal:   30000,
			wantDiscount:   2000,
			wantTax:        2800,
			wantService:    1000,
			wantTotal:      31800,
			wantPaid:       40000,
			wantChange:     8200,
		},
		{
			name: "exact payment",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  1,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(15000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 15000,
			wantSubtotal:   15000,
			wantTotal:      15000,
			wantPaid:       15000,
			wantChange:     0,
		},
		{
			name: "insufficient payment",
			input: domainreceipt.Receipt{
				Items: []domainreceipt.Item{
					{
						Name:      "Kopi Susu",
						Quantity:  1,
						UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
					},
				},
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(10000, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 15000,
			wantSubtotal:   15000,
			wantTotal:      15000,
			wantPaid:       10000,
			wantChange:     0,
		},
		{
			name: "empty receipt",
			input: domainreceipt.Receipt{
				Payment: domainreceipt.Payment{
					Paid: domainreceipt.NewMoney(0, domainreceipt.IDR),
				},
			},
			wantItemsTotal: 0,
			wantSubtotal:   0,
			wantDiscount:   0,
			wantTax:        0,
			wantService:    0,
			wantTotal:      0,
			wantPaid:       0,
			wantChange:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculator.Calculate(tt.input)

			if got.ItemsTotal.Amount != tt.wantItemsTotal {
				t.Fatalf(
					"ItemsTotal = %d, want %d",
					got.ItemsTotal.Amount,
					tt.wantItemsTotal,
				)
			}

			if got.Subtotal.Amount != tt.wantSubtotal {
				t.Fatalf(
					"Subtotal = %d, want %d",
					got.Subtotal.Amount,
					tt.wantSubtotal,
				)
			}

			if got.Discount.Amount != tt.wantDiscount {
				t.Fatalf(
					"Discount = %d, want %d",
					got.Discount.Amount,
					tt.wantDiscount,
				)
			}

			if got.Tax.Amount != tt.wantTax {
				t.Fatalf(
					"Tax = %d, want %d",
					got.Tax.Amount,
					tt.wantTax,
				)
			}

			if got.ServiceCharge.Amount != tt.wantService {
				t.Fatalf(
					"ServiceCharge = %d, want %d",
					got.ServiceCharge.Amount,
					tt.wantService,
				)
			}

			if got.Total.Amount != tt.wantTotal {
				t.Fatalf(
					"Total = %d, want %d",
					got.Total.Amount,
					tt.wantTotal,
				)
			}

			if got.Paid.Amount != tt.wantPaid {
				t.Fatalf(
					"Paid = %d, want %d",
					got.Paid.Amount,
					tt.wantPaid,
				)
			}

			if got.Change.Amount != tt.wantChange {
				t.Fatalf(
					"Change = %d, want %d",
					got.Change.Amount,
					tt.wantChange,
				)
			}
		})
	}
}

func TestCalculatorItemTotal(t *testing.T) {
	calculator := NewCalculator()

	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Produk A",
				Quantity:  99,
				UnitPrice: domainreceipt.NewMoney(12500000, domainreceipt.IDR),
			},
		},
		Payment: domainreceipt.Payment{
			Paid: domainreceipt.NewMoney(2000000000, domainreceipt.IDR),
		},
	}

	got := calculator.Calculate(input)

	want := int64(99) * 12500000

	assertMoney(
		t,
		got.ItemsTotal,
		want,
		domainreceipt.IDR,
	)

	assertMoney(
		t,
		got.Subtotal,
		want,
		domainreceipt.IDR,
	)
}

func TestCalculatorDoesNotUseExistingSummarySubtotal(t *testing.T) {
	calculator := NewCalculator()

	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
		},
		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(999999, domainreceipt.IDR),
			Total:    domainreceipt.NewMoney(999999, domainreceipt.IDR),
		},
		Payment: domainreceipt.Payment{
			Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
		},
	}

	got := calculator.Calculate(input)

	assertMoney(
		t,
		got.Subtotal,
		30000,
		domainreceipt.IDR,
	)

	assertMoney(
		t,
		got.Total,
		30000,
		domainreceipt.IDR,
	)
}

func TestCalculatorIsSingleSourceOfTruth(t *testing.T) {
	calculator := NewCalculator()

	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
			{
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal:      domainreceipt.NewMoney(999999, domainreceipt.IDR),
			Discount:      domainreceipt.NewMoney(5000, domainreceipt.IDR),
			Tax:           domainreceipt.NewMoney(4200, domainreceipt.IDR),
			ServiceCharge: domainreceipt.NewMoney(1000, domainreceipt.IDR),
			Total:         domainreceipt.NewMoney(999999, domainreceipt.IDR),
		},

		Payment: domainreceipt.Payment{
			Paid:   domainreceipt.NewMoney(50000, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(999999, domainreceipt.IDR),
		},
	}

	got := calculator.Calculate(input)

	expectedSubtotal := int64(42000)
	expectedTotal := int64(42200)
	expectedChange := int64(7800)

	assertMoney(
		t,
		got.Subtotal,
		expectedSubtotal,
		domainreceipt.IDR,
	)

	assertMoney(
		t,
		got.Total,
		expectedTotal,
		domainreceipt.IDR,
	)

	assertMoney(
		t,
		got.Change,
		expectedChange,
		domainreceipt.IDR,
	)
}
