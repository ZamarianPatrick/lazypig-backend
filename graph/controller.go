package graph

import (
	"github.com/ZamarianPatrick/lazypig-backend/graph/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Controller interface {
	DB() *gorm.DB
}

type controller struct {
	db *gorm.DB
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

	c := controller{
		db: db,
	}

	var station model.Station
	r = db.First(&station, 1)
	if r.Error != nil {
		station = model.Station{
			Name: "Station 1",
		}
		db.Create(&station)
	}

	return &c, nil
}

func (c *controller) DB() *gorm.DB {
	return c.db
}
