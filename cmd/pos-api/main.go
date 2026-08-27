package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pos-system/internal/api"
	"pos-system/internal/db"
	applogger "pos-system/internal/logger"
	"pos-system/internal/printer/agent"
	postgres "pos-system/internal/repository/postgres"
	checkoutsvc "pos-system/internal/service/checkout"
	productservice "pos-system/internal/service/product"
	receiptsvc "pos-system/internal/service/receipt"
	transactionservice "pos-system/internal/service/transaction"
)

func main() {
	applogger.Init()
	database, err := db.OpenPostgres(
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}

	productRepo := postgres.NewProductRepository(database)
	transactionRepo := postgres.NewTransactionRepository(database)

	productSvc := productservice.NewService(productRepo)
	atomicSvc := checkoutsvc.NewAtomicService(database, transactionRepo, productRepo)
	printAgentURL := os.Getenv("PRINT_AGENT_URL")
	if printAgentURL == "" {
		printAgentURL = "http://127.0.0.1:8081"
	}
	orchestratorSvc := checkoutsvc.NewOrchestratorService(
		atomicSvc,
		receiptsvc.NewPrintService(),
		agent.NewHTTPClient(printAgentURL),
		checkoutsvc.DefaultReceiptDefaults(),
	)
	transactionSvc := transactionservice.NewService(transactionRepo)

	productHandler := api.NewProductHandler(productSvc)
	checkoutHandler := api.NewCheckoutHandler(orchestratorSvc, productRepo)
	transactionHandler := api.NewTransactionHandler(transactionSvc)

	server := &http.Server{
		Addr:    ":8080",
		Handler: api.NewRouter(productHandler, transactionHandler, checkoutHandler, database),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop

		slog.Info("Shutting down server...")

		// Stop accepting new requests
		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown", "error", err)
		}

		// Close database connection
		slog.Info("Closing database connection...")
		if err := database.Close(); err != nil {
			slog.Error("database close", "error", err)
		}

		slog.Info("Shutdown complete")
	}()

	slog.Info("POS API listening", "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
