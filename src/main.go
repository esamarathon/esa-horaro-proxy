package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/cors"
)

var expiration = 15 * time.Minute
var cleanupInterval = 30 * time.Minute
var memoryCache = cache.New(expiration, cleanupInterval)

func getHoraro(endpoint string) (*TransformedHoraroResponse, error) {
	// Return horaro response if still cached
	response, found := memoryCache.Get(endpoint)
	castedResponse, ok := response.(*TransformedHoraroResponse)
	if found && ok {
		return castedResponse, nil
	}

	log.Printf("Fetching new data for '%s' from Horaro", endpoint)

	horaro, err := FetchHoraro(endpoint)
	if err != nil {
		return nil, err
	}

	// Transform Horaro response into a better format and save it in cache
	transformedHoraro := TransformHoraro(horaro)
	defer memoryCache.Set(endpoint, &transformedHoraro, cache.DefaultExpiration)

	return &transformedHoraro, nil
}

func upcomingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	// Get endpoint parameter from URL
	parameter := mux.Vars(r)["endpoint"]
	endpoint, err := FormatHoraroEndpoint(parameter)
	if err != nil {
		log.Printf("Invalid horaro link '%s': %s", parameter, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid Horaro link: '%s'", err.Error()),
		})
		return
	}

	horaro, err := getHoraro(*endpoint)
	if err != nil {
		log.Printf("Could not find the horaro data from '%s': %s", *endpoint, err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Could not find the Horaro data",
		})
		return
	}

	amountStr := r.FormValue("amount")
	amount, err := strconv.Atoi(amountStr)
	if amountStr == "" || err == nil {
		amount = 5
	}

	upcoming := UpcomingHoraro(*horaro, amount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(upcoming)
}

func schedulePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	// Get endpoint parameter from URL
	parameter := mux.Vars(r)["endpoint"]
	endpoint, err := FormatHoraroEndpoint(parameter)
	if err != nil {
		log.Printf("Invalid horaro link '%s': %s", parameter, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid Horaro link: '%s'", err.Error()),
		})
		return
	}

	horaro, err := getHoraro(*endpoint)
	if err != nil {
		log.Printf("Could not find the horaro data from '%s': %s", *endpoint, err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Could not find the Horaro data",
		})
		return
	}

	schedule := OrganizeHoraro(*horaro)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

func main() {
	router := mux.NewRouter()
	router.SkipClean(true)
	router.HandleFunc("/v1/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Queries("amount", "{amount}").Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v1/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v1/esa/schedule/{endpoint:.+}", schedulePageHandler).Methods(http.MethodGet, http.MethodOptions)

	handler := CaselessMatcher(router)
	handler = cors.Default().Handler(handler)

	// Create address for HTTP server to listen on
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	server := &http.Server{
		Handler:      handler,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
