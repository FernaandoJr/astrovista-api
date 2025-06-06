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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetApod retorna o APOD mais recente
// @Summary Obtém o APOD mais recente
// @Description Retorna a imagem astronômica do dia mais recente
// @Tags APOD
// @Accept json
// @Produce json
// @Success 200 {object} Apod
// @Failure 400 {object} map[string]string
// @Router /apod [get]
func GetApod(w http.ResponseWriter, r *http.Request) {
	// Cria contexto com timeout para a operação no banco
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var apod Apod // struct para armazenar o resultado

	// Chave de cache para o APOD mais recente
	cacheKey := "apod:latest"

	// Tenta recuperar do cache primeiro
	found, err := cache.Get(ctx, cacheKey, &apod)
	if err != nil {
		// Erro ao acessar o cache, apenas registra e continua
		log.Printf("Erro ao acessar cache: %v", err)
	}
	// Se encontrou no cache, retorna imediatamente
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
	// Se não encontrou no cache, busca no banco de dados
	err = database.ApodCollection.FindOne(
		ctx,
		bson.M{}, // filtro vazio = todos
		options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}}), // ordenar desc
	).Decode(&apod) // decodificar o resultado na variável apod

	// Se encontrou no banco de dados, armazena no cache para futuras requisições
	if err == nil {
		// Armazena no cache por 1 hora (o mais recente pode mudar diariamente)
		if cacheErr := cache.Set(ctx, cacheKey, apod, 1*time.Hour); cacheErr != nil {
			log.Printf("Erro ao armazenar no cache: %v", cacheErr)
		}
	}
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
