package i18n

import (
	"astrovista-api/i18n"
	"os"
	"testing"
)

// TestTranslationService testa os diferentes serviços de tradução
func TestTranslationService(t *testing.T) {
	// Backup das variáveis de ambiente originais
	originalGoogleKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	originalDeepLKey := os.Getenv("DEEPL_API_KEY")

	// Limpa as variáveis ao final do teste
	defer func() {
		os.Setenv("GOOGLE_TRANSLATE_API_KEY", originalGoogleKey)
		os.Setenv("DEEPL_API_KEY", originalDeepLKey)
	}()

	// Testa o serviço Mock (padrão quando não há chaves configuradas)
	t.Run("MockTranslationService", func(t *testing.T) {
		// Limpa as variáveis para garantir que o mock será usado
		os.Setenv("GOOGLE_TRANSLATE_API_KEY", "")
		os.Setenv("DEEPL_API_KEY", "")

		// Reinicializa o serviço
		i18n.InitTranslationService()

		text := "Hello, world!"
		translated, err := i18n.TranslateText(text, "pt-BR")

		if err != nil {
			t.Errorf("Erro ao traduzir texto: %v", err)
		}

		// Verifica se o texto foi modificado (mock adiciona [pt-BR] ao final)
		if translated == text {
			t.Errorf("Texto não foi traduzido pelo mock")
		}
	})

	// Testes integrando com APIs reais podem ser adicionados aqui,
	// mas não seriam executados em CI automaticamente (precisariam de chaves reais)
}

// TestTranslateAPOD testa a tradução de campos em um documento APOD
func TestTranslateAPOD(t *testing.T) {
	// Criar um APOD fictício para testar
	apodData := map[string]interface{}{
		"title":       "Amazing Galaxy",
		"explanation": "This is a beautiful galaxy far away.",
	}

	// Reinicia o serviço de tradução para usar o mock
	os.Setenv("GOOGLE_TRANSLATE_API_KEY", "")
	os.Setenv("DEEPL_API_KEY", "")
	i18n.InitTranslationService()

	// Testa com inglês (não deve modificar)
	i18n.TranslateAPOD(apodData, "en")
	if apodData["title"] != "Amazing Galaxy" {
		t.Errorf("O texto em inglês não deveria ser modificado")
	}

	// Testa com português
	i18n.TranslateAPOD(apodData, "pt-BR")

	// Com o mock, deve ter adicionado [pt-BR] ao título
	if title, ok := apodData["title"].(string); !ok || title == "Amazing Galaxy" {
		t.Errorf("O texto deveria ter sido traduzido, mas continua: %v", apodData["title"])
	}
}

// TestTranslationCache testa o funcionamento do cache de traduções
func TestTranslationCache(t *testing.T) {
	cache := i18n.NewTranslationCache()

	// Testa adicionar e recuperar do cache
	testKey := "test:en:pt-BR:hello"
	testValue := "olá"

	// Inicialmente o valor não deve existir
	_, found := cache.Get(testKey)
	if found {
		t.Errorf("Valor não deveria existir no cache ainda")
	}

	// Adiciona ao cache
	cache.Set(testKey, testValue)

	// Agora deve ser encontrado
	value, found := cache.Get(testKey)
	if !found {
		t.Errorf("Valor deveria existir no cache após Set")
	}

	if value != testValue {
		t.Errorf("Valor recuperado (%s) não corresponde ao valor armazenado (%s)", value, testValue)
	}

	// Testa limpeza
	cache.Clear()
	_, found = cache.Get(testKey)
	if found {
		t.Errorf("Valor ainda existe no cache após Clear")
	}
}
