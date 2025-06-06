package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// DeepLTranslateRequest representa o formato de requisição para a API DeepL
type DeepLTranslateRequest struct {
	Text       []string `json:"text"`
	SourceLang string   `json:"source_lang,omitempty"`
	TargetLang string   `json:"target_lang"`
	Formality  string   `json:"formality,omitempty"`
}

// DeepLTranslateResponse representa a resposta da API DeepL
type DeepLTranslateResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// DeepLClient implementa o serviço de tradução usando a API DeepL
type DeepLClient struct {
	apiKey     string
	httpClient *http.Client
	cache      *TranslationCache
	freeAPI    bool // Indica se está usando a API gratuita ou Pro
}

// NewDeepLClient cria um novo cliente para a API DeepL
func NewDeepLClient(apiKey string) *DeepLClient {
	// DeepL distingue API gratuita e Pro pelo prefixo da chave
	isFreeAPI := strings.HasPrefix(apiKey, "DeepL-Auth-Key ")

	return &DeepLClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:   NewTranslationCache(),
		freeAPI: isFreeAPI,
	}
}

// Translate implementa a interface TranslationService para DeepL
func (c *DeepLClient) Translate(text, sourceLang, targetLang string) (string, error) {
	// Verificar no cache primeiro
	cacheKey := fmt.Sprintf("deepl:%s:%s:%s", sourceLang, targetLang, getHashKey(text))
	if cachedText, found := c.cache.Get(cacheKey); found {
		return cachedText, nil
	}

	// Adapta o código de idioma para o formato esperado pelo DeepL
	targetLang = adaptLanguageForDeepL(targetLang)

	// Prepara a requisição
	reqBody := DeepLTranslateRequest{
		Text:       []string{text},
		TargetLang: targetLang,
	}

	// Apenas define o idioma de origem se for especificado
	if sourceLang != "" {
		reqBody.SourceLang = adaptLanguageForDeepL(sourceLang)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// URL da API dependendo do tipo (Free ou Pro)
	var url string
	if c.freeAPI {
		url = "https://api-free.deepl.com/v2/translate"
	} else {
		url = "https://api.deepl.com/v2/translate"
	}

	// Cria a requisição HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Define os headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

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
	var translateResp DeepLTranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translateResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	// Verifica se há traduções
	if len(translateResp.Translations) == 0 {
		return "", fmt.Errorf("nenhuma tradução retornada")
	}

	// Obtém o texto traduzido
	translatedText := translateResp.Translations[0].Text

	// Armazena no cache
	c.cache.Set(cacheKey, translatedText)

	return translatedText, nil
}

// adaptLanguageForDeepL converte códigos de idioma para o formato esperado pelo DeepL
func adaptLanguageForDeepL(lang string) string {
	// DeepL usa códigos como "PT-BR", "EN-US" (maiúsculos)
	parts := strings.Split(lang, "-")
	if len(parts) == 1 {
		return strings.ToUpper(parts[0])
	}

	// Para códigos compostos, capitaliza ambas as partes
	return strings.ToUpper(parts[0]) + "-" + strings.ToUpper(parts[1])
}

// DeepLAPIKey retorna a chave da API DeepL do ambiente
func DeepLAPIKey() string {
	return os.Getenv("DEEPL_API_KEY")
}
