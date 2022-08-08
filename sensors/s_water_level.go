package sensors

import (
	"periph.io/x/conn/v3/i2c"
)

type waterLevel struct {
	bus i2c.BusCloser
}

func NewWaterLevel(bus i2c.BusCloser) Sensor {
	return &waterLevel{
		bus: bus,
	}
}

func (s *waterLevel) Name() string {
	return "Water Level"
}

func (s *waterLevel) ReadValue() (float64, error) {

	high := i2c.Dev{
		Bus:  s.bus,
		Addr: 0x78,
	}

	low := i2c.Dev{
		Bus:  s.bus,
		Addr: 0x77,
	}

	write := []byte{0x0}
	readHigh := make([]byte, 12)
	readLow := make([]byte, 8)

	if err := high.Tx(write, readHigh); err != nil {
		return 0, err
	}

	if err := low.Tx(write, readLow); err != nil {
		return 0, err
	}

	threshold := byte(100)
	sensorValueMin := byte(250)
	sensorValueMax := byte(255)

	touchValue := 0
	trigSection := 0
	lowCount := 0
	highCount := 0

	for i := 0; i < 8; i++ {
		if readLow[i] >= sensorValueMin && readLow[i] <= sensorValueMax {
			lowCount++
		}
	}

	for i := 0; i < 12; i++ {
		if readHigh[i] >= sensorValueMin && readHigh[i] <= sensorValueMax {
			highCount++
		}
	}

	for i := 0; i < 8; i++ {
		if readLow[i] > threshold {
			touchValue |= 1 << i
		}
	}

	for i := 0; i < 12; i++ {
		if readHigh[i] > threshold {
			touchValue |= 1 << (8 + i)
		}
	}

	for touchValue&0x01 != 0 {
		trigSection++
		touchValue >>= 1
	}

	return float64(trigSection * 5), nil
}
