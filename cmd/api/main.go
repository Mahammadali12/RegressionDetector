package main

import (
	"context"
	"log"
	"net/http"
	"regressiondetector/api"
	"regressiondetector/engine"
	"regressiondetector/internal/collector/config"
	"regressiondetector/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main(){

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v",err)
	}

	pool,err := pgxpool.New(context.Background(),cfg.ConnStr)
	if err != nil {
		log.Fatalf("Failed to create pool: %v",err)
	}

	s := store.NewStore(pool)
	d := engine.NewDetector(pool)
	


	http.HandleFunc("/api/v1/agent/payload",api.HandleIngest(cfg.APIToken,s,d))
	http.HandleFunc("/api/v1/anomalies",api.HandleGetAnomalies(s))
	http.Handle("/", http.FileServer(http.Dir("ui")))
	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080",nil))

}