package cart

import (
	"errors"

	"pos-system/internal/domain/product"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Item struct {
	Product   product.Product
	Quantity  int
	UnitPrice int64
	Subtotal  int64
}

type Cart struct {
	Items []Item
	Total int64
}

func New() *Cart {
	return &Cart{
		Items: make([]Item, 0),
	}
}

func (c *Cart) AddItem(
	p product.Product,
	quantity int,
) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if p.Stock < quantity {
		return ErrInsufficientStock
	}

	for i := range c.Items {
		if c.Items[i].Product.ID == p.ID {
			c.Items[i].Quantity += quantity
			c.Items[i].Subtotal =
				int64(c.Items[i].Quantity) * p.Price

			return c.recalculate()
		}
	}

	c.Items = append(
		c.Items,
		Item{
			Product:   p,
			Quantity:  quantity,
			UnitPrice: p.Price,
			Subtotal:  int64(quantity) * p.Price,
		},
	)

	return c.recalculate()
}

func (c *Cart) RemoveItem(
	productID string,
) error {
	for i := range c.Items {
		if c.Items[i].Product.ID == productID {
			c.Items = append(
				c.Items[:i],
				c.Items[i+1:]...,
			)

			return c.recalculate()
		}
	}

	return ErrProductNotFound
}

func (c *Cart) UpdateQuantity(
	productID string,
	quantity int,
) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for i := range c.Items {
		if c.Items[i].Product.ID == productID {
			if c.Items[i].Product.Stock < quantity {
				return ErrInsufficientStock
			}

			c.Items[i].Quantity = quantity
			c.Items[i].Subtotal =
				int64(quantity) * c.Items[i].UnitPrice

			return c.recalculate()
		}
	}

	return ErrProductNotFound
}

func (c *Cart) Clear() {
	c.Items = c.Items[:0]
	c.Total = 0
}

func (c *Cart) recalculate() error {
	var total int64

	for i := range c.Items {
		c.Items[i].Subtotal =
			int64(c.Items[i].Quantity) *
				c.Items[i].UnitPrice

		total += c.Items[i].Subtotal
	}

	c.Total = total

	return nil
}
