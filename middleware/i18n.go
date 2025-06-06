package middleware

import (
	"context"
	"net/http"
	"strings"
)

// Chave de contexto para armazenar o idioma
type langKey struct{}

// LanguageDetector é um middleware que detecta o idioma preferido do usuário
func LanguageDetector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtém o idioma do header Accept-Language
		acceptLang := r.Header.Get("Accept-Language")

		// Obtém o idioma da query string 'lang' (tem precedência sobre o header)
		queryLang := r.URL.Query().Get("lang")
		if queryLang != "" {
			acceptLang = queryLang
		}

		// Extrai o código de idioma principal (por exemplo, "pt-BR" -> "pt")
		lang := "en" // padrão
		if acceptLang != "" {
			parts := strings.Split(acceptLang, ",")
			langParts := strings.Split(parts[0], ";") // Remove q-factor
			lang = strings.TrimSpace(langParts[0])
		}

		// Armazena o idioma no contexto da requisição
		ctx := context.WithValue(r.Context(), langKey{}, lang)

		// Chama o próximo handler com o contexto atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetLanguageFromContext extrai o idioma do contexto da requisição
func GetLanguageFromContext(ctx context.Context) string {
	lang, ok := ctx.Value(langKey{}).(string)
	if !ok {
		return "en" // idioma padrão
	}
	return lang
}
