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

type BannerS24Register uint16

const (

	//
	// Sensor Data
	//

	// Humidity (%RH)
	// Sensor Address: 40001
	// I/O Range
	// 		Min Value: 0
	// 		Max Value: 100%
	// Holding Register Registration
	// 		Min (Dec): 0
	// 		Max (Dec): 10000
	//
	// Readonly
	//
	// The humidity = (Holding register value) ÷ 100
	BannerS24RegisterHumidity BannerS24Register = 0

	// Temperature (°C)
	// Sensor Address: 40002
	// I/O Range
	// 		Min Value: -1638.4
	// 		Max Value: 1638.3
	// Holding Register Registration
	// 		Min (Dec): -32768
	// 		Max (Dec): 32767
	//
	// Readonly
	//
	// The temperature = (Modbus register value) ÷ 20
	BannerS24RegisterTemperatureC BannerS24Register = 1

	// Temperature (°F)
	// Sensor Address: 40003
	// I/O Range
	// 		Min Value: -1638.4
	// 		Max Value: 1638.3
	// Holding Register Registration
	// 		Min (Dec): -32768
	// 		Max (Dec): 32767
	//
	// Readonly
	//
	// The temperature = (Modbus register value) ÷ 20
	BannerS24RegisterTemperatureF BannerS24Register = 2

	// Dew Point (°C)
	// Sensor Address: 40004
	// I/O Range
	// 		Min Value: -1638.4
	// 		Max Value: 1638.3
	// Holding Register Registration
	// 		Min (Dec): -32768
	// 		Max (Dec): 32767
	//
	// Readonly
	//
	// The dew point = (Holding register value) ÷ 100
	BannerS24RegisterDewPointC BannerS24Register = 3

	// Dew Point (°F)
	// Sensor Address: 40005
	// I/O Range
	// 		Min Value: -1638.4
	// 		Max Value: 1638.3
	// Holding Register Registration
	// 		Min (Dec): -32768
	// 		Max (Dec): 32767
	//
	// Readonly
	//
	// The dew point = (Holding register value) ÷ 100
	BannerS24RegisterDewPointF BannerS24Register = 4

	//
	// COMs setting
	//

	// Baud rate
	//
	// Default: 1
	//
	// IO Range:
	// 0 = 9600
	// 1 = 19200
	// 2 = 38400
	//
	// Read Write
	BannerS24BaudRate BannerS24Register = 601

	// Parity
	//
	// Default: 0
	//
	// IO Range:
	// 0 = None
	// 1 = Odd
	// 2 = Even
	//
	// Read Write
	BannerS24Parity BannerS24Register = 602

	// Address
	//
	// Default: 1
	//
	// IO Range:
	// 1-254
	//
	// Read Write
	BannerS24Address BannerS24Register = 603

	// Restore Factory Configuration
	//
	// Default: 1
	//
	// IO Range:
	// 0 = No Operation
	// 1 = Restore
	//
	// Write only
	BannerS24RestoreFactoryConfig BannerS24Register = 605

	//
	// Device Information
	//

)

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

func readRegister(client *modbus.ModbusClient, reg BannerS24Register) (uint16, error) {
	rawValue, err := client.ReadRegister(
		uint16(reg),
		modbus.HOLDING_REGISTER,
	)

	if err != nil {
		return 0, err
	}

	slog.With(
		"reg", reg,
		"rawValue", rawValue,
	).Debug("Read Register")

	return rawValue, nil
}
