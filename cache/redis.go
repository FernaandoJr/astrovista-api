package cache

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// Client é o cliente Redis compartilhado
	Client *redis.Client
	// DefaultExpiration é o tempo padrão de expiração para itens em cache (24 horas)
	DefaultExpiration = 24 * time.Hour
)

// Connect estabelece a conexão com o Redis
func Connect() {
	// Verifica se há um URL do Redis nas variáveis de ambiente (para uso em produção)
	redisURL := os.Getenv("REDIS_URL")

	// Se não houver, usa um padrão para desenvolvimento local
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	// Configuração do cliente Redis
	Client = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: os.Getenv("REDIS_PASSWORD"), // Sem senha se não estiver definido
		DB:       0,                           // Usar banco de dados 0
	})
	// Verifica se a conexão está funcionando
	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Aviso: Não foi possível conectar ao Redis: %v", err)
		log.Println("O cache será desativado. Para ativar o cache, instale o Redis e execute-o em localhost:6379")
		log.Println("A API continuará funcionando normalmente, mas sem o benefício do cache")
		Client = nil
		return
	}

	log.Println("Conexão com Redis estabelecida com sucesso")
}

// Set armazena um item no cache
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return nil // Cache desativado
	}

	// Convertendo para JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Armazenando no Redis
	return Client.Set(ctx, key, data, expiration).Err()
}

// Get recupera um item do cache
func Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if Client == nil {
		return false, nil // Cache desativado
	}

	// Buscando do Redis
	data, err := Client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Item não encontrado no cache
		return false, nil
	} else if err != nil {
		// Erro ao acessar o Redis
		return false, err
	}

	// Convertendo de JSON para o tipo de destino
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Delete remove um item do cache
func Delete(ctx context.Context, key string) error {
	if Client == nil {
		return nil // Cache desativado
	}

	return Client.Del(ctx, key).Err()
}

// Clear limpa todo o cache
func Clear(ctx context.Context) error {
	if Client == nil {
		return nil // Cache desativado
	}

	return Client.FlushAll(ctx).Err()
}
