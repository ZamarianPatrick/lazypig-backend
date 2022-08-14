package graph

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZamarianPatrick/lazypig-backend/graph/model"
	"github.com/ZamarianPatrick/lazypig-backend/sensors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
	"sync"
	"time"
)

type Controller interface {
	DB() *gorm.DB
	PossibleStationPorts() []string
	StationChannel(ctx context.Context) chan *model.Station
	SetMoistureFakeValue(port string, value float64)
	SetWaterLevelFakeValue(value float64)
}

type controller struct {
	db              *gorm.DB
	bus             i2c.BusCloser
	sensorWorker    sensors.Worker
	mutex           sync.RWMutex
	stationChannels map[string]chan *model.Station
	stationSettings *sensors.StationSettings

	moistureFakes  []*sensors.MoistureFake
	waterLevelFake *sensors.WaterFake

	fakeValues bool
	random     *rand.Rand
}

func NewController(fakeValues bool) (Controller, error) {

	basePath := "./"
	settingsFileName := "stationSettings.yml"

	var stationSettings sensors.StationSettings

	if _, err := os.Stat(basePath + settingsFileName); errors.Is(err, os.ErrNotExist) {
		stationSettings = sensors.DefaultStationSettings
		f, err := os.Create(basePath + settingsFileName)
		if err != nil {
			return nil, err
		}

		data, err := yaml.Marshal(stationSettings)
		if err != nil {
			return nil, err
		}

		_, err = f.Write(data)
		if err != nil {
			return nil, err
		}

		f.Close()
	} else {
		data, err := ioutil.ReadFile(basePath + settingsFileName)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(data, &stationSettings)
		if err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(basePath+"db.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	r := db.Exec("PRAGMA foreign_keys = ON", nil)
	if r.Error != nil {
		return nil, err
	}

	db.AutoMigrate(&model.PlantTemplate{})
	db.AutoMigrate(&model.Station{})
	db.AutoMigrate(&model.Plant{})

	var station model.Station
	r = db.First(&station, 1)
	if r.Error != nil {
		station = model.Station{
			Name: "Station 1",
		}
		db.Create(&station)
	}

	randomSource := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randomSource)

	c := controller{
		db:              db,
		stationChannels: make(map[string]chan *model.Station),
		stationSettings: &stationSettings,
		moistureFakes:   make([]*sensors.MoistureFake, 0),
		random:          random,
	}

	fake := sensors.NewWaterFake(100)
	waterFake := fake.(*sensors.WaterFake)

	var sensorWorker sensors.Worker
	if fakeValues {
		sensorWorker =
			sensors.NewWorker().
				Add(waterFake)

	} else {
		_, err = host.Init()
		if err != nil {
			return nil, err
		}

		bus, err := i2creg.Open("1")
		if err != nil {
			return nil, err
		}

		c.bus = bus

		sensorWorker =
			sensors.NewWorker().
				Add(sensors.NewWaterLevel(bus, stationSettings.WaterLevelHighAddress, stationSettings.WaterLevelLowAddress))

	}

	c.waterLevelFake = waterFake
	c.sensorWorker = sensorWorker

	for _, port := range stationSettings.Ports {
		var s sensors.Sensor
		if fakeValues {
			s = sensors.NewMoistureFake(port)
			moistureFake := s.(*sensors.MoistureFake)
			c.moistureFakes = append(c.moistureFakes, moistureFake)
		} else {
			s = sensors.NewMoisture(c.bus, stationSettings.MoistureAddress, port)
		}

		sensorWorker.Add(s)
	}

	pin, err := sensors.GetGPIO(c.stationSettings.PumpGPIO)
	if err != nil {
		return nil, err
	}
	pin.Out(gpio.Low)

	for _, p := range stationSettings.Ports {
		pin, err := sensors.GetGPIO(p.ValveGPIO)
		if err != nil {
			return nil, err
		}
		pin.Out(gpio.Low)
	}

	c.ReadSensors()
	sensorWorker.Start()
	return &c, nil
}

func (c *controller) SetMoistureFakeValue(port string, value float64) {
	for _, m := range c.moistureFakes {
		if m.Port().Port == port {
			m.SetValue(value)
		}
	}
}

func (c *controller) SetWaterLevelFakeValue(value float64) {
	actualVal, _ := c.waterLevelFake.ReadValue()
	for actualVal != value {
		if actualVal > value {
			actualVal -= 1
		} else {
			actualVal += 1
		}
		c.waterLevelFake.SetValue(actualVal)

		timeout := time.Duration(c.random.Intn(200) + 100)
		time.Sleep(time.Millisecond * timeout)
	}
}

type plantState struct {
	moistureValue float64
	pumpRequired  bool
}

func (c *controller) ReadSensors() {
	go func() {
		var ch chan sensors.SensorData
		ch = c.sensorWorker.DataChannel()

		stationID := 1
		lastWaterLevel := -1.0
		lastPlantStates := make(map[string]*plantState)

		for true {
			data := <-ch

			switch data.SensorName {
			case "Water Level":
				if lastWaterLevel < 0 || math.Abs(lastWaterLevel-data.Value) > 1 {
					var station model.Station
					c.db.First(&station, stationID)
					station.WaterLevel = data.Value
					c.db.Save(&station)

					c.mutex.RLock()
					for _, out := range c.stationChannels {
						out <- &station
					}
					c.mutex.RUnlock()

					lastWaterLevel = data.Value
				}

				break

			case "Moisture":
				var lastPlantState *plantState
				var ok bool
				if lastPlantState, ok = lastPlantStates[data.Port.Port]; !ok {
					lastPlantState = &plantState{}
					lastPlantStates[data.Port.Port] = lastPlantState
				}

				if !ok || math.Abs(lastPlantState.moistureValue-data.Value) > 3 {
					var plant model.Plant
					c.db.Preload("Template").Where("port = ? AND station_id = ?", data.Port.Port, stationID).First(&plant)

					if plant.Active {
						if plant.Template.WaterThreshold >= data.Value {
							fmt.Println("Port", data.Port.Port, data.Value)
							if lastWaterLevel > 1 {
								lastPlantState.pumpRequired = true
								pin, err := sensors.GetGPIO(data.Port.ValveGPIO)
								if err != nil {
									fmt.Println(err)
									continue
								}
								pin.Out(gpio.Low)
							} else {
								log.Println("Port", data.Port.Port, "plant is thirsty but no water is there :(")
							}
						} else {
							log.Println("Port", data.Port.Port, "plant not thirsty", data.Value)
							lastPlantState.pumpRequired = false
							pin, err := sensors.GetGPIO(data.Port.ValveGPIO)
							if err != nil {
								fmt.Println(err)
								continue
							}
							pin.Out(gpio.High)
						}
					} else {
						log.Println("Port", data.Port.Port, "plant not active")
					}
					lastPlantState.moistureValue = data.Value
				}

				break
			}

			pumpOn := false
			for _, v := range lastPlantStates {
				if v.pumpRequired {
					pumpOn = true
					break
				}
			}

			if pumpOn {
				pin, err := sensors.GetGPIO(c.stationSettings.PumpGPIO)
				if err != nil {
					fmt.Println(err)
					continue
				}
				pin.Out(gpio.Low)
			} else {
				pin, err := sensors.GetGPIO(c.stationSettings.PumpGPIO)
				if err != nil {
					fmt.Println(err)
					continue
				}
				pin.Out(gpio.High)
			}
		}
	}()
}

func (c *controller) DB() *gorm.DB {
	return c.db
}

func (c *controller) StationChannel(ctx context.Context) chan *model.Station {
	ch := make(chan *model.Station)
	uuid, _ := uuid.NewUUID()

	c.mutex.Lock()
	c.stationChannels[uuid.String()] = ch
	c.mutex.Unlock()

	go func() {
		<-ctx.Done()
		c.mutex.Lock()
		delete(c.stationChannels, uuid.String())
		c.mutex.Unlock()

		log.Println("ws client closed", uuid.String())
	}()

	return ch
}

func (c *controller) PossibleStationPorts() []string {

	portNames := make([]string, len(c.stationSettings.Ports))

	for i, p := range c.stationSettings.Ports {
		portNames[i] = p.Port
	}

	return portNames
}
