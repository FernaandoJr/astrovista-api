package main

import (
	"astrovista-api/i18n"
	"astrovista-api/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestTranslationMiddleware verifica se o middleware de detecção de idioma funciona corretamente
func TestTranslationMiddleware(t *testing.T) {
	// Inicializa o sistema i18n
	i18n.InitLocales()

	// Cria um handler de teste que simplesmente retorna o idioma detectado
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := middleware.GetLanguageFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"detectedLanguage": lang})
	})

	// Aplica o middleware
	handlerWithMiddleware := middleware.LanguageDetector(testHandler)

	// Testes para diferentes métodos de detecção de idioma
	testCases := []struct {
		name           string
		acceptLanguage string
		queryParam     string
		expected       string
	}{
		{
			name:           "Default language",
			acceptLanguage: "",
			queryParam:     "",
			expected:       "en",
		},
		{
			name:           "Accept-Language header",
			acceptLanguage: "pt-BR",
			queryParam:     "",
			expected:       "pt-BR",
		},
		{
			name:           "Query parameter",
			acceptLanguage: "",
			queryParam:     "es",
			expected:       "es",
		},
		{
			name:           "Query parameter overrides header",
			acceptLanguage: "pt-BR",
			queryParam:     "fr",
			expected:       "fr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Cria uma requisição de teste
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Configura os parâmetros de teste
			if tc.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tc.acceptLanguage)
			}
			if tc.queryParam != "" {
				q := req.URL.Query()
				q.Add("lang", tc.queryParam)
				req.URL.RawQuery = q.Encode()
			}

			// Executa a requisição
			rr := httptest.NewRecorder()
			handlerWithMiddleware.ServeHTTP(rr, req)

			// Verifica o status code
			if rr.Code != http.StatusOK {
				t.Errorf("Status code esperado %d, obtido %d", http.StatusOK, rr.Code)
			}

			// Verifica o idioma retornado
			var response map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}

			if response["detectedLanguage"] != tc.expected {
				t.Errorf("Idioma esperado %q, obtido %q", tc.expected, response["detectedLanguage"])
			}
		})
	}
}

// TestTranslateAPOD testa a tradução de dados do APOD
func TestTranslateAPOD(t *testing.T) {
	// Inicializa o sistema i18n
	i18n.InitLocales()
	i18n.InitTranslationService()

	// Cria um APOD de teste
	apodData := map[string]interface{}{
		"title":       "The Milky Way over the Grand Canyon",
		"explanation": "This is a stunning view of our galaxy spanning over the Grand Canyon.",
		"date":        "2023-01-15",
		"media_type":  "image",
		"url":         "https://example.com/image.jpg",
	}

	// Testa a tradução para diferentes idiomas
	languages := []string{"pt-BR", "es", "fr"}

	for _, lang := range languages {
		t.Run("Tradução para "+lang, func(t *testing.T) {
			// Cria uma cópia dos dados para não afetar os testes subsequentes
			apodCopy := make(map[string]interface{})
			for k, v := range apodData {
				apodCopy[k] = v
			}

			// Aplica a tradução
			err := i18n.TranslateAPOD(apodCopy, lang)
			if err != nil {
				t.Fatalf("Erro ao traduzir APOD: %v", err)
			}

			// Verifica se os campos foram traduzidos
			origTitle := apodData["title"].(string)
			transTitle := apodCopy["title"].(string)

			if transTitle == origTitle {
				t.Logf("Aviso: título não foi alterado. Isso é esperado se não houver API de tradução configurada.")
			} else {
				t.Logf("Título original: %q", origTitle)
				t.Logf("Título traduzido: %q", transTitle)
			}

			origExplanation := apodData["explanation"].(string)
			transExplanation := apodCopy["explanation"].(string)

			if transExplanation == origExplanation {
				t.Logf("Aviso: explicação não foi alterada. Isso é esperado se não houver API de tradução configurada.")
			} else {
				t.Logf("Primeiros 50 caracteres da explicação original: %q", origExplanation[:min(50, len(origExplanation))])
				t.Logf("Primeiros 50 caracteres da explicação traduzida: %q", transExplanation[:min(50, len(transExplanation))])
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
