package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

var c = cache.New(15*time.Minute, 30*time.Minute)

func eventsPageHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	// Get year parameter from URL
	year := mux.Vars(r)["year"]

	// Return horaro response if still cached
	response, found := c.Get(year)
	if found {
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Fetching new data for '%s' from Horaro", year)

	horaro, err := FetchHoraro(year)
	if err != nil {
		log.Printf("Failed fetching horaro: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed fetching Horaro data",
		})
		return
	}

	// Transform Horaro response into a better format
	transformedHoraro := TransformHoraro(horaro)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transformedHoraro)

	// Store current horaro response in cache
	c.Set(year, transformedHoraro, cache.DefaultExpiration)
}

var port = 8080

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/esa/{year}", eventsPageHandler).Methods(http.MethodGet, http.MethodOptions)

	// Create address for HTTP server to listen on
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
