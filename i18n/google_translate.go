package i18n

import (
	"astrovista-api/cache"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GoogleTranslateRequest representa o formato de requisição para a API
type GoogleTranslateRequest struct {
	Q      []string `json:"q"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Format string   `json:"format"`
}

// GoogleTranslateResponse representa a resposta da API
type GoogleTranslateResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText string `json:"translatedText"`
		} `json:"translations"`
	} `json:"data"`
}

// GoogleTranslateClient implementa o serviço de tradução usando Google Translate API
type GoogleTranslateClient struct {
	apiKey     string
	httpClient *http.Client
	cache      *TranslationCache
}

// NewGoogleTranslateClient cria um novo cliente para a API do Google Translate
func NewGoogleTranslateClient(apiKey string) *GoogleTranslateClient { // Use Redis cache if available, otherwise use in-memory cache
	var translationCache *TranslationCache
	if cache.Client != nil {
		log.Println("Google Translate usando cache Redis para traduções")
		translationCache = NewTranslationCache()
		translationCache.EnableRedisCache()
	} else {
		log.Println("Google Translate usando cache em memória para traduções")
		translationCache = NewTranslationCache()
	}

	return &GoogleTranslateClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: translationCache,
	}
}

// Translate implementa a interface TranslationService para Google Translate
func (c *GoogleTranslateClient) Translate(text, sourceLang, targetLang string) (string, error) {
	// Verificar no cache primeiro
	cacheKey := fmt.Sprintf("%s:%s:%s", sourceLang, targetLang, getHashKey(text))
	if cachedText, found := c.cache.Get(cacheKey); found {
		return cachedText, nil
	}

	// Sanitiza os idiomas para o formato esperado pelo Google
	sourceLang = sanitizeLanguageCode(sourceLang)
	targetLang = sanitizeLanguageCode(targetLang)

	// Prepara a requisição
	reqBody := GoogleTranslateRequest{
		Q:      []string{text},
		Source: sourceLang,
		Target: targetLang,
		Format: "text", // ou "html" se o texto contiver HTML
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// URL da API com chave de API
	url := fmt.Sprintf("https://translation.googleapis.com/language/translate/v2?key=%s", c.apiKey)

	// Cria a requisição HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Define os headers
	req.Header.Set("Content-Type", "application/json")

	// Executa a requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao executar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API retornou status não-OK: %d", resp.StatusCode)
	}

	// Decodifica a resposta
	var translateResp GoogleTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	// Verifica se há traduções
	if len(translateResp.Data.Translations) == 0 {
		return "", fmt.Errorf("nenhuma tradução retornada")
	}

	// Obtém o texto traduzido
	translatedText := translateResp.Data.Translations[0].TranslatedText

	// Armazena no cache
	c.cache.Set(cacheKey, translatedText)

	return translatedText, nil
}

// Sanitiza o código de idioma para o formato aceito pelo Google Translate
func sanitizeLanguageCode(lang string) string {
	// Google Translate usa códigos simples como "pt" em vez de "pt-BR"
	parts := strings.Split(lang, "-")
	return strings.ToLower(parts[0])
}

// getHashKey cria uma chave de hash simplificada para textos longos
func getHashKey(text string) string {
	if len(text) <= 32 {
		return text
	}
	// Uma implementação simples para textos longos
	return fmt.Sprintf("%s...%s:%d", text[:16], text[len(text)-16:], len(text))
}

// GoogleTranslateAPIKey retorna a chave da API do Google Translate do ambiente
func GoogleTranslateAPIKey() string {
	return os.Getenv("GOOGLE_TRANSLATE_API_KEY")
}
