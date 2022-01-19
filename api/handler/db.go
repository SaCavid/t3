package handler

import (
	"log"
	"os"
	"sync"
	"t3/api/router"
	repo "t3/api/sql"
	"t3/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitializeDB(srv *router.Srv) {

	// connect to database --> postgres
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv("DATABASE"),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Couldnt connect to database: %s", err.Error())
	}

	db.AutoMigrate(&models.Flights{})

	srv.DB = repo.Repo{
		Db: db,
	}
}

func DBHub(srv *router.Srv) {
	for {
		select {
		case validFlights := <-srv.InsertDataCh:
			go processNewValues(validFlights, srv)
		}
	}
}

func processNewValues(validFlights []models.Flight, srv *router.Srv) {

	// check if flight num exists or not
	var registeredFlights, notRegisteredFlights []models.Flights

	srv.Mu.RLock()
	for _, v := range validFlights {
		if _, ok := srv.MFlights[v.FlightNum]; ok {
			registeredFlights = append(registeredFlights, models.Flights{})
		} else {
			notRegisteredFlights = append(notRegisteredFlights, models.Flights{})
		}
	}
	srv.Mu.RUnlock()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go insert(notRegisteredFlights, srv, wg)
	go update(registeredFlights, srv, wg)
	wg.Wait()

	go sortLists(srv)
}

func insert(notRegisteredFlights []models.Flights, srv *router.Srv, wg sync.WaitGroup) {

	defer wg.Done()
	_, err := srv.DB.Insert(notRegisteredFlights)
	if err != nil {
		log.Println(err)
	}
}

func update(registeredFlights []models.Flights, srv *router.Srv, wg sync.WaitGroup) {

	defer wg.Done()
	_, err := srv.DB.Update(registeredFlights)
	if err != nil {
		log.Println(err)
	}
}

func sortLists(srv *router.Srv) {

	flights, err := srv.DB.GellAllFlights()
	if err != nil {
		log.Println(err)
		return
	}

	srv.Mu.Lock()
	srv.UnSorted = flights

	// clean map
	srv.MFlights = make(map[int]uint)
	for k, v := range flights {
		srv.MFlights[v.FlightNum] = uint(k) // dump uint
	}
	srv.Mu.Unlock()

	go sortByFlightNum(flights, srv)
	go sortbyDepartureCity(flights, srv)
	go sortByDepartureTime(flights, srv)
}

func sortByFlightNum(unSorted []models.Flight, srv *router.Srv) {
	//	merge.ParallelMergesort()
}

func sortbyDepartureCity(unSorted []models.Flight, srv *router.Srv) {
	//
}

func sortByDepartureTime(unSorted []models.Flight, srv *router.Srv) {
	//
}
