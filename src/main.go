package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/cors"
)

var expiration = 15 * time.Minute
var cleanupInterval = 30 * time.Minute
var memoryCache = cache.New(expiration, cleanupInterval)

func eventsPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	// Get endpoint parameter from URL
	parameter := mux.Vars(r)["endpoint"]
	endpointPtr, err := FormatHoraroEndpoint(parameter)
	if err != nil {
		defer log.Printf("Invalid horaro link '%s': %s", parameter, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid Horaro link: %s", err.Error()),
		})
		return
	}
	endpoint := *endpointPtr

	// Return horaro response if still cached
	response, found := memoryCache.Get(endpoint)
	if found {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Fetching new data for '%s' from Horaro", endpoint)

	horaro, err := FetchHoraro(endpoint)
	if err != nil {
		defer log.Printf("Could not find the horaro data from '%s': %s", endpoint, err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Could not find the Horaro data",
		})
		return
	}

	// Transform Horaro response into a better format and save it in cache
	transformedHoraro := TransformHoraro(horaro)
	defer memoryCache.Set(endpoint, &transformedHoraro, cache.DefaultExpiration)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transformedHoraro)
}

func main() {
	router := mux.NewRouter()
	router.SkipClean(true)
	router.HandleFunc("/v1/esa/{endpoint:.+}", eventsPageHandler)

	handler := cors.Default().Handler(router)

	// Create address for HTTP server to listen on
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
