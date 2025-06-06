package i18n

import (
	"astrovista-api/cache"
	"context"
	"fmt"
	"log"
	"time"
)

// RedisTranslationCache implementa um cache de traduções usando Redis
// para persistir traduções entre reinicializações do servidor
type RedisTranslationCache struct {
	// Prefixo para evitar colisões com outras chaves no Redis
	prefix string
	// Tempo de expiração para traduções armazenadas
	expiration time.Duration
}

// NewRedisTranslationCache cria uma nova instância do cache Redis para traduções
func NewRedisTranslationCache() *RedisTranslationCache {
	return &RedisTranslationCache{
		prefix:     "translation:",
		expiration: 30 * 24 * time.Hour, // 30 dias de cache
	}
}

// Get recupera uma tradução do cache Redis
func (c *RedisTranslationCache) Get(key string) (string, bool) {
	// Se Redis não está configurado, retorna não encontrado
	if cache.Client == nil {
		return "", false
	}

	ctx := context.Background()
	redisKey := c.prefix + key

	var result string
	found, err := cache.Get(ctx, redisKey, &result)
	if err != nil {
		log.Printf("Erro ao acessar cache Redis para tradução: %v", err)
		return "", false
	}

	return result, found
}

// Set armazena uma tradução no cache Redis
func (c *RedisTranslationCache) Set(key string, value string) {
	// Se Redis não está configurado, não faz nada
	if cache.Client == nil {
		return
	}

	ctx := context.Background()
	redisKey := c.prefix + key

	if err := cache.Set(ctx, redisKey, value, c.expiration); err != nil {
		log.Printf("Erro ao armazenar tradução no cache Redis: %v", err)
	}
}

// Clear remove todas as traduções do cache Redis com o prefixo especificado
func (c *RedisTranslationCache) Clear() {
	// Se Redis não está configurado, não faz nada
	if cache.Client == nil {
		return
	}

	ctx := context.Background()
	// Usa o comando KEYS para encontrar todas as chaves com o prefixo (menos eficiente mas mais simples)
	pattern := fmt.Sprintf("%s*", c.prefix)
	keys, err := cache.Client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Erro ao buscar chaves de tradução no Redis: %v", err)
		return
	}

	if len(keys) > 0 {
		if err := cache.Client.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Erro ao deletar chaves de tradução do Redis: %v", err)
		}
	}
}
