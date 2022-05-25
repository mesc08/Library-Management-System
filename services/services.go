package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"project3/cache"
	"project3/database"
	"project3/models"
	"project3/utils"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Services struct {
	RedisClient *redis.Client
	Mysql       *sql.DB
}

func (service *Services) SetDBConnection(redisClient *redis.Client, sqlDB *sql.DB) {
	service.RedisClient = redisClient
	service.Mysql = sqlDB
}

func (service *Services) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users models.User
	json.NewDecoder(r.Body).Decode(&users)

	users.UserId = utils.GenerateID()
	logrus.Println("Validate User Request Body")
	err := utils.ValidateUsersRequestBody(users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	logrus.Println("Check User exists")
	result, err := database.CheckUserIfExist(service.Mysql, users.Email)
	if err != nil {
		logrus.Println("Error ", err)
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("error checking to mysql db")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	logrus.Println("Result Obtained ", result)
	if result {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("db error: user already exist")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusInternalServerError})
		return
	}

	if users.Password != users.ConfirmPassword {
		w.WriteHeader(http.StatusFound)
		json.NewEncoder(w).Encode(&models.Response{Status: "user password does not match", StatusCode: http.StatusFound})
		return
	}
	validToken, err := utils.GenerateJWT(users.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("jwt token error: error in jwt")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	pwd, _ := utils.GeneratehashPassword(users.Password)
	users.Password = pwd
	err = cache.AddJWT(service.RedisClient, users.Email, validToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("db error: error in adding to db")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	err = database.AddUser(service.Mysql, users)
	logrus.Println("Error ", err)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("db error: error in adding to db")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Data: validToken, Status: "user successfully added", StatusCode: http.StatusOK})

}

func (service *Services) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users models.User
	_ = json.NewDecoder(r.Body).Decode(&users)

	err := utils.ValidateUsersRequestBody(users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result, err := database.CheckUserIfExist(service.Mysql, users.Email)
	if !result {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("db error: user does not exist")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	res, err := cache.GetJWT(service.RedisClient, users.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("Redis broke")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	if res == 1 {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("user already logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	dbuser, _ := database.GetUserByEmail(service.Mysql, users.Email)

	check := utils.CheckPasswordHash(users.Password, dbuser.Password)

	if !check {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("password error: password in db not matched")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	validtoken, _ := utils.GenerateJWT(users.Email)

	_ = cache.AddJWT(service.RedisClient, users.Email, validtoken)

	json.NewEncoder(w).Encode(&models.Response{Data: validtoken, Status: "user successfully added", StatusCode: http.StatusOK})
}

func (service *Services) LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]
	logrus.Println(jwtToken)
	data, err := cache.GetJWT(service.RedisClient, jwtToken)
	if err != nil {
		if err == redis.Nil {
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("json key not found")
			json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
			return
		}
	}
	if data == 1 {
		err := cache.DelJWT(service.RedisClient, jwtToken)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
			return
		}
		json.NewEncoder(w).Encode(&models.Response{Status: "user successfully loggedout", StatusCode: http.StatusOK})
	}
}

func (service *Services) AddBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]

	data, errors := cache.GetJWT(service.RedisClient, jwtToken)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	if data == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("user not logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := utils.ValidateBookRequestBody(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	book.ID = utils.GenerateID()

	errs := database.InsertBook(service.Mysql, book)
	if errs != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errs.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "successfully added into DB", StatusCode: http.StatusOK})
}

func (service *Services) GetBooks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]

	data, errors := cache.GetJWT(service.RedisClient, jwtToken)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	if data == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("user not logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	params := mux.Vars(r)
	book, error := database.GetBook(service.Mysql, params["id"])
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "no books found", StatusCode: http.StatusInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(&models.Response{Data: book, Status: "success", StatusCode: http.StatusOK})
}

func (service *Services) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]

	data, errors := cache.GetJWT(service.RedisClient, jwtToken)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	if data == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("user not logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	bookdata, error := database.GetAllBooksByKey(service.Mysql)

	if error != nil {
		logrus.Print(error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "error finding books", StatusCode: http.StatusInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&models.Response{Data: bookdata, Status: "success", StatusCode: http.StatusOK})
}

func (service *Services) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]

	data, errors := cache.GetJWT(service.RedisClient, jwtToken)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	if data == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("user not logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	params := mux.Vars(r)
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := utils.ValidateBookRequestBody(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	_, err = database.GetBook(service.Mysql, params["id"])
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			json.NewEncoder(w).Encode(&models.Response{Status: "no books found", StatusCode: http.StatusNoContent})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: "Server Error", StatusCode: http.StatusInternalServerError})
		return
	}
	err = database.UpdateBook(service.Mysql, params["id"], book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusInternalServerError})
		return
	}

	json.NewEncoder(w).Encode(&models.Response{Status: "successfully updated into DB", StatusCode: http.StatusOK})
}

func (service *Services) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := strings.Split(r.Header["Authorization"][0], " ")[1]

	data, errors := cache.GetJWT(service.RedisClient, jwtToken)

	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	if data == 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("user not logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	params := mux.Vars(r)

	errors = database.DeleteBook(service.Mysql, params["id"])

	if errors != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "book item successfully deleted", StatusCode: http.StatusOK})
}
