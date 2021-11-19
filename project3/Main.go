package main

import (
	"log"
	"net/http"
	routes "project3/routes"
)

func main() {
	r := routes.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", r))
}
