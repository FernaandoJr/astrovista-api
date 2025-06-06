package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// GetApodsDateRange retorna os APODs dentro de um intervalo de datas
func GetApodsDateRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	// Se o parâmetro "end" não for passado, usa a data atual
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
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
	json.NewEncoder(w).Encode(response)
}
