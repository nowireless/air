package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	// Set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewTextHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))
	slog.Info("Laser Compressed Air Filtration collector")

	m, err := Connect("tcp://172.30.0.140:502")
	defer m.Close()
	if err != nil {
		slog.Error("Failed to create modbus client", "err", err)

		// Do not emit stale or invented sensor values.
		// return nil, err
		return
	}

	queue := make(chan *Metrics)

	go Batcher(queue, time.Second, func(metrics []*Metrics) {
		slog.Info("Batching!", "len(metrics)", len(metrics))
	})

	collectionTicker := time.NewTicker(200 * time.Millisecond)
	for range collectionTicker.C {
		metrics, err := collectMetrics(m)
		if err != nil {
			panic(err)
		}

		queue <- metrics
	}
}
