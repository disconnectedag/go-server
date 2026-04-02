package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/disconnectedag/go-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey	   string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("DB_URL must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	const port = "8080"
	dbQueries := database.New(db)
	apiCfg := apiConfig{db: dbQueries, platform: platform, jwtSecret: jwtSecret, polkaKey: polkaKey}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	mux.Handle("/app/", http.StripPrefix("/app", fsHandler))
	mux.HandleFunc("GET /api/healthz", hcHandler)

	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getOneChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)

	mux.HandleFunc("POST /api/users", apiCfg.createUsersHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginUserHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeTokenHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.updatePasswordHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.upgradeToRedHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	s := &http.Server{Addr: ":" + port, Handler: mux}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(s.ListenAndServe())
}
