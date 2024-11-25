package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	elastic "github.com/snykk/simple-go-elasticsearch/elastic"
	"github.com/snykk/simple-go-elasticsearch/models"
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

	category := r.URL.Query().Get("category")

	priceMin := r.URL.Query().Get("price_min")
	priceMax := r.URL.Query().Get("price_max")
	stockMin := r.URL.Query().Get("stock_min")
	stockMax := r.URL.Query().Get("stock_max")
	createdAtMin := r.URL.Query().Get("created_at_min")
	createdAtMax := r.URL.Query().Get("created_at_max")

	products, err := elastic.SearchProductsWithPaginationSortingAndFilter(query, page, size, sort, priceMin, priceMax, stockMin, stockMax, category, createdAtMin, createdAtMax)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search products: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Search query: %s, Page: %s, Category: %s, Price range: %s-%s, Stock range: %s-%s, CreatedAt range: %s-%s, IP: %s\n",
		query, page, category, priceMin, priceMax, stockMin, stockMax, createdAtMin, createdAtMax, r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func ImportHandler(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	for _, product := range products {
		if product.Category == "" {
			http.Error(w, "All products must include a 'category'", http.StatusBadRequest)
			return
		}
	}

	err := elastic.IndexProducts(products)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to index products: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Products successfully indexed"))
}

func BatchUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var updates []models.Product
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	for _, update := range updates {
		if update.Category == "" {
			http.Error(w, "All products in the update must include a 'category'", http.StatusBadRequest)
			return
		}
	}

	err := elastic.UpdateProducts(updates)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update products: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Products updated successfully"))
}

func AggregationHandler(w http.ResponseWriter, r *http.Request) {
	aggregation, err := elastic.AggregateProductData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to perform aggregation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aggregation)
}

func SuggestHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	suggestions, err := elastic.SuggestProductData(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to perform suggest query: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}
