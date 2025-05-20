package handlers

import (
	"astrovista-api/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Apod struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"` // ID do MongoDB
	Date           string             `bson:"date" json:"date"`         // Data no formato string (ex: "1995-06-16")
	Explanation    string             `bson:"explanation" json:"explanation"`
	Hdurl          string             `bson:"hdurl" json:"hdurl"`
	MediaType      string             `bson:"media_type" json:"media_type"`
	ServiceVersion string             `bson:"service_version" json:"service_version"`
	Title          string             `bson:"title" json:"title"`
	Url            string             `bson:"url" json:"url"`
}

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
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Document not found! Please check the date format (YYYY-MM-DD).",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apod)
}

type ApodsDateRangeResponse struct {
	Count int    `json:"count"`
	Apods []Apod `json:"apods"`
}

func GetApodsDateRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	if (startDate == "" && endDate != "") || (startDate != "" && endDate == "") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Both 'start' and 'end' query parameters are required in the format YYYY-MM-DD.",
		})
		return
	}

	filter := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	if startDate == "" && endDate == "" {
		filter = bson.M{} // Se ambos os parâmetros estiverem vazios, retorna todos os documentos
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.ApodCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Erro ao buscar documentos", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var apods []Apod
	if err = cursor.All(ctx, &apods); err != nil {
		http.Error(w, "Erro ao ler documentos", http.StatusInternalServerError)
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
