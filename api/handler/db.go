package handler

import (
	"log"
	"os"
	"sort"
	"sync"
	"t3/api/router"
	repo "t3/api/sql"
	"t3/models"
	merge "t3/utils"

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

	initializeLists(srv)
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
			registeredFlights = append(registeredFlights, models.Flights{
				Flight: v,
			})
		} else {
			notRegisteredFlights = append(notRegisteredFlights, models.Flights{
				Flight: v,
			})
		}
	}
	srv.Mu.RUnlock()

	wg := sync.WaitGroup{}

	if len(notRegisteredFlights) > 0 {
		wg.Add(1)
		log.Println(notRegisteredFlights)
		go insert(notRegisteredFlights, srv, wg)
	}

	if len(registeredFlights) > 0 {
		wg.Add(1)
		go update(registeredFlights, srv, wg)
	}
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

func initializeLists(srv *router.Srv) {

	flights, err := srv.DB.GellAllFlights()
	if err != nil {
		log.Println(err)
		return
	}

	srv.Mu.Lock()
	srv.UnSorted = flights

	// clean map
	srv.MFlights = make(map[int]models.Flight)
	for _, v := range flights {
		srv.MFlights[v.FlightNum] = v // dump uint
	}
	srv.Mu.Unlock()

	go sortByFlightNum(flights, srv)
	go sortbyDepartureCity(flights, srv)
	go sortByDepartureTime(flights, srv)
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
	srv.MFlights = make(map[int]models.Flight)
	for _, v := range flights {
		srv.MFlights[v.FlightNum] = v // dump uint
	}
	srv.Mu.Unlock()

	srv.Hub.Broadcasts <- &flights

	go sortByFlightNum(flights, srv)
	go sortbyDepartureCity(flights, srv)
	go sortByDepartureTime(flights, srv)
}

func sortByFlightNum(unSorted []models.Flight, srv *router.Srv) {
	var list []int
	for _, v := range unSorted {
		list = append(list, v.FlightNum)
	}

	merge.ParallelMergesort(list)

	var sorted []models.Flight
	for _, v := range list {
		srv.Mu.RLock()
		flight := srv.MFlights[v]
		srv.Mu.RUnlock()
		sorted = append(sorted, flight)
	}

	srv.Mu.Lock()
	defer srv.Mu.Unlock()
	srv.SortedByFlightNum = sorted // sorted
}

func sortbyDepartureCity(unSorted []models.Flight, srv *router.Srv) {

	sort.SliceStable(unSorted, func(i, j int) bool {
		return unSorted[i].From < unSorted[j].From
	})

	srv.Mu.Lock()
	defer srv.Mu.Unlock()
	sorted := unSorted
	srv.SortedByCity = sorted // sorted
}

func sortByDepartureTime(unSorted []models.Flight, srv *router.Srv) {

	sort.SliceStable(unSorted, func(i, j int) bool {
		return unSorted[i].Departure.String() < unSorted[j].Departure.String()
	})

	srv.Mu.Lock()
	defer srv.Mu.Unlock()
	sorted := unSorted
	srv.SortedByDepartureTime = sorted // sorted
}
