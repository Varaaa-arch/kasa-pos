package checkout

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	domaintransaction "pos-system/internal/domain/transaction"
)

// TxTransactionRepository is the write side of the transaction repo
// scoped to a single sql.Tx. Extracted so AtomicService can be tested
// without a real database.
type TxTransactionRepository interface {
	CreateTx(
		ctx context.Context,
		tx db.DBTX,
		t domaintransaction.Transaction,
	) error
}

// TxProductRepository is the write side of the product repo
// scoped to a single sql.Tx. Extracted so AtomicService can be tested
// without a real database.
type TxProductRepository interface {
	ReduceStockTx(
		ctx context.Context,
		tx db.DBTX,
		productID string,
		quantity int,
	) error
}

// TxBeginner is satisfied by *sql.DB and any fake that can open a transaction.
type TxBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type AtomicService struct {
	db              TxBeginner
	transactionRepo TxTransactionRepository
	productRepo     TxProductRepository
}

func NewAtomicService(
	db TxBeginner,
	transactionRepo TxTransactionRepository,
	productRepo TxProductRepository,
) *AtomicService {
	return &AtomicService{
		db:              db,
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
	}
}

type AtomicRequest struct {
	Cart          *cart.Cart
	PaidAmount    int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
	PaymentMethod string
	InvoiceNumber string
}

func (s *AtomicService) Execute(
	ctx context.Context,
	req AtomicRequest,
) (domaintransaction.Transaction, error) {
	if req.Cart == nil || len(req.Cart.Items) == 0 {
		return domaintransaction.Transaction{}, ErrEmptyCart
	}

	if strings.TrimSpace(req.PaymentMethod) == "" {
		req.PaymentMethod = "CASH"
	}

	subtotal := req.Cart.Total

	total := subtotal -
		req.Discount +
		req.Tax +
		req.ServiceCharge

	if total < 0 {
		total = 0
	}

	if req.PaidAmount < total {
		return domaintransaction.Transaction{},
			ErrInsufficientCash
	}

	if req.InvoiceNumber == "" {
		req.InvoiceNumber = "INV-" + uuid.NewString()
	}

	now := time.Now().UTC()
	transactionID := uuid.NewString()

	trx := domaintransaction.Transaction{
		ID:            transactionID,
		InvoiceNumber: req.InvoiceNumber,
		Subtotal:      subtotal,
		Discount:      req.Discount,
		Tax:           req.Tax,
		ServiceCharge: req.ServiceCharge,
		Total:         total,
		PaidAmount:    req.PaidAmount,
		Change:        req.PaidAmount - total,
		PaymentMethod: req.PaymentMethod,
		Status:        domaintransaction.StatusCompleted,
		CreatedAt:     now,
		Items:         make([]domaintransaction.Item, 0, len(req.Cart.Items)),
	}

	for _, item := range req.Cart.Items {
		trx.Items = append(
			trx.Items,
			domaintransaction.Item{
				ID:            uuid.NewString(),
				TransactionID: transactionID,
				ProductID:     item.Product.ID,
				SKU:           item.Product.SKU,
				Name:          item.Product.Name,
				Quantity:      item.Quantity,
				UnitPrice:     item.UnitPrice,
				Subtotal:      item.Subtotal,
			},
		)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domaintransaction.Transaction{}, fmt.Errorf(
			"begin checkout transaction: %w",
			err,
		)
	}

	defer tx.Rollback()

	if err := s.transactionRepo.CreateTx(ctx, tx, trx); err != nil {
		return domaintransaction.Transaction{}, err
	}

	for _, item := range req.Cart.Items {
		if err := s.productRepo.ReduceStockTx(
			ctx,
			tx,
			item.Product.ID,
			item.Quantity,
		); err != nil {
			return domaintransaction.Transaction{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domaintransaction.Transaction{}, fmt.Errorf(
			"commit checkout: %w",
			err,
		)
	}

	return trx, nil
}
