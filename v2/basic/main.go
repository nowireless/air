package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simonvetter/modbus"
)

type Unit struct {
	Name string

	ID uint8

	Registers      []*Register
	RegisterRanges []*RegisterRange
}

type Register struct {
	Name     string
	Register uint16
	Type     modbus.RegType

	Value    any
	RawValue uint16

	Transform func(rawValue uint16) (any, error)
}

type RegisterRange struct {
	Type modbus.RegType

	StartRegister uint16

	Registers []*RegisterRangeElement
}

type RegisterRangeElement struct {
	Name string

	Value    any
	RawValue uint16

	Transform func(rawValue uint16) (any, error)
}

// func (r *RegisterValue[T]) Process(rawValue uint16) error {
// 	value, err := r.Transform(rawValue)
// 	if err != nil {
// 		return err
// 	}

// 	r.Value = value
// 	return nil
// }

// type RegisterGetter interface {
// 	Process(rawValue uint16) error
// }

func main() {
	// Set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewTextHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	slog.Info("Laser Compressed Air Filtration collector")

	testReadAll()
}

func testReadAll() {
	slog.Info("Performing test read of all sensors")

	m, err := Connect("tcp://172.30.0.140:502")
	defer m.Close()
	if err != nil {
		slog.Error("Failed to create modbus client", "err", err)

		// Do not emit stale or invented sensor values.
		return
	}

	//
	// Gather all data
	//
	transformBannerHumidity := func(rawValue uint16) (any, error) {
		return float64(rawValue) / 100.0, nil
	}

	transformBannerTemp := func(rawValue uint16) (any, error) {
		return float64(int16(rawValue)) / 20.0, nil
	}

	transformBannerDewPoint := func(rawValue uint16) (any, error) {
		return float64(int16(rawValue)) / 100.0, nil
	}

	transformWaveSharePressureSensor := func(rawValue uint16) (any, error) {
		voltage := float64(rawValue / 1000.0)

		vcc := 5.0
		return float64(250.0*(float64(voltage)/vcc) - 25.0), nil
	}

	transformIFMSD6001Flow := func(rawValue uint16) (any, error) {
		// ASP and AEP need to make sensor config
		asp := 0.0
		aep := 44.1
		return (asp + (float64(rawValue)-4000.0)/16000.0*(aep-asp)), nil
	}

	units := []*Unit{
		{
			Name: "Banner S24",
			ID:   1,
			// Registers: []*Register{
			// 	{Name: "input_humidity", Register: 0, Type: modbus.HOLDING_REGISTER, Transform: func(rawValue uint16) (any, error) { return float64(rawValue) / 100.0, nil }},
			// 	{Name: "input_temperature_c", Register: 1, Type: modbus.HOLDING_REGISTER, Transform: transformBannerTemp},
			// 	{Name: "input_temperature_f", Register: 2, Type: modbus.HOLDING_REGISTER, Transform: transformBannerTemp},
			// 	{Name: "input_pressure_dewpoint_c", Register: 3, Type: modbus.HOLDING_REGISTER, Transform: transformBannerDewPoint},
			// 	{Name: "input_pressure_dewpoint_f", Register: 4, Type: modbus.HOLDING_REGISTER, Transform: transformBannerDewPoint},
			// },
			RegisterRanges: []*RegisterRange{
				{
					Type:          modbus.HOLDING_REGISTER,
					StartRegister: 0,
					Registers: []*RegisterRangeElement{
						{Name: "input_humidity", Transform: transformBannerHumidity},
						{Name: "input_temperature_c", Transform: transformBannerTemp},
						{Name: "input_temperature_f", Transform: transformBannerTemp},
						{Name: "input_pressure_dewpoint_c", Transform: transformBannerDewPoint},
						{Name: "input_pressure_dewpoint_f", Transform: transformBannerDewPoint},
					},
				},
			},
		},
		{
			Name: "Waveshare Analog 8ch",
			ID:   3,
			// Registers: []*Register{
			// 	{Name: "input_pressure_psi", Register: 0, Type: modbus.INPUT_REGISTER, Transform: transformWaveSharePressureSensor},
			// 	{Name: "input_flow_sfcm", Register: 7, Type: modbus.INPUT_REGISTER, Transform: transformIFMSD6001Flow},
			// },
			RegisterRanges: []*RegisterRange{
				{
					Type:          modbus.INPUT_REGISTER,
					StartRegister: 0,
					Registers: []*RegisterRangeElement{
						{Name: "input_pressure_psi", Transform: transformWaveSharePressureSensor},
						nil,
						nil,
						nil,
						nil,
						nil,
						nil,
						{Name: "input_flow_sfcm", Transform: transformIFMSD6001Flow},
					},
				},
			},
		},
	}

	for _, unit := range units {
		logCtx := slog.With("unit.id", unit.ID, "unit.name", unit.Name)
		logCtx.Info("Fetching data")

		if err := m.SetUnitId(unit.ID); err != nil {
			logCtx.Error("Failed to set modbus unit id", "err", err)
			return
		}

		for _, register := range unit.Registers {
			logCtx := logCtx.With(
				"register.Name", register.Name,
				"register.Register", register.Register,
			)
			logCtx.Debug("Reading Register")

			rawValue, err := m.ReadRegister(register.Register, register.Type)
			if err != nil {
				logCtx.Error("Failed to read register", "err", err)
				return
			}

			register.RawValue = rawValue
			register.Value, err = register.Transform(rawValue)
			if err != nil {
				logCtx.Error("Failed to transform raw value", "err", err)
				return
			}

			logCtx.Debug("Read Register",
				"rawValue", rawValue,
				"value", register.Value,
			)
		}

		for _, registerRange := range unit.RegisterRanges {
			logCtx := logCtx.With(
				"registerRange.Type", registerRange.Type,
				"registerRange.StartRegister", registerRange.StartRegister,
			)
			logCtx.Debug("Reading Register range")

			rawValues, err := m.ReadRegisters(
				registerRange.StartRegister,
				uint16(len(registerRange.Registers)),
				registerRange.Type,
			)

			if err != nil {
				logCtx.Error("Failed to read register", "err", err)
				return
			}

			if len(rawValues) != len(registerRange.Registers) {
				logCtx.With(
					"len(rawValues)", len(rawValues),
					"len(registerRange.Registers)", len(registerRange.Registers),
				).Error("Unexpected number of raw values retrieved from device")
				return
			}

			for i, register := range registerRange.Registers {
				if register == nil {
					logCtx.With(
						"registerOffset", i,
					).Debug("Skipping register")
					continue
				}

				register.RawValue = rawValues[i]
				register.Value, err = register.Transform(register.RawValue)
				if err != nil {
					logCtx.Error("Failed to transform raw value", "err", err)
					return
				}

				logCtx.Debug("Read Register",
					"name", register.Name,
					"rawValue", register.RawValue,
					"value", register.Value,
				)
			}

		}
		logCtx.Info("Done")
	}

	//
	// Transform
	//

	//
	// Display
	//

}
