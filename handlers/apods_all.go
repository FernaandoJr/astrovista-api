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

// GetAllApods retorna todos os APODs no banco de dados
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
	json.NewEncoder(w).Encode(response)
}
