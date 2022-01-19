package repo

import (
	"t3/models"

	"gorm.io/gorm"
)

type Repo struct {
	Db *gorm.DB
}

func (r *Repo) GellAllFlights() ([]models.Flight, error) {
	flights := []models.Flight{}

	if err := r.Db.Table("flights").Find(&flights).Error; err != nil {
		return flights, err
	}

	return flights, nil
}

func (r *Repo) GellAllFlightsByCity(city string) ([]models.Flight, error) {
	flights := []models.Flight{}

	if err := r.Db.Table("flights").Where("departure = ?", city).Find(&flights).Error; err != nil {
		return flights, err
	}

	return flights, nil
}

func (r *Repo) Insert(flights []models.Flights) ([]models.Flights, error) {

	if err := r.Db.Create(&flights).Error; err != nil {
		return flights, err
	}

	return flights, nil
}

func (r *Repo) Update(flights []models.Flights) ([]models.Flights, error) {

	if err := r.Db.Save(&flights).Error; err != nil {
		return flights, err
	}

	return flights, nil
}
