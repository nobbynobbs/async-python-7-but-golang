package main

import (
	"log"
	"net/http"

	"buses/internal"
)

func main() {
	busesStorage := &internal.MapBasedBusStorage{
		Buses: make(map[string]*internal.Bus),
	}
	go func() {
		log.Println("starting webclients server...")
		defer log.Println("webclients server stopped...")
		service := internal.WebclientsServer{BusStorage: busesStorage}
		server := http.NewServeMux()
		server.Handle("/", &service)
		_ = http.ListenAndServe(":8000", server)
	}()
	go func() {
		log.Println("starting buses server...")
		defer log.Println("buses server stopped...")
		service := internal.BusesServer{BusStorage: busesStorage}
		server := http.NewServeMux()
		server.Handle("/", &service)
		_ = http.ListenAndServe(":8080", server)
	}()
	forever := make(chan struct{})
	<-forever
}
