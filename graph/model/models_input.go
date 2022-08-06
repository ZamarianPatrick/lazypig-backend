package model

type DeletePlantInput struct {
	StationID uint64 `json:"stationID"`
	PlantID   uint64 `json:"plantID"`
}

type PlantInput struct {
	TemplateID uint64 `json:"templateID"`
	Active     bool   `json:"active"`
	Name       string `json:"name"`
	Port       string `json:"port"`
}

type PlantTemplateInput struct {
	Name           string  `json:"name"`
	WaterThreshold float64 `json:"waterThreshold"`
}

type StationInput struct {
	Name string `json:"name"`
}
