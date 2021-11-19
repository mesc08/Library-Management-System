package routes

import (
	"log"

	"project3/cache"
	services "project3/services"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	redisDB, err := cache.InitRedisClient()

	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()

	services := &services.Services{}
	services.SetRedis(redisDB)
	r.HandleFunc("/signup", services.RegisterUser).Methods("POST")
	r.HandleFunc("/signin", services.LoginUser).Methods("GET")
	r.HandleFunc("/signout", services.LogoutUser).Methods("GET")
	r.HandleFunc("/library/books", services.GetBooks).Methods("GET")
	r.HandleFunc("/library/books/{id}", services.GetBook).Methods("GET")
	r.HandleFunc("/library/books", services.CreateBook).Methods("POST")
	r.HandleFunc("/library/books/{id}", services.UpdateBook).Methods("PUT")
	r.HandleFunc("/library/books/{id}", services.DeleteBook).Methods("DELETE")

	return r
}
