package main

import (
	"log"
	"net/http"
)

type MyHttpApp struct {
	healthcheck HealthcheckHandler
}

func main() {
	var app MyHttpApp

	app.healthcheck = HealthcheckHandler{}
	http.Handle("/healthcheck", app.healthcheck)
	log.Printf("Starting server on port 9000")

	startServerErr := http.ListenAndServe(":9000", nil)
	if startServerErr != nil {
		log.Fatal(startServerErr)
	}
}
