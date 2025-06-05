package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Apod struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"` // ID do MongoDB
	Date           string             `bson:"date" json:"date"`         // Data no formato string (ex: "1995-06-16")
	Explanation    string             `bson:"explanation" json:"explanation"`
	Hdurl          string             `bson:"hdurl" json:"hdurl"`
	MediaType      string             `bson:"media_type" json:"media_type"`
	ServiceVersion string             `bson:"service_version" json:"service_version"`
	Title          string             `bson:"title" json:"title"`
	Url            string             `bson:"url" json:"url"`
}

func GetApod(w http.ResponseWriter, r *http.Request) {
	// Cria contexto com timeout para a operação no banco
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var apod Apod // sua struct para armazenar o resultado

	// Buscar 1 documento ordenado por "date" desc (mais recente)
	err := database.ApodCollection.FindOne(
		ctx,
		bson.M{}, // filtro vazio = todos
		options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}}), // ordenar desc
	).Decode(&apod) // decodificar o resultado na variável apod

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Document not found",
		})
		return
	}

	// Se achou, retorna JSON para o cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apod)
}

func GetApodDate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	date := params["date"]

	var apod Apod

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filtro: buscar documento com campo "date" igual ao parâmetro recebido
	filter := bson.M{"date": date}

	err := database.ApodCollection.FindOne(ctx, filter).Decode(&apod)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Document not found! Please check the date format (YYYY-MM-DD).",
			"details": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apod)
}

type AllApodsResponse struct {
	Count int    `json:"count"`
	Apods []Apod `json:"apods"`
}

func GetAllApods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("MongoDB error: %v\n", err) // Add this line
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error fetching documents",
			"details": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var apods []Apod
	if err = cursor.All(ctx, &apods); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding documents",
			"details": err.Error(),
		})
		return
	}

	// Check if no documents were found
	if len(apods) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "No documents found!",
			"details": "No APODs found in the database.",
		})
		return
	}

	var response AllApodsResponse
	response.Count = len(apods)

	if len(apods) == 1 {
		response.Apods = []Apod{apods[0]} // um único objeto como slice
	} else {
		response.Apods = apods // array de objetos
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Size", fmt.Sprintf("%d", len(apods)))
	json.NewEncoder(w).Encode(response)
}

func PostApod(w http.ResponseWriter, r *http.Request) {
	// Verifica token básico de API (para serviço interno/agendado)
	apiToken := r.Header.Get("X-API-Token")
	if apiToken != os.Getenv("INTERNAL_API_TOKEN") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized - Valid API token required",
		})
		return
	}

	// Obtém a chave da API NASA das variáveis de ambiente
	nasaAPIKey := os.Getenv("NASA_API_KEY")
	if nasaAPIKey == "" {
		nasaAPIKey = "DEMO_KEY" // Chave de demonstração (limite de uso baixo)
	}

	// URL da API NASA APOD
	nasaURL := fmt.Sprintf("https://api.nasa.gov/planetary/apod?api_key=%s", nasaAPIKey)

	// Faz a requisição para a API da NASA
	resp, err := http.Get(nasaURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error fetching data from NASA API",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Verifica se a resposta foi bem-sucedida
	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "NASA API returned an error",
			"details": resp.Status,
		})
		return
	}

	// Decodifica a resposta JSON
	var apod Apod
	if err := json.NewDecoder(resp.Body).Decode(&apod); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding NASA API response",
			"details": err.Error(),
		})
		return
	}

	// Cria um contexto com timeout para operações no MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verifica se já existe um documento com essa data
	filter := bson.M{"date": apod.Date}
	existingApod := database.ApodCollection.FindOne(ctx, filter)

	// Se não houve erro, significa que já existe um documento com essa data
	if existingApod.Err() == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "APOD already exists for this date",
			"details": apod.Date,
		})
		return
	} else if existingApod.Err() != mongo.ErrNoDocuments {
		// Se o erro for diferente de ErrNoDocuments, houve um problema no banco
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error checking for existing APOD",
			"details": existingApod.Err().Error(),
		})
		return
	}

	// Insere o novo documento
	result, err := database.ApodCollection.InsertOne(ctx, apod)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error inserting APOD into database",
			"details": err.Error(),
		})
		return
	}

	// Retorna sucesso com o ID inserido
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "APOD successfully added to database",
		"id":      result.InsertedID,
		"date":    apod.Date,
	})
}
