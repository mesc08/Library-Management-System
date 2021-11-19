package cache

import (
	"encoding/json"
	"errors"
	"project3/models"
)

const Bookprefix = "book:"
const UserPrefix = "user:"
const JWTPrefix = "jwt:"

func (redis *RedisClient) GetBookByKey(key string) (string, error) {
	bookKey := Bookprefix + key

	results, err := redis.Get(bookKey).Bytes()

	if err != nil {
		return "", err
	}

	result := string(results)
	return result, err
}

func (redis *RedisClient) SetBookValueByKey(key string, value string) error {

	bookKey := Bookprefix + key
	var addbook models.Book

	json.Unmarshal([]byte(value), &addbook)
	patternbookey := Bookprefix + "*"
	result := redis.Keys(patternbookey)
	for _, key := range result.Val() {

		result, err := redis.Get(key).Bytes()

		if err != nil {
			return err
		}

		var book models.Book

		json.Unmarshal(result, &book)

		if addbook.Title == book.Title {
			err := errors.New("db error: Book already exist")
			return err
		}

	}
	err := redis.Set(bookKey, value, 0).Err()
	return err
}

func (redis *RedisClient) DelBookValueByKey(key string) error {
	bookKey := Bookprefix + key
	err := redis.Del(bookKey).Err()
	return err
}

func (redis *RedisClient) GetAllBooksByKey() ([]string, error) {
	bookKey := Bookprefix + "*"

	result := redis.Keys(bookKey)
	empty := []string{}
	response := []string{}
	for _, key := range result.Val() {

		results, err := redis.Get(key).Bytes()

		if err != nil {
			return empty, err
		}

		result := string(results)
		response = append(response, result)
	}
	return response, nil
}

func (redis *RedisClient) AddUserByKey(key string, value string, jwttoken string) error {

	userkey := UserPrefix + key
	jwtkey := JWTPrefix + key

	errs := redis.Set(userkey, value, 0).Err()
	if errs != nil {
		return errs
	}

	errs = redis.Set(jwtkey, jwttoken, 0).Err()
	if errs != nil {
		return errs
	}
	return nil
}

func (redis *RedisClient) AddJWTByKey(key string, value string) error {
	jwtkey := JWTPrefix + key

	errs := redis.Set(jwtkey, value, 0).Err()
	if errs != nil {
		return errs
	}
	return nil

}

func (redis *RedisClient) CheckUserByKey(key string) bool {
	user := UserPrefix + key

	Result, err := redis.Exists(user).Result()

	if err != nil {
		panic(err)
	}
	if Result == 1 {
		return true
	}
	return false
}

func (redis *RedisClient) CheckUserLoggedInByKey(key string) bool {
	user := JWTPrefix + key

	Result, err := redis.Exists(user).Result()

	if err != nil {
		panic(err)
	}
	if Result == 1 {
		return true
	}
	return false
}

func (redis *RedisClient) DelUserJWTByKey(key string) error {

	jwtkey := JWTPrefix + key
	err := redis.Del(jwtkey).Err()
	return err
}

func (redis *RedisClient) GetUserByKey(key string) (string, error) {
	userKey := UserPrefix + key

	results, err := redis.Get(userKey).Bytes()

	if err != nil {
		return "", err
	}

	result := string(results)
	return result, err
}
