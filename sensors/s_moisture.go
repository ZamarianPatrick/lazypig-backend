package sensors

import (
	"encoding/binary"
	"periph.io/x/conn/v3/i2c"
)

type moisture struct {
	bus     i2c.BusCloser
	dev     i2c.Dev
	setting PortSetting
}

func NewMoisture(bus i2c.BusCloser, address uint16, setting PortSetting) Sensor {
	return &moisture{
		bus: bus,
		dev: i2c.Dev{
			Bus:  bus,
			Addr: address,
		},
		setting: setting,
	}
}

func (s *moisture) Name() string {
	return "Moisture"
}

func (s *moisture) Port() PortSetting {
	return s.setting
}

func (s *moisture) ReadValue() (float64, error) {

	write := []byte{0x20 + s.setting.MoistureChannel}
	read := make([]byte, 2)
	if err := s.dev.Tx(write, read); err != nil {
		return 0, err
	}

	val := float64(binary.LittleEndian.Uint16(read))

	if val <= 1000 {
		return 100, nil
	}

	if val >= 2000 {
		return 0, nil
	}

	diff := -(val - 2000)
	percentage := diff / 10

	return percentage, nil
}
