package models

import "fmt"

type Filter struct {
	// 0 - filter by departure city name 1 - flight number, 2 - city from, 3 - departure time
	SortBy uint8 `json:"SortBy,omitempty"`
	// OrderBy false - asc, true - desc
	OrderBy   bool   `json:"OrderBy,omitempty"`
	Departure string `json:"Departure,omitempty"`
}

func (f *Filter) Validate() error {

	if f.SortBy > 3 {
		return fmt.Errorf("bad request: wrong sort value")
	}

	if len(f.Departure) < 2 {
		return fmt.Errorf("there is no city name with length of 2 or less")
	}

	return nil
}
