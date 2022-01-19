package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Filter struct {
	// 0 - filter by departure city name 1 - flight number, 2 - city from, 3 - departure time
	SortBy uint8 `json:"SortBy"`
	// asc, desc
	OrderBy   string `json:"OrderBy"`
	Departure string `json:"Departure"`
}

type Flight struct {
	FlightNum int       `json:"Flight" db:"flight_number" gorm:"index" required:"true"`
	From      string    `json:"From" db:"departured_from" required:"true"`
	Departure time.Time `json:"Departure" db:"departure" required:"true"`
	To        string    `json:"To" db:"arrival_to" required:"true"`
	Arrival   time.Time `json:"Arrival" db:"arrival" required:"true"`
}

type Flights struct {
	gorm.Model
	Flight
}

func (f *Flight) Validate() error {

	// example length 2
	if f.FlightNum < 0 {
		return fmt.Errorf("error with flight num length")
	}

	// example length 2
	if len(f.From) < 2 {
		return fmt.Errorf("error with departured city(from) field length")
	}

	if f.Departure.IsZero() {
		return fmt.Errorf("error with departure time(departure) field length")
	}

	// example length 2
	if len(f.To) < 2 {
		return fmt.Errorf("error with arrival city(to) field length")
	}

	if f.Arrival.IsZero() {
		return fmt.Errorf("error with arrival time(departure) field length")
	}

	return nil
}
