package sensors

import (
	"fmt"
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3/rpi"
	"time"
)

type StationSettings struct {
	GroveBus string `yaml:"groveBus"`

	WaterLevelHighAddress uint16 `yaml:"waterLevelHighAddress"`
	WaterLevelLowAddress  uint16 `yaml:"waterLevelLowAddress"`
	MoistureAddress       uint16 `yaml:"moistureAddress"`
	PumpGPIO              int    `yaml:"pumpGPIO"`

	Ports []PortSetting `yaml:"ports"`
}

type PortSetting struct {
	Port            string `yaml:"port"`
	MoistureChannel byte   `yaml:"moistureChannel"`
	ValveGPIO       int    `yaml:"valveGPIO"`
}

var (
	DefaultStationSettings = StationSettings{
		GroveBus:              "1",
		WaterLevelHighAddress: 0x78,
		WaterLevelLowAddress:  0x77,
		MoistureAddress:       0x08,
		PumpGPIO:              23,
		Ports: []PortSetting{
			{
				Port:            "A",
				MoistureChannel: 0x0,
				ValveGPIO:       24,
			},
			{
				Port:            "B",
				MoistureChannel: 0x02,
				ValveGPIO:       25,
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

func GetGPIO(gpio int) (gpio.PinIO, error) {
	switch gpio {
	case 2:
		return rpi.P1_3, nil
	case 3:
		return rpi.P1_5, nil
	case 4:
		return rpi.P1_7, nil
	case 5:
		return rpi.P1_29, nil
	case 6:
		return rpi.P1_31, nil
	case 7:
		return rpi.P1_26, nil
	case 8:
		return rpi.P1_24, nil
	case 9:
		return rpi.P1_21, nil
	case 10:
		return rpi.P1_19, nil
	case 11:
		return rpi.P1_23, nil
	case 12:
		return rpi.P1_32, nil
	case 13:
		return rpi.P1_33, nil
	case 16:
		return rpi.P1_36, nil
	case 17:
		return rpi.P1_11, nil
	case 18:
		return rpi.P1_12, nil
	case 19:
		return rpi.P1_35, nil
	case 20:
		return rpi.P1_38, nil
	case 21:
		return rpi.P1_40, nil
	case 22:
		return rpi.P1_15, nil
	case 23:
		return rpi.P1_16, nil
	case 24:
		return rpi.P1_18, nil
	case 25:
		return rpi.P1_22, nil
	case 26:
		return rpi.P1_37, nil
	case 27:
		return rpi.P1_13, nil
	default:
		return nil, fmt.Errorf("gpio %d cant found", gpio)
	}
}
