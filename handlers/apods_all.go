package handlers

import (
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// GetAllApods retorna todos os APODs no banco de dados
// @Summary Obtém todos os APODs
// @Description Retorna todas as imagens astronômicas do dia cadastradas
// @Tags APODs
// @Accept json
// @Produce json
// @Success 200 {object} AllApodsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods [get]
func GetAllApods(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("MongoDB error: %v\n", err)
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

	// Obtém o idioma da requisição
	lang := middleware.GetLanguageFromContext(r.Context())

	// Se não for inglês, tenta traduzir cada APOD no resultado
	if lang != "en" {
		// Cria uma resposta traduzida
		var translatedResponse AllApodsResponse
		translatedResponse.Count = response.Count
		translatedApods := make([]map[string]interface{}, 0, len(response.Apods))

		// Traduz cada APOD
		for _, apod := range response.Apods {
			// Converte para map para permitir tradução
			apodMap := map[string]interface{}{
				"_id":             apod.ID,
				"date":            apod.Date,
				"explanation":     apod.Explanation,
				"hdurl":           apod.Hdurl,
				"media_type":      apod.MediaType,
				"service_version": apod.ServiceVersion,
				"title":           apod.Title,
				"url":             apod.Url,
			}

			// Traduz os campos necessários
			if err := i18n.TranslateAPOD(apodMap, lang); err != nil {
				log.Printf("Erro ao traduzir APOD: %v", err)
			}

			translatedApods = append(translatedApods, apodMap)
		}

		// Cria uma resposta personalizada
		customResponse := map[string]interface{}{
			"count": translatedResponse.Count,
			"apods": translatedApods,
		}

		// Envia a versão traduzida
		json.NewEncoder(w).Encode(customResponse)
	} else {
		// Sem tradução, envia original
		json.NewEncoder(w).Encode(response)
	}
}
