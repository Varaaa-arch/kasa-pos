package main

import (
	"log"
	"net/http"

	"pos-system/internal/printer/api"
	"pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const (
	printerDevice = "/dev/usb/lp0"
	listenAddress = "127.0.0.1:8080"
)

func main() {
	printer := transport.NewUSBPrinter(printerDevice)
	renderer := receipt.NewRenderer()

	handler := api.NewHandler(
		printer,
		renderer,
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/print",
		handler.Print,
	)

	server := &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	log.Printf(
		"print agent listening on http://%s",
		listenAddress,
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
