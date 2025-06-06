package handlers

import (
	"astrovista-api/i18n"
	"encoding/json"
	"net/http"
)

// LanguageInfo contém informações sobre um idioma suportado
type LanguageInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"nativeName"`
}

// GetSupportedLanguages retorna a lista de idiomas suportados pela API
// @Summary Lista idiomas suportados
// @Description Retorna a lista de idiomas suportados pela API AstroVista
// @Tags Configuração
// @Accept json
// @Produce json
// @Success 200 {array} LanguageInfo
// @Router /languages [get]
func GetSupportedLanguages(w http.ResponseWriter, r *http.Request) { // Mapeia os códigos de idioma para seus nomes
	languageNames := map[string]map[string]string{
		"en": {
			"name":       "English",
			"nativeName": "English",
		},
		"pt-BR": {
			"name":       "Brazilian Portuguese",
			"nativeName": "Português do Brasil",
		},
		"es": {
			"name":       "Spanish",
			"nativeName": "Español",
		},
		"fr": {
			"name":       "French",
			"nativeName": "Français",
		},
		"de": {
			"name":       "German",
			"nativeName": "Deutsch",
		},
		"it": {
			"name":       "Italian",
			"nativeName": "Italiano",
		},
	}

	// Prepara a lista de idiomas suportados
	var languages []LanguageInfo

	for _, lang := range i18n.SupportedLanguages {
		info, exists := languageNames[lang]
		if !exists {
			info = map[string]string{
				"name":       lang,
				"nativeName": lang,
			}
		}

		languages = append(languages, LanguageInfo{
			Code:       lang,
			Name:       info["name"],
			NativeName: info["nativeName"],
		})
	}

	// Retorna a lista em formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(languages)
}
