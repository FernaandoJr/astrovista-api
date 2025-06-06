package handlers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Apod representa um registro de APOD da NASA
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

// AllApodsResponse é a estrutura de resposta para endpoints que retornam múltiplos APODs
type AllApodsResponse struct {
	Count int    `json:"count"`
	Apods []Apod `json:"apods"`
}

// ApodsDateRangeResponse é a estrutura de resposta para pesquisa por intervalo de datas
type ApodsDateRangeResponse struct {
	Count int    `json:"count"`
	Apods []Apod `json:"apods"`
}
