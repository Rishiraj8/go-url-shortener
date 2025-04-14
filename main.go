package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	// Load .env (optional in cloud, helpful locally)
	_ = godotenv.Load()

	// Get DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// Connect to DB
	var err error
	db, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer db.Close()

	// Handlers with CORS support
	http.HandleFunc("/", withCORS(redirectHandler))
	http.HandleFunc("/shorten", withCORS(shortenHandler))

	// Get PORT from env (Render sets PORT=10000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}
	log.Println("Server is listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.URL == "" {
		http.Error(w, "Invalid JSON or missing 'url'", http.StatusBadRequest)
		return
	}

	// Generate short code
	code := fmt.Sprintf("%05d", rand.Intn(100000))

	// Insert into DB
	_, err = db.Exec(context.Background(),
		"INSERT INTO url_mappings (short_code, original_url) VALUES ($1, $2)",
		code, body.URL)
	if err != nil {
		http.Error(w, "Database insert failed", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Construct short URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	// fmt.Print(baseURL);

	resp := struct {
		ShortURL string `json:"short_url"`
	}{
		ShortURL: baseURL + "/" + code,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:] // remove leading "/"
	fmt.Println("Redirect code:", code)

	var originalURL string
	err := db.QueryRow(context.Background(),
		"SELECT original_url FROM url_mappings WHERE short_code = $1", code).Scan(&originalURL)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
