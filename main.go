package main

import (
	"log"
	"net/http"
	"project3/config"
	"project3/routes"

	"github.com/sirupsen/logrus"
)

func main() {
	config.InitConfig("./config.json")
	logrus.Println(config.ViperConfig)
	r := routes.NewRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}
