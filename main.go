package main

import (
	"log"
	"net/http"

	"github.com/snykk/simple-go-elasticsearch/elastic"
	"github.com/snykk/simple-go-elasticsearch/handlers"

	"github.com/gorilla/mux"
)

func main() {
	products, err := elastic.LoadProductsFromFile("data/products.json")
	if err != nil {
		log.Fatalf("Failed to load products: %v", err)
	}

	err = elastic.IndexProducts(products)
	if err != nil {
		log.Fatalf("Failed to index products: %v", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/search", handlers.SearchHandler).Methods("GET")
	router.HandleFunc("/stats", handlers.StatsHandler).Methods("GET")
	router.HandleFunc("/import", handlers.ImportHandler).Methods("POST")
	router.HandleFunc("/update", handlers.BatchUpdateHandler).Methods("PUT")
	router.HandleFunc("/aggregation", handlers.AggregationHandler).Methods("GET")
	router.HandleFunc("/suggest", handlers.SuggestHandler)

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
