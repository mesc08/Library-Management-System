package services

import (
	"encoding/json"
	"net/http"
	"project3/cache"
	models "project3/models"

	"github.com/go-playground/validator"
	"github.com/rs/xid"

	"github.com/gorilla/mux"
)

type Services struct {
	redis *cache.RedisClient
}

func (service *Services) Redis() *cache.RedisClient {
	return service.redis
}

func (service *Services) SetRedis(redis *cache.RedisClient) {
	service.redis = redis
}

func (service *Services) GetBooks(w http.ResponseWriter, r *http.Request) {
	result, error := service.redis.GetAllByKey()
	w.Header().Set("Content-Type", "application/json")
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "error finding books", StatusCode: http.StatusInternalServerError})
		return
	}
	bookdata := []models.Book{}
	for data := range result {
		bytes := []byte(result[data])
		var book models.Book
		json.Unmarshal(bytes, &book)
		bookdata = append(bookdata, book)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&models.Response{Data: bookdata, Status: "success", StatusCode: http.StatusOK})
}

func (service *Services) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, error := service.redis.GetByKey(params["id"])
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "no books found", StatusCode: http.StatusInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	book := &models.Book{}
	json.Unmarshal([]byte(result), &book)
	json.NewEncoder(w).Encode(&models.Response{Data: book, Status: "success", StatusCode: http.StatusOK})
}

func (service *Services) CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := validateRequestBody(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	book.ID = generateID()

	data, err := json.Marshal(book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	errs := service.redis.SetValueByKey(book.ID, string(data))
	if errs != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errs.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "successfully added into DB", StatusCode: http.StatusOK})
}

func (service *Services) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := validateRequestBody(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result, error := service.redis.GetByKey(params["id"])
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "Server Error", StatusCode: http.StatusInternalServerError})
		return
	}
	if result == "" {
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(&models.Response{Status: "no books found", StatusCode: http.StatusNoContent})
		return
	}
	error = service.redis.DelValueByKey(params["id"])
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "Server Error", StatusCode: http.StatusInternalServerError})
		return
	}
	book.ID = params["id"]
	data, error := json.Marshal(book)

	if error != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: error.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	error = service.redis.SetValueByKey(string(book.ID), string(data))

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: error.Error(), StatusCode: http.StatusInternalServerError})
		return
	}

	json.NewEncoder(w).Encode(&models.Response{Status: "successfully updated into DB", StatusCode: http.StatusOK})
}

func (service *Services) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	errors := service.redis.DelValueByKey(params["id"])

	if errors != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "book item successfully deleted", StatusCode: http.StatusOK})
}

func validateRequestBody(book models.Book) error {
	validate := validator.New()
	err := validate.Struct(book)

	return err
}

func generateID() string {
	guid := xid.New()
	return guid.String()
}
