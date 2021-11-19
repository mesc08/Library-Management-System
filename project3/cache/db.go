package cache

import (
	"encoding/json"
	"errors"
	"project3/models"
)

const Bookprefix = "book:"

func (redis *RedisClient) GetByKey(key string) (string, error) {
	bookKey := Bookprefix + key

	results, err := redis.Get(bookKey).Bytes()

	if err != nil {
		return "", err
	}

	result := string(results)
	return result, err
}

func (redis *RedisClient) SetValueByKey(key string, value string) error {

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

func (redis *RedisClient) DelValueByKey(key string) error {
	bookKey := Bookprefix + key
	err := redis.Del(bookKey).Err()
	return err
}

func (redis *RedisClient) GetAllByKey() ([]string, error) {
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
