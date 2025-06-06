package i18n

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	// Bundle contém todas as mensagens traduzidas
	Bundle *i18n.Bundle
	// Idiomas suportados pela API
	SupportedLanguages = []string{"en", "pt-BR", "es", "fr", "de", "it"}
)

// InitLocales inicializa o sistema de internacionalização
func InitLocales() {
	// Cria um novo bundle com inglês como idioma base
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Carrega os arquivos de tradução
	if err := loadTranslationFiles(); err != nil {
		log.Printf("Aviso: Falha ao carregar traduções: %v", err)
		log.Println("API funcionará apenas com o idioma inglês")
	}
}

// LoadTranslationFiles carrega os arquivos de tradução do diretório i18n/locales
func loadTranslationFiles() error {
	localesDir := filepath.Join("i18n", "locales")
	// Verifica se o diretório existe
	if _, err := os.ReadDir(localesDir); err != nil {
		// Se o diretório não existe, cria-o e adiciona arquivos padrão de tradução
		if err := createDefaultTranslationFiles(localesDir); err != nil {
			return err
		}
	}

	// Carrega todos os arquivos de tradução
	for _, lang := range SupportedLanguages {
		filename := filepath.Join(localesDir, lang+".json")
		if _, err := Bundle.LoadMessageFile(filename); err != nil {
			log.Printf("Erro ao carregar tradução para %s: %v", lang, err)
		}
	}

	return nil
}

// createDefaultTranslationFiles cria os arquivos padrão de tradução se não existirem
func createDefaultTranslationFiles(localesDir string) error {
	// Cria o diretório se não existir
	if err := os.MkdirAll(localesDir, 0755); err != nil {
		return err
	}

	// Cria arquivos de tradução com conteúdo básico
	translations := map[string]map[string]string{
		"en": {
			"apod_title":        "Astronomy Picture of the Day",
			"apod_not_found":    "Document not found! Please check the date format (YYYY-MM-DD).",
			"search_no_results": "No documents found matching the search criteria",
		},
		"pt-BR": {
			"apod_title":        "Imagem Astronômica do Dia",
			"apod_not_found":    "Documento não encontrado! Por favor verifique o formato da data (AAAA-MM-DD).",
			"search_no_results": "Nenhum documento encontrado para os critérios de busca",
		},
		"es": {
			"apod_title":        "Imagen Astronómica del Día",
			"apod_not_found":    "¡Documento no encontrado! Por favor verifique el formato de la fecha (AAAA-MM-DD).",
			"search_no_results": "No se encontraron documentos que coincidan con los criterios de búsqueda",
		},
		"fr": {
			"apod_title":        "Image Astronomique du Jour",
			"apod_not_found":    "Document non trouvé! Veuillez vérifier le format de la date (AAAA-MM-JJ).",
			"search_no_results": "Aucun document trouvé correspondant aux critères de recherche",
		},
	}

	for lang, msgs := range translations {
		filename := filepath.Join(localesDir, lang+".json")

		// Converte para o formato esperado pelo go-i18n
		i18nMsgs := make(map[string]map[string]string)
		for id, msg := range msgs {
			i18nMsgs[id] = map[string]string{"other": msg}
		}

		// Serializa para JSON
		data, err := json.MarshalIndent(i18nMsgs, "", "  ")
		if err != nil {
			return err
		}
		// Escreve no arquivo
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// Localizer retorna um localizer para o idioma especificado
func Localizer(lang string) *i18n.Localizer {
	// Se o idioma não for especificado ou não for suportado, usa inglês
	if lang == "" {
		lang = "en"
	}

	// Verifica se o idioma é suportado
	supported := false
	for _, supportedLang := range SupportedLanguages {
		if strings.HasPrefix(lang, supportedLang) {
			supported = true
			break
		}
	}

	if !supported {
		lang = "en"
	}

	return i18n.NewLocalizer(Bundle, lang, "en")
}

// TranslateAPOD traduz os campos do APOD para o idioma solicitado
func TranslateAPOD(apodData map[string]interface{}, lang string) error {
	// Se não é um idioma suportado ou é inglês, retorna sem modificar
	if lang == "" || lang == "en" {
		return nil
	}

	// Inicializa o serviço de tradução se ainda não foi inicializado
	if currentService == nil {
		InitTranslationService()
	}

	// Traduz o título
	if title, ok := apodData["title"].(string); ok && title != "" {
		translatedTitle, err := TranslateText(title, lang)
		if err == nil {
			apodData["title"] = translatedTitle
		} else {
			log.Printf("Erro ao traduzir título: %v", err)
		}
	}

	// Traduz a explicação
	if explanation, ok := apodData["explanation"].(string); ok && explanation != "" {
		translatedExplanation, err := TranslateText(explanation, lang)
		if err == nil {
			apodData["explanation"] = translatedExplanation
		} else {
			log.Printf("Erro ao traduzir explicação: %v", err)
		}
	}

	// Outros campos podem ser traduzidos aqui se necessário
	// Por exemplo, copyright, etc.
	if copyright, ok := apodData["copyright"].(string); ok && copyright != "" {
		translatedCopyright, err := TranslateText(copyright, lang)
		if err == nil {
			apodData["copyright"] = translatedCopyright
		}
	}

	return nil
}

// Função auxiliar para obter o mínimo entre dois inteiros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
