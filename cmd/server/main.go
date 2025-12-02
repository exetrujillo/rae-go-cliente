package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"rae-client/pkg/rae"
)

func main() {
	client := rae.NewClient()

	http.HandleFunc("/wotd", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetWordOfTheDay()
		handleResponse(w, resp, err)
	})

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetRandomWord()
		handleResponse(w, resp, err)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("w")
		if query == "" {
			http.Error(w, "Falta el parámetro 'w'", http.StatusBadRequest)
			return
		}
		resp, err := client.SearchWord(query)
		handleResponse(w, resp, err)
	})

	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Falta el parámetro 'id'", http.StatusBadRequest)
			return
		}
		
		withConjugations := r.URL.Query().Get("conjugaciones") == "true" || r.URL.Query().Get("conjugations") == "true"
		
		resp, err := client.FetchWord(id, withConjugations)
		handleResponse(w, resp, err)
	})

	http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Falta el parámetro 'q'", http.StatusBadRequest)
			return
		}
		resp, err := client.KeyQuery(query)
		handleResponse(w, resp, err)
	})

	http.HandleFunc("/anagram", func(w http.ResponseWriter, r *http.Request) {
		word := r.URL.Query().Get("w")
		if word == "" {
			http.Error(w, "Falta el parámetro 'w'", http.StatusBadRequest)
			return
		}
		resp, err := client.SearchAnagram(word)
		handleResponse(w, resp, err)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor iniciando en puerto %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleResponse(w http.ResponseWriter, body []byte, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Siempre establecer tipo de contenido JSON ya que estamos limpiando JSONP
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(body)
}
