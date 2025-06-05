package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Estrutura de resposta para o endpoint de pesquisa
type SearchResponse struct {
	TotalResults int    `json:"totalResults"`
	Page         int    `json:"page"`
	PerPage      int    `json:"perPage"`
	TotalPages   int    `json:"totalPages"`
	Results      []Apod `json:"results"`
}

// Função para procurar APODs com vários filtros e paginação
func SearchApods(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
