package sensors

import (
	"log"
	"time"
)

type StationSettings struct {
	GroveBus string `yaml:"groveBus"`

	WaterLevelHighAddress uint16 `yaml:"waterLevelHighAddress"`
	WaterLevelLowAddress  uint16 `yaml:"waterLevelLowAddress"`
	MoistureAddress       uint16 `yaml:"moistureAddress"`

	Ports []PortSetting `yaml:"ports"`
}

type PortSetting struct {
	Port            string `yaml:"port"`
	MoistureChannel byte   `yaml:"moistureChannel"`
}

var (
	DefaultStationSettings = StationSettings{
		GroveBus:              "1",
		WaterLevelHighAddress: 0x78,
		WaterLevelLowAddress:  0x77,
		MoistureAddress:       0x08,
		Ports: []PortSetting{
			{
				Port:            "A",
				MoistureChannel: 0x0,
			},
			{
				Port:            "B",
				MoistureChannel: 0x02,
			},
			{
				Port:            "C",
				MoistureChannel: 0x04,
			},
			{
				Port:            "D",
				MoistureChannel: 0x06,
			},
		},
	}
)

type Sensor interface {
	Name() string
	ReadValue() (float64, error)
}

type PortSensor interface {
	Port() PortSetting
}

type Worker interface {
	Add(sensor Sensor) Worker
	Start()
	Stop()
	DataChannel() chan SensorData
}

type SensorData struct {
	SensorName string
	Value      float64
	Port       PortSetting
}

type sensorWorker struct {
	running      bool
	sensors      []Sensor
	valueChannel chan SensorData
}

func NewWorker() Worker {
	return &sensorWorker{
		running:      false,
		valueChannel: make(chan SensorData),
	}
}

func (sw *sensorWorker) Add(sensor Sensor) Worker {
	sw.sensors = append(sw.sensors, sensor)
	return sw
}

func (sw *sensorWorker) Start() {
	sw.running = true
	go func() {
		for sw.running {
			for _, sensor := range sw.sensors {
				portSensor, ok := sensor.(PortSensor)

				val, err := sensor.ReadValue()
				if err != nil {
					if !ok {
						log.Println(err)
					}
					continue
				}

				data := SensorData{
					SensorName: sensor.Name(),
					Value:      val,
				}

				if ok {
					data.Port = portSensor.Port()
				}

				sw.valueChannel <- data
			}
			time.Sleep(time.Second)
		}
	}()
}

func (sw *sensorWorker) Stop() {
	sw.running = false
}

func (sw *sensorWorker) DataChannel() chan SensorData {
	return sw.valueChannel
}
