package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostApod busca o APOD mais recente da API da NASA e o adiciona ao banco de dados
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
