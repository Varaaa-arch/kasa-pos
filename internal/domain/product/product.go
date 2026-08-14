package product

import "time"

type Product struct {
	ID        string
	SKU       string
	Name      string
	Price     int64
	Stock     int
	CreatedAt time.Time
	UpdatedAt time.Time
}
