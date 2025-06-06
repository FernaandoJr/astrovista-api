package handlers

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Apod representa um registro de APOD da NASA
// swagger:model Apod
type Apod struct {
	// ID do MongoDB
	// example: 507f1f77bcf86cd799439011
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	// Data no formato string (ex: "1995-06-16")
	// example: 2023-01-15
	// format: date
	Date string `bson:"date" json:"date"`
	// Explicação da imagem astronômica do dia
	// example: Uma bela nebulosa capturada pelo telescópio Hubble
	Explanation string `bson:"explanation" json:"explanation"`
	// URL da imagem em alta definição
	// example: https://apod.nasa.gov/apod/image/2301/M31_HubbleSpitzerGendler_960.jpg
	// format: uri
	Hdurl string `bson:"hdurl" json:"hdurl"`
	// Tipo de mídia (imagem ou vídeo)
	// example: image
	// enum: image,video
	MediaType string `bson:"media_type" json:"media_type"`
	// Versão do serviço da API
	// example: v1
	ServiceVersion string `bson:"service_version" json:"service_version"`
	// Título da imagem astronômica do dia
	// example: Galáxia de Andrômeda
	Title string `bson:"title" json:"title"`
	// URL da imagem em resolução padrão
	// example: https://apod.nasa.gov/apod/image/2301/M31_HubbleSpitzerGendler_960.jpg
	// format: uri
	Url string `bson:"url" json:"url"`
}

// AllApodsResponse é a estrutura de resposta para endpoints que retornam múltiplos APODs
// swagger:model AllApodsResponse
type AllApodsResponse struct {
	// Número total de APODs encontrados
	// example: 15
	Count int `json:"count"`
	// Lista de APODs
	Apods []Apod `json:"apods"`
}

// ApodsDateRangeResponse é a estrutura de resposta para pesquisa por intervalo de datas
// swagger:model ApodsDateRangeResponse
type ApodsDateRangeResponse struct {
	// Número total de APODs encontrados
	// example: 7
	Count int `json:"count"`
	// Lista de APODs
	Apods []Apod `json:"apods"`
}

// SearchResponse é a estrutura de resposta para o endpoint de pesquisa
// swagger:model SearchResponse
type SearchResponse struct {
	// Número total de resultados encontrados
	// example: 42
	TotalResults int `json:"totalResults"`
	// Número da página atual
	// example: 1
	Page int `json:"page"`
	// Itens por página
	// example: 20
	PerPage int `json:"perPage"`
	// Total de páginas disponíveis
	// example: 3
	TotalPages int `json:"totalPages"`
	// Resultados da busca
	Results []Apod `json:"results"`
}

// MarshalJSON customiza a serialização JSON para suportar tradução
func (a Apod) MarshalJSON() ([]byte, error) {
	// Cria um map com os campos do APOD
	apodMap := map[string]interface{}{
		"_id":             a.ID,
		"date":            a.Date,
		"explanation":     a.Explanation,
		"hdurl":           a.Hdurl,
		"media_type":      a.MediaType,
		"service_version": a.ServiceVersion,
		"title":           a.Title,
		"url":             a.Url,
	}

	// Na serialização padrão não fazemos nada
	// A tradução será aplicada nos handlers antes de chamar json.Marshal

	return json.Marshal(apodMap)
}
