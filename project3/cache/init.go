package cache

import "github.com/go-redis/redis"

type RedisStore interface {
	GetByKey(key string) (string, error)
	GetAllByKey() (string, error)
	SetValueByKey(key string, value string) error
	DelValueByKey(key string) error
}

type RedisClient struct {
	*redis.Client
}

func InitRedisClient() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()

	return &RedisClient{client}, err
}
