package router

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	repo "t3/api/sql"
	"t3/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Srv struct {
	DB                    repo.Repo
	MFlights              map[int]models.Flight // key = flight num, value = index in notsorted
	UnSorted              []models.Flight
	SortedByFlightNum     []models.Flight
	SortedByCity          []models.Flight
	SortedByDepartureTime []models.Flight

	InsertDataCh chan []models.Flight

	Mu sync.RWMutex

	Hub *Hub

	ServerInitialized bool
}

func Route(srv *Srv) {
	r := gin.Default()

	hub := NewHub()
	go hub.Pool()
	go hub.Run()
	srv.Hub = hub

	r.POST("/flights", srv.Flights)
	r.POST("/insert", srv.InsertData)
	r.POST("/ws", srv.ServeWs)

	log.Fatal(r.Run())
}

func (srv *Srv) Flights(c *gin.Context) {
	filter := models.Filter{}
	err := c.Bind(&filter)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	log.Println(filter)

	srv.Mu.RLock()
	flights := make([]models.Flight, 0)
	// TODO Sort it here depended on filter
	switch filter.SortBy {
	case 0:
		flights = srv.UnSorted
	case 1:
		// sort by flight number
		flights = srv.SortedByFlightNum
	case 2:
		// sort by departure city
		flights = srv.SortedByCity
	case 3:
		// sort by departure time
		flights = srv.SortedByDepartureTime
	}
	srv.Mu.RUnlock()

	if filter.Departure != "" {
		var filteredByDeparture []models.Flight
		for _, v := range flights {
			if v.From == filter.Departure {
				filteredByDeparture = append(filteredByDeparture, v)
			}
		}
		flights = filteredByDeparture
	}

	// lists sorted as ASC
	if filter.OrderBy {
		l := 0
		r := len(flights)
		orderedFlights := make([]models.Flight, len(flights))
		for l <= r {
			first := flights[l]
			last := flights[r]
			flights[l] = last
			flights[r] = first
			l++
			r--
		}

		flights = orderedFlights
	}

	c.JSON(http.StatusOK, flights)
}

func (srv *Srv) InsertData(c *gin.Context) {
	flights := []models.Flight{}
	err := c.Bind(&flights)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var validFlights []models.Flight

	for _, v := range flights {
		if v.Validate() == nil {

			validFlights = append(validFlights, models.Flight{
				FlightNum: v.FlightNum,
				From:      v.From,
				Departure: v.Departure,
				To:        v.To,
				Arrival:   v.Arrival,
			})
		}
	}

	srv.InsertDataCh <- validFlights

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Accepted successfully: Valid data: %d lines from %d", len(validFlights), len(flights)),
	})
}
