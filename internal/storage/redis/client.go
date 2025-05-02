package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/pkg/sl"
)

// ClientStorage реализация хранилища клиентов в Redis
type ClientStorage struct {
	client *redis.Client
	cfg    *config.RateLimiter
	prefix string
}

func NewClientStorage(cfg *config.RateLimiter, redisConfig *config.Redis) *ClientStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis connection error", err.Error())
	}

	return &ClientStorage{
		client: client,
		cfg:    cfg,
		prefix: redisConfig.Prefix,
	}
}

func (rs *ClientStorage) getKey(clientID string) string {
	return fmt.Sprintf("%sclient:%s", rs.prefix, clientID)
}

func (rs *ClientStorage) GetAllClientIDs() ([]string, error) {
	ctx := context.Background()
	pattern := fmt.Sprintf("%sclient:*", rs.prefix)

	keys, err := rs.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// Убираем префикс из ключей для получения ID клиентов
	ids := make([]string, 0, len(keys))
	prefixLen := len(rs.prefix) + 7 // длина "client:"

	for _, key := range keys {
		if len(key) > prefixLen {
			ids = append(ids, key[prefixLen:])
		}
	}

	return ids, nil
}

func (rs *ClientStorage) GetClient(clientID string) (*model.Client, bool) {
	ctx := context.Background()
	key := rs.getKey(clientID)

	data, err := rs.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		slog.Error("Ошибка получения данных из Redis", sl.Err(err))
		return nil, false
	}

	var client model.Client
	if err := json.Unmarshal(data, &client); err != nil {
		slog.Error("Ошибка десериализации данных клиента", sl.Err(err))
		return nil, false
	}

	return &client, true
}

func (rs *ClientStorage) SaveClient(clientID string, client *model.Client) {
	ctx := context.Background()
	key := rs.getKey(clientID)

	data, err := json.Marshal(client)
	if err != nil {
		slog.Error("Ошибка сериализации данных клиента", sl.Err(err))
		return
	}

	// Устанавливаем время жизни ключа для автоматической очистки неактивных клиентов
	err = rs.client.Set(ctx, key, data, 48*time.Hour).Err()
	if err != nil {
		slog.Error("Ошибка сохранения данных в Redis", sl.Err(err))
	}
}

func (rs *ClientStorage) DeleteClient(clientID string) error {
	ctx := context.Background()
	key := rs.getKey(clientID)

	return rs.client.Del(ctx, key).Err()
}

func (rs *ClientStorage) SetClientRate(clientID string, rate int) error {
	client, exists := rs.GetClient(clientID)

	if !exists {
		client = &model.Client{
			Rate:       rate,
			Capacity:   rs.cfg.Capacity,
			LastRefill: time.Now(),
			IP:         clientID,
			Tokens:     rs.cfg.Capacity,
		}
	} else {
		client.Rate = rate
	}

	rs.SaveClient(clientID, client)
	return nil
}

func (rs *ClientStorage) SetClientCapacity(clientID string, capacity int) error {
	client, exists := rs.GetClient(clientID)

	if !exists {
		client = &model.Client{
			Rate:       rs.cfg.DefaultRate,
			Capacity:   capacity,
			LastRefill: time.Now(),
			IP:         clientID,
			Tokens:     capacity,
		}
	} else {
		client.Capacity = capacity
		client.Tokens = capacity
	}

	rs.SaveClient(clientID, client)
	return nil
}

func (rs *ClientStorage) Close() error {
	return rs.client.Close()
}
