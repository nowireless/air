package main

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/nowireless/air/banner-s24/banners24"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/simonvetter/modbus"
)

type Collector struct {
	// client *modbus.ModbusClient
	mu sync.Mutex

	humidity     *prometheus.Desc
	temperature  *prometheus.Desc
	dewPoint     *prometheus.Desc
	up           *prometheus.Desc
	readDuration *prometheus.Desc
}

func NewCollector() *Collector {
	return &Collector{
		// client: client,

		humidity: prometheus.NewDesc(
			"banner_s24_relative_humidity_ratio",
			"Relative humidity reported by the Banner S24, from 0 to 1.",
			nil,
			nil,
		),

		temperature: prometheus.NewDesc(
			"banner_s24_temperature_celsius",
			"Air temperature reported by the Banner S24.",
			nil,
			nil,
		),

		dewPoint: prometheus.NewDesc(
			"banner_s24_dew_point_celsius",
			"Dew-point temperature reported by the Banner S24.",
			nil,
			nil,
		),

		up: prometheus.NewDesc(
			"banner_s24_up",
			"Whether the latest Banner S24 query succeeded.",
			nil,
			nil,
		),

		readDuration: prometheus.NewDesc(
			"banner_s24_read_duration_seconds",
			"Time required to query the Banner S24.",
			nil,
			nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.humidity
	ch <- c.temperature
	ch <- c.dewPoint
	ch <- c.up
	ch <- c.readDuration
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Avoid concurrent Modbus operations if scrapes overlap.
	c.mu.Lock()
	defer c.mu.Unlock()

	start := time.Now()

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://192.168.2.188:502",
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

		ch <- prometheus.MustNewConstMetric(
			c.up,
			prometheus.GaugeValue,
			0,
		)

		// Do not emit stale or invented sensor values.
		return
	}

	readings, err := banners24.ReadSensorData(client)

	ch <- prometheus.MustNewConstMetric(
		c.readDuration,
		prometheus.GaugeValue,
		time.Since(start).Seconds(),
	)

	if err != nil {
		slog.Error("failed to read Banner S24", "err", err)

		ch <- prometheus.MustNewConstMetric(
			c.up,
			prometheus.GaugeValue,
			0,
		)

		// Do not emit stale or invented sensor values.
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.up,
		prometheus.GaugeValue,
		1,
	)

	// Prometheus convention generally represents ratios as 0–1.
	ch <- prometheus.MustNewConstMetric(
		c.humidity,
		prometheus.GaugeValue,
		readings.Humidity/100,
	)

	ch <- prometheus.MustNewConstMetric(
		c.temperature,
		prometheus.GaugeValue,
		readings.TemperatureC,
	)

	ch <- prometheus.MustNewConstMetric(
		c.dewPoint,
		prometheus.GaugeValue,
		readings.DewPointC,
	)

	slog.Info(
		"Read Banner S24",
		"humidity", readings.Humidity,
		"temperature_celsius", readings.TemperatureC,
		"dew_point_celsius", readings.DewPointC,
	)
}
