package handlers

import (
	"astrovista-api/cache"
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

// GetApodsDateRange retorna os APODs dentro de um intervalo de datas
// @Summary Obtém APODs por intervalo de datas
// @Description Retorna as imagens astronômicas do dia dentro de um intervalo de datas especificado
// @Tags APODs
// @Accept json
// @Produce json
// @Param start query string false "Data de início (formato YYYY-MM-DD)" example("2023-01-01")
// @Param end query string false "Data de fim (formato YYYY-MM-DD)" example("2023-01-31")
// @Success 200 {object} ApodsDateRangeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods/date-range [get]
func GetApodsDateRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	// Se o parâmetro "end" não for passado, usa a data atual
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Gera uma chave de cache com base nos parâmetros da consulta
	cacheKey := fmt.Sprintf("apods:range:%s:%s", startDate, endDate)

	// Tenta recuperar do cache
	var cachedResponse ApodsDateRangeResponse
	found, err := cache.Get(context.Background(), cacheKey, &cachedResponse)
	if err != nil {
		log.Printf("Erro ao acessar cache para intervalo de datas: %v", err)
	}
	// Se encontrou no cache, retorna imediatamente
	if found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")

		// Obtém o idioma da requisição
		lang := middleware.GetLanguageFromContext(r.Context())

		// Se não for inglês, tenta traduzir cada APOD no resultado
		if lang != "en" {
			// Cria uma resposta traduzida
			var translatedResponse ApodsDateRangeResponse
			translatedResponse.Count = cachedResponse.Count
			translatedApods := make([]map[string]interface{}, 0, len(cachedResponse.Apods))

			// Traduz cada APOD
			for _, apod := range cachedResponse.Apods {
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
			json.NewEncoder(w).Encode(cachedResponse)
		}
		return
	}

	// Verifica se endDate é uma data válida (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Invalid end date format. Use YYYY-MM-DD.",
			"details": err.Error(),
		})
		return
	}

	filter := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// Se ambos os parâmetros estiverem vazios, retorna todos os documentos
	if startDate == "" && endDate == "" {
		filter = bson.M{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, filter)
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
			"error":   "No documents found for the given date range.",
			"details": fmt.Sprintf("Start date: %s", startDate),
		})
		return
	}

	var response ApodsDateRangeResponse
	response.Count = len(apods)

	if len(apods) == 1 {
		response.Apods = []Apod{apods[0]} // um único objeto como slice
	} else {
		response.Apods = apods // array de objetos
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Size", fmt.Sprintf("%d", len(apods)))
	w.Header().Set("X-Cache", "MISS") // Indica que veio do banco, não do cache

	// Armazena no cache para consultas futuras
	// Intervalos de datas específicos podem ser armazenados por um tempo maior (12 horas)
	if cacheErr := cache.Set(context.Background(), cacheKey, response, 12*time.Hour); cacheErr != nil {
		log.Printf("Erro ao armazenar intervalo de datas no cache: %v", cacheErr)
	}

	// Obtém o idioma da requisição
	lang := middleware.GetLanguageFromContext(r.Context())

	// Se não for inglês, tenta traduzir cada APOD no resultado
	if lang != "en" {
		// Cria uma resposta traduzida
		var translatedResponse ApodsDateRangeResponse
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
