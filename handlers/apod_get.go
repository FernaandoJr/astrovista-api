package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetApod retorna o APOD mais recente
func GetApod(w http.ResponseWriter, r *http.Request) {
	// Cria contexto com timeout para a operação no banco
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var apod Apod // sua struct para armazenar o resultado

	// Buscar 1 documento ordenado por "date" desc (mais recente)
	err := database.ApodCollection.FindOne(
		ctx,
		bson.M{}, // filtro vazio = todos
		options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}}), // ordenar desc
	).Decode(&apod) // decodificar o resultado na variável apod

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
	json.NewEncoder(w).Encode(apod)
}
