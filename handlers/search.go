package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snykk/simple-go-elasticsearch/models"
	"github.com/snykk/simple-go-elasticsearch/services"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	size := r.URL.Query().Get("size")
	if size == "" {
		size = "10"
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "price" // Default sorting by price
	}

	priceMin := r.URL.Query().Get("price_min")
	priceMax := r.URL.Query().Get("price_max")
	stockMin := r.URL.Query().Get("stock_min")
	stockMax := r.URL.Query().Get("stock_max")

	products, err := services.SearchProductsWithPaginationSortingAndFilter(query, page, size, sort, priceMin, priceMax, stockMin, stockMax)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search products: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := services.GetIndexStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func ImportHandler(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	err := services.IndexProducts(products)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to index products: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Products successfully indexed"))
}
