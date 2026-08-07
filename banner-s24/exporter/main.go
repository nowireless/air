package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/nowireless/air/banner-s24/banners24"
	"github.com/simonvetter/modbus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	humidityGauge := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "humidity_rh",
		Help: "Humidity (%rh)",
	})
	tempCGauge := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "temperature_c",
		Help: "Temperature Celsius",
	})
	tempFGauge := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "temperature_f",
		Help: "Temperature Fahrenheit",
	})
	dewPointCGauge := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "dew_point_c",
		Help: "Dew Point Celsius",
	})
	dewPointFGauge := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Name: "dew_point_f",
		Help: "Dew Point Fahrenheit",
	})

	slog.Info("Starting HTTP server")
	reg.MustRegister()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":2112", nil)

	// readings, err := banners24.ReadSensorData(client)
	// if err != nil {
	// 	slog.Error("Failed to read S24 sensor data", "err", err)
	// 	os.Exit(1)
	// }
	// slog.With(
	// 	"humidity", readings.Humidity,
	// 	"tempF", readings.TemperatureF,
	// 	"tempC", readings.TemperatureC,
	// 	"dewPointC", readings.DewPointC,
	// 	"dewPointF", readings.DewPointF,
	// ).Info("Readings")

	for {
		readings, err := banners24.ReadSensorData(client)
		if err != nil {
			continue
		}

		humidityGauge.Set(readings.Humidity)
		tempCGauge.Set(readings.TemperatureC)
		tempFGauge.Set(readings.TemperatureF)
		dewPointCGauge.Set(readings.DewPointC)
		dewPointFGauge.Set(readings.DewPointF)

		slog.With(
			"humidity", readings.Humidity,
			"tempF", readings.TemperatureF,
			"tempC", readings.TemperatureC,
			"dewPointC", readings.DewPointC,
			"dewPointF", readings.DewPointF,
		).Info("Readings")

		time.Sleep(time.Second)
	}

}
