package sensors

type MoistureFake struct {
	setting PortSetting
	value   float64
}

func NewMoistureFake(setting PortSetting) Sensor {
	return &MoistureFake{
		setting: setting,
		value:   100,
	}
}

func (s *MoistureFake) Name() string {
	return "Moisture"
}

func (s *MoistureFake) Port() PortSetting {
	return s.setting
}

func (s *MoistureFake) SetValue(val float64) {
	s.value = val
}

func (s *MoistureFake) ReadValue() (float64, error) {
	return s.value, nil
}
