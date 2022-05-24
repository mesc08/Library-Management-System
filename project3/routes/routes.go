package routes

import (
	"project3/cache"
	"project3/database"
	services "project3/services"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewRouter() *mux.Router {
	redisClient, err := cache.InitRedisClient()
	if err != nil {
		logrus.Errorln("Error connecting to redis ", err)
	}
	logrus.Println("Successfully connected to redis")
	mysql, err := database.ConnecToMysql()
	if err != nil {
		logrus.Errorln("Error connecting to mysql ", err)
	}
	logrus.Println("Successfully connected to mysql")
	r := mux.NewRouter()

	services := &services.Services{}
	services.SetDBConnection(redisClient, mysql)
	r.HandleFunc("/signup", services.RegisterUser).Methods("POST")
	r.HandleFunc("/signin", services.LoginUser).Methods("GET")
	r.HandleFunc("/signout", services.LogoutUser).Methods("GET")
	r.HandleFunc("/library/books", services.GetAllBooks).Methods("GET")
	r.HandleFunc("/library/books/{id}", services.GetBooks).Methods("GET")
	r.HandleFunc("/library/books", services.AddBook).Methods("POST")
	r.HandleFunc("/library/books/{id}", services.UpdateBook).Methods("PUT")
	r.HandleFunc("/library/books/{id}", services.DeleteBook).Methods("DELETE")
	return r
}
