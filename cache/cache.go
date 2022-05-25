package cache

import (
	"context"
	"project3/config"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const jwtPrefix = "jwt:"

func InitRedisClient() (*redis.Client, error) {
	ctx := context.Background()
	logrus.Debugln("Trying redis connection")
	client := redis.NewClient(&redis.Options{
		Addr:     config.ViperConfig.RedisHost,
		Password: "",
		DB:       0,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		logrus.Errorln("Error connecting to redis database ", err)
		return nil, err
	}
	return client, nil
}

func AddJWT(redisClient *redis.Client, key, validToken string) error {
	if !strings.HasPrefix(key, jwtPrefix) {
		key = jwtPrefix + key
	}
	ctx := context.Background()
	err := redisClient.Set(ctx, key, validToken, time.Duration(time.Minute*30)).Err()
	return err
}

func DelJWT(redisClient *redis.Client, key string) error {
	if !strings.HasPrefix(key, jwtPrefix) {
		key = jwtPrefix + key
	}
	ctx := context.Background()
	err := redisClient.Del(ctx, key).Err()
	return err
}

func GetJWT(redisClient *redis.Client, key string) error {
	if !strings.HasPrefix(key, jwtPrefix) {
		key = jwtPrefix + key
	}
	ctx := context.Background()
	err := redisClient.Get(ctx, key).Err()
	return err
}
