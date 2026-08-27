package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pos-system/internal/printer/api"
	"pos-system/internal/printer/logging"
	"pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const (
	printerDevice = "/dev/usb/lp0"
	listenAddress = "127.0.0.1:8081"
	shutdownTimeout = 10 * time.Second
)

func main() {
	printer := transport.NewUSBPrinter(printerDevice)
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		printerDevice,
		logger,
		receipt.NewIdempotencyStore(),
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/print",
		handler.Print,
	)

	mux.HandleFunc(
		"/status",
		handler.Status,
	)

	mux.HandleFunc(
		"/health",
		handler.Status,
	)

	server := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Printf(
			"starting print agent on %s",
			listenAddress,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Println("shutting down print agent...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("error during shutdown: %v", err)
	}

	logger.Println("print agent stopped")
}
