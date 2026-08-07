package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/nowireless/air/banner-s24/banners24"
	"github.com/simonvetter/modbus"
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

	readings, err := banners24.ReadSensorData(client)
	if err != nil {
		slog.Error("Failed to read S24 sensor data", "err", err)
		os.Exit(1)
	}
	slog.With(
		"humidity", readings.Humidity,
		"tempF", readings.TemperatureF,
		"tempC", readings.TemperatureC,
		"dewPointC", readings.DewPointC,
		"dewPointF", readings.DewPointF,
	).Info("Readings")
}
