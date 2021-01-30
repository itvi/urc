package main

import (
	"log"
	"net/http"
	"project/internal/handler"
)

func main() {

	config, db := handler.Config()
	defer db.Close()

	server := &http.Server{
		Addr:    "localhost:9000",
		Handler: config.Route(),
	}

	log.Println("Starting...", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
