package banners24

import (
	"log/slog"
	"math"

	"github.com/simonvetter/modbus"
)

type Register uint16

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
	RegisterHumidity Register = 0

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
	RegisterTemperatureC Register = 1

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
	RegisterTemperatureF Register = 2

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
	RegisterDewPointC Register = 3

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
	RegisterDewPointF Register = 4

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
	BaudRate Register = 601

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
	Parity Register = 602

	// Address
	//
	// Default: 1
	//
	// IO Range:
	// 1-254
	//
	// Read Write
	Address Register = 603

	// Restore Factory Configuration
	//
	// Default: 1
	//
	// IO Range:
	// 0 = No Operation
	// 1 = Restore
	//
	// Write only
	RestoreFactoryConfig Register = 605

	//
	// Device Information
	//

)

type SensorData struct {
	Humidity     float64
	TemperatureF float64
	TemperatureC float64
	DewPointC    float64
	DewPointF    float64
}

func ReadSensorData(client *modbus.ModbusClient) (*SensorData, error) {
	humidity, err := readHumidity(client, RegisterHumidity)
	if err != nil {
		return nil, err
	}

	tempC, err := readTemperatureRegister(client, RegisterTemperatureC)
	if err != nil {
		return nil, err
	}

	tempF, err := readTemperatureRegister(client, RegisterTemperatureF)
	if err != nil {
		return nil, err
	}

	dewPointC, err := readDewPointRegister(client, RegisterDewPointC)
	if err != nil {
		return nil, err
	}

	dewPointF, err := readDewPointRegister(client, RegisterDewPointF)
	if err != nil {
		return nil, err
	}

	// slog.With(
	// 	"humidity", humidity,
	// 	"tempF", tempF,
	// 	"tempC", tempC,
	// 	"dewPointC", dewPointC,
	// 	"dewPointF", dewPointF,
	// ).Debug("Sensor Readings")

	return &SensorData{
		Humidity:     humidity,
		TemperatureF: tempF,
		TemperatureC: tempC,
		DewPointC:    dewPointC,
		DewPointF:    dewPointF,
	}, nil
}

func readRegister(client *modbus.ModbusClient, reg Register) (uint16, error) {
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

func readTemperatureRegister(client *modbus.ModbusClient, reg Register) (float64, error) {
	raw, err := readRegister(client, reg)
	if err != nil {
		return math.NaN(), err
	}

	return float64(int16(raw)) / 20.0, nil
}

func readDewPointRegister(client *modbus.ModbusClient, reg Register) (float64, error) {
	raw, err := readRegister(client, reg)
	if err != nil {
		return math.NaN(), err
	}

	return float64(int16(raw)) / 100.0, nil
}

func readHumidity(client *modbus.ModbusClient, reg Register) (float64, error) {
	raw, err := readRegister(client, reg)
	if err != nil {
		return math.NaN(), err
	}

	return float64(raw) / 100.0, nil
}

func readString(client *modbus.ModbusClient, start, length uint16) (string, error) {
	rawBytes, err := client.ReadRawBytes(start, length, modbus.HOLDING_REGISTER)
	if err != nil {
		return "", err
	}

	slog.With(
		"start", start,
		"length", length,
		"rawBytes", rawBytes,
	).Debug("Read Register")

	return string(rawBytes), err
}

func readBannerName(client *modbus.ModbusClient) (string, error) {
	return readString(client, 605, 18)
}
