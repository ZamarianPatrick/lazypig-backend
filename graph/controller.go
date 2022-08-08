package graph

import (
	"context"
	"github.com/ZamarianPatrick/lazypig-backend/graph/model"
	"github.com/ZamarianPatrick/lazypig-backend/sensors"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
	"sync"
)

type Controller interface {
	DB() *gorm.DB
	PossibleStationPorts() []string
	StationChannel(ctx context.Context) chan *model.Station
}

type controller struct {
	db              *gorm.DB
	sensorWorker    sensors.Worker
	mutex           sync.RWMutex
	stationChannels map[string]chan *model.Station
}

func NewController() (Controller, error) {
	db, err := gorm.Open(sqlite.Open("/root/db.sqlite"), &gorm.Config{
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

	_, err = host.Init()
	if err != nil {
		return nil, err
	}

	bus, err := i2creg.Open("1")
	if err != nil {
		return nil, err
	}

	sensorWorker := sensors.NewWorker().
		Add(sensors.NewWaterLevel(bus))

	c := controller{
		db:              db,
		sensorWorker:    sensorWorker,
		stationChannels: make(map[string]chan *model.Station),
	}

	c.ReadSensors()
	sensorWorker.Start()
	return &c, nil
}

func (c *controller) ReadSensors() {
	go func() {
		var ch chan sensors.SensorData
		ch = c.sensorWorker.DataChannel()

		stationID := 1
		lastWaterLevel := -1.0

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
					break
				}
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
	return []string{
		"A",
		"B",
		"C",
		"D",
	}
}
