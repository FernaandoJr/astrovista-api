package main

import (
	"astrovista-api/database"
	"astrovista-api/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	database.Connect()

	router := mux.NewRouter()
	router.HandleFunc("/apod", handlers.GetApod).Methods("GET")
	router.HandleFunc("/apod/{date}", handlers.GetApodDate).Methods("GET")
	router.HandleFunc("/apods", handlers.GetApodsDateRange).Methods("GET")

	log.Println("Server running!")
	log.Fatal(http.ListenAndServe(":8080", router))
}
