package handlers

import (
	"fmt"
	"net/http"
)

var searchStats = 0

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	searchStats++
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Total searches: %d\n", searchStats)
}
