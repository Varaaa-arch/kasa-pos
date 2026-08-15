package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/transaction"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(
	db *sql.DB,
) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	t transaction.Transaction,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO transactions (
			id,
			invoice_number,
			subtotal,
			discount,
			tax,
			service_charge,
			total,
			paid_amount,
			change,
			payment_method,
			status,
			created_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		)
		`,
		t.ID,
		t.InvoiceNumber,
		t.Subtotal,
		t.Discount,
		t.Tax,
		t.ServiceCharge,
		t.Total,
		t.PaidAmount,
		t.Change,
		t.PaymentMethod,
		t.Status,
		t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"insert transaction: %w",
			err,
		)
	}

	for _, item := range t.Items {
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO transaction_items (
				id,
				transaction_id,
				product_id,
				sku,
				name,
				quantity,
				unit_price,
				subtotal
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			`,
			item.ID,
			item.TransactionID,
			item.ProductID,
			item.SKU,
			item.Name,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
		)
		if err != nil {
			return fmt.Errorf(
				"insert transaction item: %w",
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"commit transaction: %w",
			err,
		)
	}

	return nil
}

func (r *TransactionRepository) CreateTx(
	ctx context.Context,
	tx db.DBTX,
	t transaction.Transaction,
) error {
	_, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO transactions (
			id,
			invoice_number,
			subtotal,
			discount,
			tax,
			service_charge,
			total,
			paid_amount,
			change,
			payment_method,
			status,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`,
		t.ID,
		t.InvoiceNumber,
		t.Subtotal,
		t.Discount,
		t.Tax,
		t.ServiceCharge,
		t.Total,
		t.PaidAmount,
		t.Change,
		t.PaymentMethod,
		t.Status,
		t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	for _, item := range t.Items {
		if _, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO transaction_items (
				id,
				transaction_id,
				product_id,
				sku,
				name,
				quantity,
				unit_price,
				subtotal
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			`,
			item.ID,
			item.TransactionID,
			item.ProductID,
			item.SKU,
			item.Name,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
		); err != nil {
			return fmt.Errorf(
				"insert transaction item: %w",
				err,
			)
		}
	}

	return nil
}

func (r *TransactionRepository) GetByID(
	ctx context.Context,
	id string,
) (transaction.Transaction, error) {
	if _, err := uuid.Parse(id); err != nil {
		return transaction.Transaction{}, ErrTransactionNotFound
	}

	var t transaction.Transaction

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			invoice_number,
			subtotal,
			discount,
			tax,
			service_charge,
			total,
			paid_amount,
			change,
			payment_method,
			status,
			created_at
		FROM transactions
		WHERE id = $1
		`,
		id,
	).Scan(
		&t.ID,
		&t.InvoiceNumber,
		&t.Subtotal,
		&t.Discount,
		&t.Tax,
		&t.ServiceCharge,
		&t.Total,
		&t.PaidAmount,
		&t.Change,
		&t.PaymentMethod,
		&t.Status,
		&t.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return transaction.Transaction{},
			ErrTransactionNotFound
	}

	if err != nil {
		return transaction.Transaction{},
			fmt.Errorf(
				"get transaction: %w",
				err,
			)
	}

	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			transaction_id,
			product_id,
			sku,
			name,
			quantity,
			unit_price,
			subtotal
		FROM transaction_items
		WHERE transaction_id = $1
		ORDER BY id
		`,
		id,
	)
	if err != nil {
		return transaction.Transaction{},
			fmt.Errorf(
				"get transaction items: %w",
				err,
			)
	}
	defer rows.Close()

	for rows.Next() {
		var item transaction.Item

		if err := rows.Scan(
			&item.ID,
			&item.TransactionID,
			&item.ProductID,
			&item.SKU,
			&item.Name,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
		); err != nil {
			return transaction.Transaction{},
				fmt.Errorf(
					"scan transaction item: %w",
					err,
				)
		}

		t.Items = append(t.Items, item)
	}

	if err := rows.Err(); err != nil {
		return transaction.Transaction{},
			fmt.Errorf(
				"iterate transaction items: %w",
				err,
			)
	}

	return t, nil
}

func (r *TransactionRepository) List(
	ctx context.Context,
) ([]transaction.Transaction, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			invoice_number,
			subtotal,
			discount,
			tax,
			service_charge,
			total,
			paid_amount,
			change,
			payment_method,
			status,
			created_at
		FROM transactions
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list transactions: %w",
			err,
		)
	}
	defer rows.Close()

	var result []transaction.Transaction

	for rows.Next() {
		var t transaction.Transaction

		if err := rows.Scan(
			&t.ID,
			&t.InvoiceNumber,
			&t.Subtotal,
			&t.Discount,
			&t.Tax,
			&t.ServiceCharge,
			&t.Total,
			&t.PaidAmount,
			&t.Change,
			&t.PaymentMethod,
			&t.Status,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan transaction: %w",
				err,
			)
		}

		result = append(result, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate transactions: %w",
			err,
		)
	}

	return result, nil
}
