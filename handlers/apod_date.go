package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// GetApodDate retorna um APOD específico por data
// @Summary Obtém um APOD por data específica
// @Description Retorna a imagem astronômica do dia para a data especificada
// @Tags APOD
// @Accept json
// @Produce json
// @Param date path string true "Data no formato YYYY-MM-DD" example("2023-01-15")
// @Success 200 {object} Apod
// @Failure 400 {object} map[string]interface{} "Erro ao obter APOD"
// @Router /apod/{date} [get]
func GetApodDate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	date := params["date"]

	var apod Apod

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Chave de cache específica para a data
	cacheKey := "apod:date:" + date

	// Tenta recuperar do cache primeiro
	found, err := cache.Get(ctx, cacheKey, &apod)
	if err != nil {
		// Erro ao acessar o cache, apenas registra e continua
		log.Printf("Erro ao acessar cache: %v", err)
	}
	// Se encontrou no cache, aplica tradução e retorna
	if found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")

		// Obtém o idioma da requisição
		lang := middleware.GetLanguageFromContext(r.Context())

		// Se não for inglês, tenta traduzir
		if lang != "en" {
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

			// Envia a versão traduzida
			json.NewEncoder(w).Encode(apodMap)
		} else {
			// Sem tradução, envia original
			json.NewEncoder(w).Encode(apod)
		}
		return
	}

	// Filtro: buscar documento com campo "date" igual ao parâmetro recebido
	filter := bson.M{"date": date}
	// Se não encontrou no cache, busca no banco de dados
	err = database.ApodCollection.FindOne(ctx, filter).Decode(&apod)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Document not found! Please check the date format (YYYY-MM-DD).",
			"details": err.Error(),
		})
		return
	}

	// Se encontrou no banco de dados, armazena no cache para futuras requisições
	// APODs históricos nunca mudam, então podemos usar uma expiração longa (30 dias)
	if cacheErr := cache.Set(ctx, cacheKey, apod, 30*24*time.Hour); cacheErr != nil {
		log.Printf("Erro ao armazenar no cache: %v", cacheErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS") // Indica que veio do banco, não do cache

	// Obtém o idioma da requisição
	lang := middleware.GetLanguageFromContext(r.Context())

	// Se não for inglês, tenta traduzir
	if lang != "en" {
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

		// Envia a versão traduzida
		json.NewEncoder(w).Encode(apodMap)
	} else {
		// Sem tradução, envia original
		json.NewEncoder(w).Encode(apod)
	}
}
