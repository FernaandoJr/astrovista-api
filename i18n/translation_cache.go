package i18n

import (
	"astrovista-api/cache"
	"sync"
	"time"
)

// TranslationCache implementa um cache em memória para traduções
// para reduzir requisições à API de tradução
type TranslationCache struct {
	translations map[string]cacheEntry
	mutex        sync.RWMutex
	maxSize      int
	expiration   time.Duration
	// Funções opcionais para integração com Redis
	redisEnabled bool
}

// cacheEntry representa uma entrada no cache com informações de expiração
type cacheEntry struct {
	value      string
	expiration time.Time
}

// NewTranslationCache cria uma nova instância do cache de traduções
func NewTranslationCache() *TranslationCache {
	return &TranslationCache{
		translations: make(map[string]cacheEntry),
		maxSize:      1000,           // Limita o número máximo de entradas
		expiration:   24 * time.Hour, // Tempo padrão de expiração
		redisEnabled: false,          // Por padrão, não usa Redis
	}
}

// EnableRedisCache configura o cache para usar também o Redis
func (c *TranslationCache) EnableRedisCache() {
	c.redisEnabled = cache.Client != nil
}

// Get recupera uma tradução do cache
func (c *TranslationCache) Get(key string) (string, bool) {
	// Se Redis estiver habilitado, tenta buscar do cache Redis primeiro
	if c.redisEnabled && cache.Client != nil {
		redisCache := NewRedisTranslationCache()
		if value, found := redisCache.Get(key); found {
			return value, true
		}
	}

	// Caso contrário, usa o cache em memória
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, found := c.translations[key]
	if !found {
		return "", false
	}

	// Verifica se a entrada expirou
	if time.Now().After(entry.expiration) {
		// Poderia fazer a remoção aqui, mas para evitar
		// lock upgrading, deixamos para o processo de limpeza periódica
		return "", false
	}

	return entry.value, true
}

// Set armazena uma tradução no cache
func (c *TranslationCache) Set(key string, value string) {
	// Se Redis estiver habilitado, armazena também no Redis
	if c.redisEnabled && cache.Client != nil {
		redisCache := NewRedisTranslationCache()
		redisCache.Set(key, value)
	}

	// Também armazena no cache em memória para acesso rápido
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Se o cache estiver cheio e a chave não existir, fazemos uma limpeza
	if len(c.translations) >= c.maxSize && c.translations[key].value == "" {
		c.cleanupLocked()
	}

	c.translations[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.expiration),
	}
}

// Clear limpa todo o cache
func (c *TranslationCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.translations = make(map[string]cacheEntry)
}

// cleanupLocked remove entradas expiradas do cache (assume que o lock já está obtido)
func (c *TranslationCache) cleanupLocked() {
	now := time.Now()

	// Remove entradas expiradas
	for key, entry := range c.translations {
		if now.After(entry.expiration) {
			delete(c.translations, key)
		}
	}

	// Se ainda estiver muito grande, remove as mais antigas
	// Esta é uma implementação simples, não é LRU completo
	if len(c.translations) >= c.maxSize {
		// Removemos cerca de 25% do cache para não ter que fazer isso frequentemente
		toRemove := c.maxSize / 4
		removed := 0

		for key := range c.translations {
			delete(c.translations, key)
			removed++
			if removed >= toRemove {
				break
			}
		}
	}
}
