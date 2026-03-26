package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const port = "8080"
	apiCfg := apiConfig{}
	
	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	mux.Handle("/app/", http.StripPrefix("/app", fsHandler))
	mux.HandleFunc("GET /healthz", hcHandler)
	mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /reset", apiCfg.resetHandler)

	s := &http.Server{Addr: ":" + port, Handler: mux}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(s.ListenAndServe())
}
