package main

import (
	"astrovista-api/database"
	"astrovista-api/handlers"
	"astrovista-api/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	database.Connect()
	router := mux.NewRouter()
	// Endpoints GET públicos (sem limite de taxa)
	router.HandleFunc("/apod", handlers.GetApod).Methods("GET")
	router.HandleFunc("/apod/{date}", handlers.GetApodDate).Methods("GET")
	router.HandleFunc("/apods", handlers.GetAllApods).Methods("GET")
	router.HandleFunc("/apods/search", handlers.SearchApods).Methods("GET")

	// Rate limiter: 1 requisição por minuto
	rateLimiter := middleware.NewRateLimiter(1, 1*time.Minute)

	// Endpoint POST com limite de taxa aplicado
	postRouter := router.PathPrefix("/apod").Subrouter()
	postRouter.Use(rateLimiter.Limit)
	postRouter.HandleFunc("", handlers.PostApod).Methods("POST")
	// Determina a porta do servidor (padrão 8080, ou usa variável de ambiente PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server running on port %s!", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
