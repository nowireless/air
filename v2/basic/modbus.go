package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/simonvetter/modbus"
)

type ModBusSession struct {
	Gateway string

	Client *modbus.ModbusClient
}

func Connect(gateway string) (*ModBusSession, error) {
	m := ModBusSession{
		Gateway: gateway,
	}

	if err := m.Connect(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *ModBusSession) Connect() error {
	if len(m.Gateway) == 0 {
		return fmt.Errorf("modbus TCP gateway not set")
	}

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     m.Gateway,
		Timeout: 1 * time.Second,
	})

	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to create modbus client"),
			err,
		)
	}

	if err := client.Open(); err != nil {
		return errors.Join(
			fmt.Errorf("failed to open modbus client"),
			err,
		)
	}

	m.Client = client

	return nil
}

func (m *ModBusSession) Close() error {
	if m.Client == nil {
		return fmt.Errorf("modbus client is nil")
	}

	return m.Client.Close()
}

func (m *ModBusSession) SetUnitId(id uint8) (err error) {
	if m.Client == nil {
		return fmt.Errorf("modbus client is nil")
	}

	return m.Client.SetUnitId(id)
}

func (m *ModBusSession) ReadRegister(reg uint16, regType modbus.RegType) (uint16, error) {
	if m == nil {
		return 0, fmt.Errorf("modbus session is nil")
	}

	if m.Client == nil {
		return 0, fmt.Errorf("modbus client is nil")
	}

	rawValue, err := m.Client.ReadRegister(
		uint16(reg),
		regType,
	)

	if err != nil {
		return 0, err
	}

	slog.With(
		"reg", reg,
		"regType", regType,
		"rawValue", rawValue,
	).Debug("Read Register")

	return rawValue, nil
}

func (m *ModBusSession) ReadRegisters(startingReg uint16, quantity uint16, regType modbus.RegType) ([]uint16, error) {
	if m == nil {
		return nil, fmt.Errorf("modbus session is nil")
	}

	if m.Client == nil {
		return nil, fmt.Errorf("modbus client is nil")
	}

	rawValues, err := m.Client.ReadRegisters(startingReg, quantity, regType)
	if err != nil {
		return nil, err
	}

	slog.With(
		"startingReg", startingReg,
		"regType", regType,
		"rawValues", rawValues,
	).Debug("Read Registers")

	return rawValues, nil
}

func (m *ModBusSession) ReadString(start, length uint16) (string, error) {
	if m.Client == nil {
		return "", fmt.Errorf("modbus client is nil")
	}

	rawBytes, err := m.Client.ReadRawBytes(start, length, modbus.HOLDING_REGISTER)
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
