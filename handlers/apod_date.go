package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// GetApodDate retorna um APOD específico por data
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
