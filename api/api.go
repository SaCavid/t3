package api

import (
	"fmt"
	"log"
	"t3/api/handler"
	"t3/api/router"
	"t3/models"

	"github.com/joho/godotenv"
)

func DescribeTitle() {
	fmt.Println(`
	___ _    _      _   _       _          _    _            _   
	| __| |  (_)__ _| |_| |_    /_\   _____(_)__| |_ __ _ _ _| |_ 
	| _|| |__| / _  | ' \  _|  / _ \ (_-<_-< (_-<  _/ _  | ' \  _|
	|_| |____|_\__, |_||_\__| /_/ \_\/__/__/_/__/\__\__,_|_||_\__|
			   |___/                                                                                                                                
	`)
}

func Startup() {
	log.SetFlags(log.Lshortfile | log.Ldate)

	// load env vars
	err := godotenv.Load("../env/.env")
	if err != nil {
		// unable to connect to database. Quit app
		log.Fatal("Failed to load env! ", err)
	}
}

func Listen() {

	srv := &router.Srv{
		MFlights:     make(map[int]models.Flight),
		InsertDataCh: make(chan []models.Flight, 10),
	}

	handler.InitializeDB(srv)
	go handler.DBHub(srv)

	router.Route(srv)
}
