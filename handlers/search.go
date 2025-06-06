package handlers

import (
	"astrovista-api/cache"
	"astrovista-api/database"
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SearchResponse está definido em models.go

// Função para procurar APODs com vários filtros e paginação
// @Summary Pesquisa avançada de APODs
// @Description Busca APODs com filtros, paginação e ordenação
// @Tags APODs
// @Accept json
// @Produce json
// @Param page query int false "Número da página" example(1) minimum(1)
// @Param perPage query int false "Itens por página (1-200)" example(20) minimum(1) maximum(200)
// @Param mediaType query string false "Tipo de mídia (image, video ou any)" example(image) Enums(image, video, any)
// @Param search query string false "Texto para busca em título e explicação" example(nebulosa)
// @Param startDate query string false "Data inicial (formato YYYY-MM-DD)" example(2023-01-01)
// @Param endDate query string false "Data final (formato YYYY-MM-DD)" example(2023-01-31)
// @Param sort query string false "Ordenação (asc ou desc)" example(desc) Enums(asc, desc)
// @Success 200 {object} SearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /apods/search [get]
func SearchApods(w http.ResponseWriter, r *http.Request) {
	// Cria uma chave de cache a partir da query string completa
	queryHash := md5.Sum([]byte(r.URL.RawQuery))
	cacheKey := "search:" + hex.EncodeToString(queryHash[:])

	// Tenta recuperar resultados do cache
	var cachedResponse SearchResponse
	found, err := cache.Get(r.Context(), cacheKey, &cachedResponse)
	if err != nil {
		log.Printf("Erro ao acessar cache para busca: %v", err)
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
			translatedResponse := SearchResponse{
				TotalResults: cachedResponse.TotalResults,
				Page:         cachedResponse.Page,
				PerPage:      cachedResponse.PerPage,
				TotalPages:   cachedResponse.TotalPages,
			}

			translatedApods := make([]map[string]interface{}, 0, len(cachedResponse.Results))

			// Traduz cada APOD
			for _, apod := range cachedResponse.Results {
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
				"totalResults": translatedResponse.TotalResults,
				"page":         translatedResponse.Page,
				"perPage":      translatedResponse.PerPage,
				"totalPages":   translatedResponse.TotalPages,
				"results":      translatedApods,
			}

			// Envia a versão traduzida
			json.NewEncoder(w).Encode(customResponse)
		} else {
			// Sem tradução, envia original
			json.NewEncoder(w).Encode(cachedResponse)
		}
		return
	}

	// Obtém os parâmetros da query string
	query := r.URL.Query()
	// Paginação (padrões: página 1, 20 itens por página)
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		if query.Get("page") != "" {
			fmt.Printf("Valor inválido para page ignorado: %s (usando 1 como padrão)\n", query.Get("page"))
		}
		page = 1
	} else if page < 1 {
		fmt.Printf("Valor inválido para page (menor que 1): %d (usando 1 como padrão)\n", page)
		page = 1
	}
	perPage, err := strconv.Atoi(query.Get("perPage"))
	if err != nil {
		if query.Get("perPage") != "" {
			fmt.Printf("Valor inválido para perPage ignorado: %s (usando 20 como padrão)\n", query.Get("perPage"))
		}
		perPage = 20 // Limite padrão
	} else if perPage < 1 || perPage > 200 {
		fmt.Printf("Valor fora dos limites para perPage: %d (deve estar entre 1 e 200, usando 20 como padrão)\n", perPage)
		perPage = 20 // Limite padrão
	}
	// Construção do filtro MongoDB
	filter := bson.M{}
	// Filtros diversos - valida que mediaType seja apenas "image" ou "video"
	if mediaType := query.Get("mediaType"); mediaType != "" && mediaType != "any" {
		// Verifica se o valor pertence ao enum permitido
		if mediaType == "image" || mediaType == "video" {
			filter["media_type"] = mediaType
		} else {
			// Se valor inválido, ignora o filtro (como se não tivesse sido fornecido)
			fmt.Printf("Valor inválido para mediaType ignorado: %s\n", mediaType)
		}
	}

	// Pesquisa por texto (em título e explicação)
	if search := query.Get("search"); search != "" {
		// Pesquisa texto em múltiplos campos
		textFilter := bson.M{
			"$or": []bson.M{
				{"title": bson.M{"$regex": search, "$options": "i"}},
				{"explanation": bson.M{"$regex": search, "$options": "i"}},
			},
		}

		// Se já existem outros filtros, combina com eles
		if len(filter) > 0 {
			filter = bson.M{
				"$and": []bson.M{
					filter,
					textFilter,
				},
			}
		} else {
			filter = textFilter
		}
	}
	// Filtro por data
	if startDate := query.Get("startDate"); startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err == nil {
			if endDate := query.Get("endDate"); endDate != "" {
				if _, err := time.Parse("2006-01-02", endDate); err == nil {
					filter["date"] = bson.M{
						"$gte": startDate,
						"$lte": endDate,
					}
				} else {
					// Formato de data inválido para endDate
					fmt.Printf("Formato de data inválido para endDate ignorado: %s\n", endDate)
				}
			} else {
				filter["date"] = bson.M{"$gte": startDate}
			}
		} else {
			// Formato de data inválido para startDate
			fmt.Printf("Formato de data inválido para startDate ignorado: %s\n", startDate)
		}
	} else if endDate := query.Get("endDate"); endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err == nil {
			filter["date"] = bson.M{"$lte": endDate}
		} else {
			// Formato de data inválido para endDate
			fmt.Printf("Formato de data inválido para endDate ignorado: %s\n", endDate)
		}
	}
	// Ordenação (padrão: data decrescente / mais recente primeiro)
	sortDirection := -1 // -1 = desc, 1 = asc
	if sort := query.Get("sort"); sort != "" {
		// Transforma para minúsculo para comparação case-insensitive
		sortLower := strings.ToLower(sort)
		// Valida se é um valor permitido
		if sortLower == "asc" {
			sortDirection = 1
		} else if sortLower == "desc" {
			sortDirection = -1 // mantém o padrão
		} else {
			// Ignora valor inválido e mantém o padrão
			fmt.Printf("Valor inválido para sort ignorado: %s (usando 'desc' como padrão)\n", sort)
		}
	}

	// Configuração da ordenação e paginação
	findOptions := options.Find().
		SetSort(bson.D{{Key: "date", Value: sortDirection}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Primeiro, contamos o total de documentos para calcular a paginação
	totalResults, err := database.ApodCollection.CountDocuments(ctx, filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error counting documents",
			"details": err.Error(),
		})
		return
	}

	// Em seguida, buscamos os documentos da página atual
	cursor, err := database.ApodCollection.Find(ctx, filter, findOptions)
	if err != nil {
		fmt.Printf("MongoDB search error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error searching documents",
			"details": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	// Decodifica os resultados
	var apods []Apod
	if err = cursor.All(ctx, &apods); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Error decoding search results",
			"details": err.Error(),
		})
		return
	}
	// Verifica se foram encontrados resultados
	if len(apods) == 0 && page == 1 {
		// Log para depuração mostrando os filtros usados
		filterJSON, _ := json.Marshal(filter)
		fmt.Printf("Nenhum resultado encontrado para filtro: %s\n", string(filterJSON))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "No documents found matching the search criteria",
		})
		return
	}

	// Calcula o número total de páginas
	totalPages := int(math.Ceil(float64(totalResults) / float64(perPage)))

	// Prepara a resposta
	response := SearchResponse{
		TotalResults: int(totalResults),
		Page:         page,
		PerPage:      perPage,
		TotalPages:   totalPages,
		Results:      apods,
	}
	// Armazena a resposta no cache com expiração de 5 minutos
	if err := cache.Set(r.Context(), cacheKey, response, 5*time.Minute); err != nil {
		log.Printf("Erro ao armazenar no cache: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")

	// Obtém o idioma da requisição
	lang := middleware.GetLanguageFromContext(r.Context())

	// Se não for inglês, tenta traduzir cada APOD no resultado
	if lang != "en" {
		translatedApods := make([]map[string]interface{}, 0, len(response.Results))

		// Traduz cada APOD
		for _, apod := range response.Results {
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
			"totalResults": response.TotalResults,
			"page":         response.Page,
			"perPage":      response.PerPage,
			"totalPages":   response.TotalPages,
			"results":      translatedApods,
		}

		// Envia a versão traduzida
		json.NewEncoder(w).Encode(customResponse)
	} else {
		// Sem tradução, envia original
		json.NewEncoder(w).Encode(response)
	}
}
