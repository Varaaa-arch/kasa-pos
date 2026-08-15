package transaction

import "context"

type Repository interface {
	Create(ctx context.Context, tx Transaction) error
	GetByID(ctx context.Context, id string) (Transaction, error)
	List(ctx context.Context) ([]Transaction, error)
}
