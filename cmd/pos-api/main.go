package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"pos-system/internal/api"
	"pos-system/internal/db"
	"pos-system/internal/printer/agent"
	postgres "pos-system/internal/repository/postgres"
	checkoutsvc "pos-system/internal/service/checkout"
	productservice "pos-system/internal/service/product"
	receiptsvc "pos-system/internal/service/receipt"
	transactionservice "pos-system/internal/service/transaction"
)

func main() {
	database, err := db.OpenPostgres(
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

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
		Handler: api.NewRouter(productHandler, transactionHandler, checkoutHandler),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop

		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("POS API listening on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
