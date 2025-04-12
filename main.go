package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Global in-memory map for short code -> original URL
var urlStore = make(map[string]string)

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator once

	fmt.Println("Om Saravana Bhava 😂")

	http.HandleFunc("/", redirectHandler)         // handles GET /{code}
	http.HandleFunc("/shorten", shortenHandler)   // handles POST /shorten

	log.Println("Server is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	
}

// Handler for POST /shorten
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
	urlStore[code] = body.URL

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
	fmt.Println(urlStore);
	
}

// Handler for GET /{code}
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
	originalURL, ok := urlStore[code]
	if !ok {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Helper: generate a random 6-character string
// func randomString(n int) string {
// 	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
// 	b := make([]rune, n)
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 	}
// 	return string(b)
// }
