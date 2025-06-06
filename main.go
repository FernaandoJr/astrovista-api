package main

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	_ "astrovista-api/docs" // Importando docs para Swagger
	"astrovista-api/handlers"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           AstroVista API
// @version         1.0
// @description     API para gerenciar dados da NASA APOD (Astronomy Picture of the Day)
// @BasePath        /
func main() {
	// Inicializa conexões com banco de dados e cache
	database.Connect()
	cache.Connect()
	// Inicializa sistema de internacionalização
	i18n.InitLocales()
	i18n.InitTranslationService()

	router := mux.NewRouter()

	// Configuração do Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL para acessar a documentação JSON
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Endpoints GET públicos (sem limite de taxa)	// Adiciona middleware de detecção de idioma para todos os endpoints
	router.Use(middleware.LanguageDetector)
	router.HandleFunc("/apod", handlers.GetApod).Methods("GET")
	router.HandleFunc("/apod/{date}", handlers.GetApodDate).Methods("GET")
	router.HandleFunc("/apods", handlers.GetAllApods).Methods("GET")
	router.HandleFunc("/apods/search", handlers.SearchApods).Methods("GET")
	router.HandleFunc("/apods/date-range", handlers.GetApodsDateRange).Methods("GET")
	router.HandleFunc("/languages", handlers.GetSupportedLanguages).Methods("GET")

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
