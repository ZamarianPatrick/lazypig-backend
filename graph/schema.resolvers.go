package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/ZamarianPatrick/lazypig-backend/graph/generated"
	"github.com/ZamarianPatrick/lazypig-backend/graph/model"
	"gorm.io/gorm/clause"
)

func (r *mutationResolver) CreatePlantTemplate(ctx context.Context, input model.PlantTemplateInput) (*model.PlantTemplate, error) {
	template := &model.PlantTemplate{
		Name:           input.Name,
		WaterThreshold: input.WaterThreshold,
	}

	r.controller.DB().Create(template)
	return template, nil
}

func (r *mutationResolver) UpdatePlantTemplate(ctx context.Context, id uint64, input model.PlantTemplateInput) (*model.PlantTemplate, error) {
	template := &model.PlantTemplate{
		ID:             id,
		Name:           input.Name,
		WaterThreshold: input.WaterThreshold,
	}

	r.controller.DB().Save(template)
	return template, nil
}

func (r *mutationResolver) DeletePlantTemplate(ctx context.Context, ids []*uint64) ([]*uint64, error) {
	res := r.controller.DB().Delete(&model.PlantTemplate{}, ids)

	if res.Error != nil {
		return nil, res.Error
	}

	return ids, nil
}

func (r *mutationResolver) CreatePlant(ctx context.Context, stationID uint64, input model.PlantInput) (*model.Plant, error) {
	var template model.PlantTemplate
	r.controller.DB().First(&template, input.TemplateID)

	plant := &model.Plant{
		StationID: stationID,
		Name:      input.Name,
		Active:    input.Active,
		Port:      input.Port,
		Template:  template,
	}

	r.controller.DB().Create(plant)
	return plant, nil
}

func (r *mutationResolver) UpdatePlant(ctx context.Context, id uint64, stationID uint64, input model.PlantInput) (*model.Plant, error) {
	var template model.PlantTemplate
	r.controller.DB().First(&template, input.TemplateID)

	plant := &model.Plant{
		ID:        id,
		StationID: stationID,
		Name:      input.Name,
		Active:    input.Active,
		Port:      input.Port,
		Template:  template,
	}

	r.controller.DB().Create(plant)
	return plant, nil
}

func (r *mutationResolver) DeletePlant(ctx context.Context, id uint64) (bool, error) {
	r.controller.DB().Delete(&model.Plant{}, id)
	return true, nil
}

func (r *mutationResolver) UpdateStation(ctx context.Context, id uint64, input model.StationInput) (*model.Station, error) {
	var station model.Station
	r.controller.DB().First(&station, id)

	station.Name = input.Name
	r.controller.DB().Save(&station)

	return &station, nil
}

func (r *queryResolver) Plant(ctx context.Context, id uint64) (*model.Plant, error) {
	var plant model.Plant
	r.controller.DB().Preload("Template").First(&plant, id)

	return &plant, nil
}

func (r *queryResolver) Stations(ctx context.Context) ([]*model.Station, error) {
	var stations []*model.Station
	r.controller.DB().Preload("Plants.Template").Preload(clause.Associations).Find(&stations)
	return stations, nil
}

func (r *queryResolver) Templates(ctx context.Context) ([]*model.PlantTemplate, error) {
	var templates []*model.PlantTemplate
	r.controller.DB().Find(&templates)
	return templates, nil
}

func (r *subscriptionResolver) Station(ctx context.Context, input uint64) (<-chan []*model.Station, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
