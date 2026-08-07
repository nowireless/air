package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simonvetter/modbus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const Address uint16 = 1

func main() {
	// Set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewTextHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	slog.Info("Banner S24 Dew Point Sensor")

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://192.168.9.8:502",
		Timeout: 1 * time.Second,
	})
	defer client.Close()
	if err != nil {
		slog.Error("Failed to create modbus client", "err", err)
		os.Exit(1)
	}

	client.SetUnitId(1)

	slog.Info("Opening modbus client")
	if err := client.Open(); err != nil {
		slog.Error("Failed to open modbus client", "err", err)
		os.Exit(1)
	}

	// Setup metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewCollector(client))

	slog.Info("Starting HTTP server")
	reg.MustRegister()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":2112", nil)
}
