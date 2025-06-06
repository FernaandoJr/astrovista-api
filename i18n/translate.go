package i18n

import (
	"astrovista-api/cache"
	"fmt"
	"log"
	"strings"
)

// TranslationService define a interface para serviços de tradução
type TranslationService interface {
	Translate(text, sourceLang, targetLang string) (string, error)
}

// mockTranslationService é uma implementação de simulação para desenvolvimento
type mockTranslationService struct{}

// Translate na implementação de simulação apenas adiciona um indicador de idioma
func (s *mockTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	if len(text) > 100 {
		// Para explicações longas, truncamos para simulação
		return fmt.Sprintf("%s... [Traduzido para %s]", text[:100], targetLang), nil
	}
	return fmt.Sprintf("%s [%s]", text, targetLang), nil
}

// googleTranslationService seria uma implementação real usando a API do Google Translate
type googleTranslationService struct {
	apiKey string
}

// Translate na implementação Google (esboço)
func (s *googleTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	// Aqui estaria o código para chamar a API do Google Translate
	// Por enquanto, apenas simulamos
	log.Printf("Simulando tradução de '%s' do '%s' para '%s'",
		truncateForLogging(text), sourceLang, targetLang)
	return text + " [Google Translated]", nil
}

// deepLTranslationService seria uma implementação real usando a API DeepL
type deepLTranslationService struct {
	apiKey string
}

// Translate na implementação DeepL (esboço)
func (s *deepLTranslationService) Translate(text, sourceLang, targetLang string) (string, error) {
	// Aqui estaria o código para chamar a API DeepL
	// Por enquanto, apenas simulamos
	log.Printf("Simulando tradução DeepL de '%s' do '%s' para '%s'",
		truncateForLogging(text), sourceLang, targetLang)
	return text + " [DeepL Translated]", nil
}

// Serviço de tradução atual
var currentService TranslationService

// InitTranslationService inicializa o serviço de tradução apropriado
func InitTranslationService() {
	// Verificar qual serviço usar com base em variáveis de ambiente
	if apiKey := GoogleTranslateAPIKey(); apiKey != "" {
		log.Println("Usando Google Translate para traduções")
		googleClient := NewGoogleTranslateClient(apiKey)

		// Habilita cache Redis se disponível
		if cache.Client != nil {
			log.Println("Cache Redis habilitado para traduções")
			googleClient.cache.EnableRedisCache()
		}

		currentService = googleClient
	} else if apiKey := DeepLAPIKey(); apiKey != "" {
		log.Println("Usando DeepL para traduções")
		deepLClient := NewDeepLClient(apiKey)

		// Habilita cache Redis se disponível
		if cache.Client != nil {
			log.Println("Cache Redis habilitado para traduções")
			deepLClient.cache.EnableRedisCache()
		}

		currentService = deepLClient
	} else {
		log.Println("Nenhuma API de tradução configurada, usando simulação")
		currentService = &mockTranslationService{}
	}
}

// TranslateText traduz o texto para o idioma alvo
func TranslateText(text, targetLang string) (string, error) {
	if currentService == nil {
		InitTranslationService()
	}

	// Se o idioma alvo for inglês ou vazio, não traduzimos
	if targetLang == "" || targetLang == "en" {
		return text, nil
	}

	// Trunca texto muito longo (só para log, não para tradução real)
	logText := truncateForLogging(text)
	log.Printf("Traduzindo texto: '%s' para '%s'", logText, targetLang)

	// Assumimos inglês como idioma fonte
	return currentService.Translate(text, "en", targetLang)
}

// Método auxiliar para truncar texto longo nos logs
func truncateForLogging(text string) string {
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

// TryTranslate tenta traduzir um texto, retornando o original em caso de erro
func TryTranslate(text string, targetLang string) string {
	if targetLang == "en" || strings.TrimSpace(text) == "" {
		return text
	}

	translated, err := TranslateText(text, targetLang)
	if err != nil {
		log.Printf("Erro ao traduzir texto: %v", err)
		return text
	}
	return translated
}
