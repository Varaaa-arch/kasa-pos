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
	postgres "pos-system/internal/repository/postgres"
	productservice "pos-system/internal/service/product"
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
	productSvc := productservice.NewService(productRepo)
	productHandler := api.NewProductHandler(productSvc)

	server := &http.Server{
		Addr:    ":8080",
		Handler: api.NewRouter(productHandler),
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
