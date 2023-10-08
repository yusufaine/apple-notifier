package main

import (
	"flag"
	"log/slog"
	"net/http"

	"github.com/yusufaine/apple-inventory-notifier/cloudfunction"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.Parse()

	if port == "" {
		slog.Error("port must be set")
		return
	}

	// Exact same handler
	http.HandleFunc("/", cloudfunction.AppleNotifierHandler)

	slog.Info("Starting notifier server...", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error(err.Error())
	}
}
