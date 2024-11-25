package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snykk/simple-go-elasticsearch/elastic"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := elastic.GetIndexStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
