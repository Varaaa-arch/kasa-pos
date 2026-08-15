package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/product"
)

var ErrProductNotFound = errors.New("product not found")

var ErrInsufficientStock = errors.New("insufficient stock")

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(
	ctx context.Context,
	p product.Product,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO products (
			id,
			sku,
			name,
			price,
			stock,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		p.ID,
		p.SKU,
		p.Name,
		p.Price,
		p.Stock,
		p.CreatedAt,
		p.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id string,
) (product.Product, error) {
	if _, err := uuid.Parse(id); err != nil {
		return product.Product{}, ErrProductNotFound
	}

	var p product.Product

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			sku,
			name,
			price,
			stock,
			created_at,
			updated_at
		FROM products
		WHERE id = $1
		`,
		id,
	).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Price,
		&p.Stock,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return product.Product{}, ErrProductNotFound
	}

	if err != nil {
		return product.Product{}, fmt.Errorf(
			"get product by id: %w",
			err,
		)
	}

	return p, nil
}

func (r *ProductRepository) GetBySKU(
	ctx context.Context,
	sku string,
) (product.Product, error) {
	var p product.Product

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			sku,
			name,
			price,
			stock,
			created_at,
			updated_at
		FROM products
		WHERE sku = $1
		`,
		sku,
	).Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Price,
		&p.Stock,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return product.Product{}, ErrProductNotFound
	}

	if err != nil {
		return product.Product{}, fmt.Errorf(
			"get product by sku: %w",
			err,
		)
	}

	return p, nil
}

func (r *ProductRepository) List(
	ctx context.Context,
) ([]product.Product, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			sku,
			name,
			price,
			stock,
			created_at,
			updated_at
		FROM products
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list products: %w",
			err,
		)
	}
	defer rows.Close()

	var products []product.Product

	for rows.Next() {
		var p product.Product

		if err := rows.Scan(
			&p.ID,
			&p.SKU,
			&p.Name,
			&p.Price,
			&p.Stock,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan product: %w",
				err,
			)
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate products: %w",
			err,
		)
	}

	return products, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	p product.Product,
) error {
	if _, err := uuid.Parse(p.ID); err != nil {
		return ErrProductNotFound
	}

	result, err := r.db.ExecContext(
		ctx,
		`
		UPDATE products
		SET
			sku = $2,
			name = $3,
			price = $4,
			stock = $5,
			updated_at = $6
		WHERE id = $1
		`,
		p.ID,
		p.SKU,
		p.Name,
		p.Price,
		p.Stock,
		p.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"update product rows affected: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) Delete(
	ctx context.Context,
	id string,
) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrProductNotFound
	}

	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM products WHERE id = $1`,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"delete product: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"delete product rows affected: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) ReduceStockTx(
	ctx context.Context,
	tx db.DBTX,
	productID string,
	quantity int,
) error {
	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE products
		SET
			stock = stock - $1,
			updated_at = NOW()
		WHERE id = $2
		AND stock >= $1
		`,
		quantity,
		productID,
	)
	if err != nil {
		return fmt.Errorf(
			"reduce stock: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"reduce stock rows affected: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrInsufficientStock
	}

	return nil
}
