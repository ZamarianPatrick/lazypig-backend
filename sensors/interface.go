package sensors

import (
	"log"
	"time"
)

type Sensor interface {
	Name() string
	ReadValue() (float64, error)
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
				val, err := sensor.ReadValue()
				if err != nil {
					log.Println(err)
					continue
				}
				data := SensorData{
					SensorName: sensor.Name(),
					Value:      val,
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
