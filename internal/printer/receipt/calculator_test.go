package receipt

import (
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
)

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
						UnitPrice: 15000,
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: 12000,
					},
				},
				Payment: domainreceipt.Payment{
					Paid: 50000,
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
						UnitPrice: 15000,
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: 12000,
					},
				},
				Summary: domainreceipt.Summary{
					Discount: 5000,
				},
				Payment: domainreceipt.Payment{
					Paid: 50000,
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
						UnitPrice: 15000,
					},
					{
						Name:      "Roti Bakar",
						Quantity:  1,
						UnitPrice: 12000,
					},
				},
				Summary: domainreceipt.Summary{
					Tax: 4200,
				},
				Payment: domainreceipt.Payment{
					Paid: 50000,
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
						UnitPrice: 15000,
					},
				},
				Summary: domainreceipt.Summary{
					ServiceCharge: 3000,
				},
				Payment: domainreceipt.Payment{
					Paid: 35000,
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
						UnitPrice: 15000,
					},
				},
				Summary: domainreceipt.Summary{
					Discount:      2000,
					Tax:           2800,
					ServiceCharge: 1000,
				},
				Payment: domainreceipt.Payment{
					Paid: 40000,
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
						UnitPrice: 15000,
					},
				},
				Payment: domainreceipt.Payment{
					Paid: 15000,
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
						UnitPrice: 15000,
					},
				},
				Payment: domainreceipt.Payment{
					Paid: 10000,
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
					Paid: 0,
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

			if got.ItemsTotal != tt.wantItemsTotal {
				t.Fatalf(
					"ItemsTotal = %d, want %d",
					got.ItemsTotal,
					tt.wantItemsTotal,
				)
			}

			if got.Subtotal != tt.wantSubtotal {
				t.Fatalf(
					"Subtotal = %d, want %d",
					got.Subtotal,
					tt.wantSubtotal,
				)
			}

			if got.Discount != tt.wantDiscount {
				t.Fatalf(
					"Discount = %d, want %d",
					got.Discount,
					tt.wantDiscount,
				)
			}

			if got.Tax != tt.wantTax {
				t.Fatalf(
					"Tax = %d, want %d",
					got.Tax,
					tt.wantTax,
				)
			}

			if got.ServiceCharge != tt.wantService {
				t.Fatalf(
					"ServiceCharge = %d, want %d",
					got.ServiceCharge,
					tt.wantService,
				)
			}

			if got.Total != tt.wantTotal {
				t.Fatalf(
					"Total = %d, want %d",
					got.Total,
					tt.wantTotal,
				)
			}

			if got.Paid != tt.wantPaid {
				t.Fatalf(
					"Paid = %d, want %d",
					got.Paid,
					tt.wantPaid,
				)
			}

			if got.Change != tt.wantChange {
				t.Fatalf(
					"Change = %d, want %d",
					got.Change,
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
				UnitPrice: 12500000,
			},
		},
		Payment: domainreceipt.Payment{
			Paid: 2000000000,
		},
	}

	got := calculator.Calculate(input)

	want := int64(99) * 12500000

	if got.ItemsTotal != want {
		t.Fatalf(
			"ItemsTotal = %d, want %d",
			got.ItemsTotal,
			want,
		)
	}

	if got.Subtotal != want {
		t.Fatalf(
			"Subtotal = %d, want %d",
			got.Subtotal,
			want,
		)
	}
}

func TestCalculatorDoesNotUseExistingSummarySubtotal(t *testing.T) {
	calculator := NewCalculator()

	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
		},
		Summary: domainreceipt.Summary{
			Subtotal: 999999,
			Total:    999999,
		},
		Payment: domainreceipt.Payment{
			Paid: 50000,
		},
	}

	got := calculator.Calculate(input)

	if got.Subtotal != 30000 {
		t.Fatalf(
			"Subtotal = %d, want 30000",
			got.Subtotal,
		)
	}

	if got.Total != 30000 {
		t.Fatalf(
			"Total = %d, want 30000",
			got.Total,
		)
	}
}

func TestCalculatorIsSingleSourceOfTruth(t *testing.T) {
	calculator := NewCalculator()

	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
			{
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: 12000,
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal:      999999,
			Discount:      5000,
			Tax:           4200,
			ServiceCharge: 1000,
			Total:         999999,
		},

		Payment: domainreceipt.Payment{
			Paid:   50000,
			Change: 999999,
		},
	}

	got := calculator.Calculate(input)

	expectedSubtotal := int64(42000)
	expectedTotal := int64(42200)
	expectedChange := int64(7800)

	if got.Subtotal != expectedSubtotal {
		t.Fatalf(
			"Subtotal = %d, want %d",
			got.Subtotal,
			expectedSubtotal,
		)
	}

	if got.Total != expectedTotal {
		t.Fatalf(
			"Total = %d, want %d",
			got.Total,
			expectedTotal,
		)
	}

	if got.Change != expectedChange {
		t.Fatalf(
			"Change = %d, want %d",
			got.Change,
			expectedChange,
		)
	}
}
