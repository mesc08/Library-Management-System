package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"project3/cache"
	models "project3/models"
	"regexp"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
)

const secretkey = "srijanchakraborty1374"

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
	w.Header().Set("Content-Type", "application/json")
	jwtToken := r.Header["Authorization"][0]

	errors := CheckAuth(jwtToken, service)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	result, error := service.redis.GetAllBooksByKey()

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
	jwtToken := r.Header["Authorization"][0]

	errors := CheckAuth(jwtToken, service)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	params := mux.Vars(r)
	result, error := service.redis.GetBookByKey(params["id"])
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
	jwtToken := r.Header["Authorization"][0]

	errors := CheckAuth(jwtToken, service)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := validateBookRequestBody(book)
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
	errs := service.redis.SetBookValueByKey(book.ID, string(data))
	if errs != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errs.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "successfully added into DB", StatusCode: http.StatusOK})
}

func (service *Services) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := r.Header["Authorization"][0]

	errors := CheckAuth(jwtToken, service)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	params := mux.Vars(r)
	var book models.Book
	_ = json.NewDecoder(r.Body).Decode(&book)

	err := validateBookRequestBody(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result, error := service.redis.GetBookByKey(params["id"])
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
	error = service.redis.DelBookValueByKey(params["id"])
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
	error = service.redis.SetBookValueByKey(string(book.ID), string(data))

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: error.Error(), StatusCode: http.StatusInternalServerError})
		return
	}

	json.NewEncoder(w).Encode(&models.Response{Status: "successfully updated into DB", StatusCode: http.StatusOK})
}

func (service *Services) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := r.Header["Authorization"][0]

	errors := CheckAuth(jwtToken, service)
	if errors != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusInternalServerError})
		return
	}
	params := mux.Vars(r)

	errors = service.redis.DelBookValueByKey(params["id"])

	if errors != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: errors.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "book item successfully deleted", StatusCode: http.StatusOK})
}

func (service *Services) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users models.User
	_ = json.NewDecoder(r.Body).Decode(&users)

	err := validateUsersRequestBody(users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result := service.redis.CheckUserByKey(users.Email)
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
	validToken, err := GenerateJWT(users.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("jwt token error: error in jwt")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusInternalServerError})
		return
	}

	pwd, _ := GeneratehashPassword(users.Password)

	newusers := models.User{Email: users.Email, Password: pwd}
	data, _ := json.Marshal(newusers)
	err = service.redis.AddUserByKey(users.Email, string(data), validToken)

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

	err := validateUsersRequestBody(users)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result := service.redis.CheckUserByKey(users.Email)
	if !result {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("db error: user does not exist")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	result = service.redis.CheckUserLoggedInByKey(users.Email)

	if result {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("user already logged in")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	results, _ := service.redis.GetUserByKey(users.Email)

	dbuser := &models.User{}
	json.Unmarshal([]byte(results), &dbuser)

	check := CheckPasswordHash(users.Password, dbuser.Password)

	if !check {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("password error: password in db not matched")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	validtoken, _ := GenerateJWT(users.Email)

	_ = service.redis.AddJWTByKey(users.Email, validtoken)

	json.NewEncoder(w).Encode(&models.Response{Data: validtoken, Status: "user successfully added", StatusCode: http.StatusOK})
}

func (service *Services) LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwtToken := r.Header["Authorization"][0]

	data, keypresent := Authenticate(jwtToken)
	if !keypresent {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("json key not found")
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	err := service.redis.DelUserJWTByKey(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&models.Response{Status: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}
	json.NewEncoder(w).Encode(&models.Response{Status: "user successfully loggedout", StatusCode: http.StatusOK})
}

func CheckAuth(jwtToken string, service *Services) error {
	data, keypresent := Authenticate(jwtToken)
	if !keypresent {
		err := errors.New("json key not found")
		return err
	}

	result1 := service.redis.CheckUserByKey(data)
	result2 := service.redis.CheckUserLoggedInByKey(data)
	if !(result1 && result2) {
		err := errors.New("data not present")
		return err
	}
	return nil
}
func Authenticate(jwtToken string) (string, bool) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretkey), nil
	})
	if err != nil {
		return "", false
	}

	keypresent := false
	var data string
	for key, value := range claims {
		if key == "email" {
			keypresent = true
			data = value.(string)
			break
		}
	}

	return data, keypresent
}
func validateBookRequestBody(book models.Book) error {
	validate := validator.New()
	err := validate.Struct(book)

	return err
}

func validateUsersRequestBody(user models.User) error {
	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return err
	}
	resemail := isValidateEmail(user.Email)
	if !resemail {
		err := errors.New("emailid error: invalid email address")
		return err
	}
	respwd := isValidPassword(user.Password)
	if !respwd {
		err := errors.New("password error: invalid password entered")
		return err
	}
	if user.ConfirmPassword != "" {
		respwd := isValidPassword(user.Password)
		if !respwd {
			err := errors.New("password error: invalid password entered")
			return err
		}
	}
	return nil
}

func generateID() string {
	guid := xid.New()
	return guid.String()
}

func isValidPassword(password string) bool {
	var upp, low, num, sym bool
	var tot uint8
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			upp = true
			tot++
		case unicode.IsLower(char):
			low = true
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
			return false
		}
	}
	if !upp || !low || !num || !sym || tot < 8 {
		return false
	}
	return true
}

func isValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)

}

func GenerateJWT(email string) (string, error) {
	var mySigningkey = []byte(secretkey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["authorized"] = true
	claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

	tokenString, err := token.SignedString(mySigningkey)

	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
