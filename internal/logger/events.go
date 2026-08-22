package logger

// Event constants untuk structured logging.
// Dipakai sebagai nilai field "event" di setiap log entry
// sehingga log bisa di-filter/query berdasarkan event name.
//
// Contoh:
//
//	slog.InfoContext(ctx, EventCheckoutCompleted,
//	    "event", EventCheckoutCompleted,
//	    "transaction_id", tx.ID,
//	    "request_id", logger.RequestIDFromContext(ctx),
//	)
const (
	// Checkout flow
	EventCheckoutStarted   = "checkout_started"
	EventCheckoutCompleted = "checkout_completed"
	EventCheckoutFailed    = "checkout_failed"

	// Print flow
	EventPrintStarted   = "print_started"
	EventPrintCompleted = "print_completed"
	EventPrintFailed    = "print_failed"

	// Infrastructure
	EventDBError = "db_error"
)
