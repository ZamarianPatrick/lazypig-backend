package sensors

type WaterFake struct {
	value float64
}

func NewWaterFake(value float64) Sensor {
	return &WaterFake{
		value: value,
	}
}

func (s *WaterFake) Name() string {
	return "Water Level"
}

func (s *WaterFake) SetValue(val float64) {
	s.value = val
}

func (s *WaterFake) ReadValue() (float64, error) {
	return s.value, nil
}
