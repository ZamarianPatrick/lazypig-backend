package model

type Plant struct {
	ID         uint64        `json:"id" gorm:"primaryKey"`
	StationID  uint64        `json:"stationID"`
	Active     bool          `json:"active"`
	Name       string        `json:"name"`
	Port       string        `json:"port"`
	TemplateID uint64        `json:"-"`
	Template   PlantTemplate `json:"template" gorm:"foreignKey:TemplateID;references:ID"`
}

type PlantTemplate struct {
	ID             uint64  `json:"id" gorm:"primaryKey"`
	Name           string  `json:"name"`
	WaterThreshold float64 `json:"waterThreshold"`
}

type Station struct {
	ID         uint64  `json:"id" gorm:"primaryKey"`
	Name       string  `json:"name"`
	WaterLevel float64 `json:"waterLevel"`
	Plants     []Plant `json:"plants"`
}
