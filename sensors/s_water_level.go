package sensors

import (
	"periph.io/x/conn/v3/i2c"
)

type waterLevel struct {
	bus  i2c.BusCloser
	high i2c.Dev
	low  i2c.Dev
}

func NewWaterLevel(bus i2c.BusCloser, highAddress uint16, lowAddress uint16) Sensor {
	return &waterLevel{
		bus: bus,
		high: i2c.Dev{
			Bus:  bus,
			Addr: highAddress,
		},
		low: i2c.Dev{
			Bus:  bus,
			Addr: lowAddress,
		},
	}
}

func (s *waterLevel) Name() string {
	return "Water Level"
}

func (s *waterLevel) ReadValue() (float64, error) {

	readHigh := make([]byte, 12)
	readLow := make([]byte, 8)

	if err := s.high.Tx(nil, readHigh); err != nil {
		return 0, err
	}

	if err := s.low.Tx(nil, readLow); err != nil {
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
