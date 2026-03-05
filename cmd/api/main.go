package main

import (
	"log"
	"net/http"
	"regressiondetector/api"
	"regressiondetector/internal/collector/config"
)

func main(){

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v",err)
	}


	http.HandleFunc("/api/v1/agent/payload",api.HandleIngest(cfg.APIToken))
	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080",nil))

}