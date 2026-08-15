package main

import (
	"net/http"

	"pos-system/internal/printer/api"
	"pos-system/internal/printer/logging"
	"pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const (
	printerDevice = "/dev/usb/lp0"
	listenAddress = "127.0.0.1:8081"
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

	server := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	logger.Printf(
		"starting print agent on %s",
		listenAddress,
	)

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}
