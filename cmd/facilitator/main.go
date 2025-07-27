package main

import (
	"log"
	"net/http"
	"x402/internal/config"

	"x402/internal/api"
)

func main() {
	srvCfg, err := config.Load("./config.env")
	if err != nil {
		log.Fatal(err)
	}

	server, err := api.NewServer(srvCfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", server))
}
