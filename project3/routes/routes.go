package routes

import (
	"log"

	"github.com/gorilla/mux"

	"project3/cache"
	services "project3/services"
)

func NewRouter() *mux.Router {
	redisDB, err := cache.InitRedisClient()

	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()

	services := &services.Services{}
	services.SetRedis(redisDB)
	r.HandleFunc("/api/books", services.GetBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", services.GetBook).Methods("GET")
	r.HandleFunc("/api/books", services.CreateBook).Methods("POST")
	r.HandleFunc("/api/books/{id}", services.UpdateBook).Methods("PUT")
	r.HandleFunc("/api/books/{id}", services.DeleteBook).Methods("DELETE")

	return r
}
