package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://postgres:081204@localhost:5432/url_shortener")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer db.Close()
	// db, err = pgxpool.New(context.Background(), "postgres://postgres:081204@localhost:5432/url_shortener")
	
	http.HandleFunc("/", redirectHandler)         // handles GET /{code}
	http.HandleFunc("/shorten", shortenHandler)   // handles POST /shorten

	log.Println("Server is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", 405)
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

	// Generate a short code
	code := fmt.Sprint(rand.Intn(99999))
	//generate a reandom integer and convert it a string
	_, err = db.Exec(context.Background(),
	"INSERT INTO url_mappings (short_code,original_url) VALUES ($1,$2)",
	code, body.URL)
if err != nil {
	http.Error(w, "Database insert failed", http.StatusInternalServerError)
	fmt.Println(err)
	return
}
	type Response struct {
		ShortURL string `json:"short_url"`
	}

	resp := Response{
		ShortURL: "http://localhost:8080/" + code,
	}
	w.Header().Set("Content-Type", "application/json")
	//actually this above command for more strucutre coding to know the written type is json
	//which may or may not be required but its a good way of practising
	json.NewEncoder(w).Encode(resp)
	//add json to the write (w) method then insert the json data(resp) to the "w" created json
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	/*	Part	What it gives you
	r.URL.Scheme	"http"
	r.Host	"localhost:8080"
	r.URL.Path	"/abc123"
	r.URL.String()	"/abc123"
	r.RequestURI	"/abc123"
	*/
		code := r.URL.Path[1:] // remove leading "/"
		code2:=r.URL.Path;
		fmt.Println(code);
		fmt.Println(code2);
		
		var originalURL string
err := db.QueryRow(context.Background(),
	"SELECT original_url from url_mappings where short_code=$1", code).Scan(&originalURL)
if err != nil {
	http.Error(w, "Short URL not found", http.StatusNotFound)
	return
}
		http.Redirect(w, r, originalURL, http.StatusFound)
	}