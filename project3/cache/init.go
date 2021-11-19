package cache

import "github.com/go-redis/redis"

type RedisStore interface {
	GetBookByKey(key string) (string, error)
	GetAllBooksByKey() (string, error)
	SetBookValueByKey(key string, value string) error
	DelBookValueByKey(key string) error
	DelUserJWTByKey(key string) error
	AddUserByKey(key string, value string, jwttoken string) error
	AddJWTByKey(key string, value string) error
	CheckUserByKey(key string) bool
	CheckUserLoggedInByKey(key string) bool
	GetUserByKey(key string) (string, error)
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
