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

	sortLists(srv, false, nil)
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

	wg := &sync.WaitGroup{}

	if len(notRegisteredFlights) > 0 {
		wg.Add(1)
		log.Println(notRegisteredFlights)
		go insert(notRegisteredFlights, srv, wg)
	}

	if len(registeredFlights) > 0 {
		wg.Add(1)
		go update(registeredFlights, srv, wg)
	}

	wg.Add(1)
	go sortLists(srv, true, wg)

	wg.Wait()

}

func insert(notRegisteredFlights []models.Flights, srv *router.Srv, wg *sync.WaitGroup) {

	defer wg.Done()
	_, err := srv.DB.Insert(notRegisteredFlights)
	if err != nil {
		log.Println(err)
	}
}

func update(registeredFlights []models.Flights, srv *router.Srv, wg *sync.WaitGroup) {

	defer wg.Done()
	_, err := srv.DB.Update(registeredFlights)
	if err != nil {
		log.Println(err)
	}
}

func sortLists(srv *router.Srv, initialized bool, Wg *sync.WaitGroup) {

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

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go sortByFlightNum(flights, srv, wg)
	go sortbyDepartureCity(flights, srv, wg)
	go sortByDepartureTime(flights, srv, wg, initialized)

	wg.Wait()
	if Wg != nil {
		Wg.Done()
	}
}

func sortByFlightNum(unSorted []models.Flight, srv *router.Srv, wg *sync.WaitGroup) {
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
	wg.Done()
}

func sortbyDepartureCity(unSorted []models.Flight, srv *router.Srv, wg *sync.WaitGroup) {

	newUnSorted := make([]models.Flight, len(unSorted))
	copy(newUnSorted, unSorted)
	sort.SliceStable(newUnSorted, func(i, j int) bool {
		return newUnSorted[i].From < newUnSorted[j].From
	})

	srv.Mu.Lock()
	defer srv.Mu.Unlock()
	sorted := newUnSorted
	srv.SortedByCity = sorted // sorted
	wg.Done()
}

func sortByDepartureTime(unSorted []models.Flight, srv *router.Srv, wg *sync.WaitGroup, initialized bool) {

	newUnSorted := make([]models.Flight, len(unSorted))
	copy(newUnSorted, unSorted)
	sort.SliceStable(newUnSorted, func(i, j int) bool {
		return newUnSorted[i].Departure.Unix() < newUnSorted[j].Departure.Unix()
	})

	srv.Mu.Lock()
	defer srv.Mu.Unlock()
	sorted := newUnSorted
	srv.SortedByDepartureTime = sorted // sorted

	// time filtered data sent to ws
	log.Println(initialized)
	if initialized {
		srv.Hub.Broadcasts <- &sorted
		log.Println("Sent invoke")
	}

	wg.Done()
}
