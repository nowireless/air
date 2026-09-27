package main

import (
	"bytes"
	"log/slog"
	"net/http"
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

	go Batcher(queue, time.Second, func(batch []*Metrics) {
		slog.Info("Batching!", "len(batch)", len(batch))

		//
		// Build Payload
		//

		buf := bytes.NewBuffer(nil)
		for _, metrics := range batch {
			for _, s := range metrics.Samples {
				if _, err := buf.WriteString(s.String() + "\n"); err != nil {
					slog.Error("Failed to write sample to buffer")
					return
				}
			}
		}

		//
		// Send it!
		//
		resp, err := http.Post(
			"http://localhost:8429/api/v1/import/prometheus",
			"",
			buf,
		)
		if err != nil {
			slog.Error("Failed to send samples", "err", err)
			return
		}

		if resp.StatusCode != http.StatusNoContent {
			slog.With("resp.StatusCode", resp.StatusCode).Error("Received unexpected status code")

			return
		}

		slog.With(
			"resp.StatusCode", resp.StatusCode,
		).Info("Pushed samples")

		defer resp.Body.Close()

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
