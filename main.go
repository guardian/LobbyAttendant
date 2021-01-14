package main

import (
	"log"
	"net/http"
	"os"
)

type MyHttpApp struct {
	healthcheck HealthcheckHandler
	lister      ListingHandler
}

func main() {
	var app MyHttpApp

	app.healthcheck = HealthcheckHandler{}
	app.lister = ListingHandler{RootPath: os.Getenv("HOME"), LevelLimit: 10}
	http.Handle("/healthcheck", app.healthcheck)
	http.Handle("/list", app.lister)
	log.Printf("Starting server on port 9000")

	startServerErr := http.ListenAndServe(":9000", nil)
	if startServerErr != nil {
		log.Fatal(startServerErr)
	}
}
