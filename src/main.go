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
var cleanupInterval = 60 * time.Minute
var memoryCache = cache.New(expiration, cleanupInterval)

func getHoraroV1(endpoint string) (*TransformedHoraroResponseV1, *time.Duration, error) {
	// Return horaro response if still cached
	response, expiresAt, found := memoryCache.GetWithExpiration(endpoint)
	expiry := expiresAt.Sub(time.Unix(0, 0))
	castedResponse, ok := response.(*TransformedHoraroResponseV1)
	if found && ok {
		return castedResponse, &expiry, nil
	}

	log.Printf("F %d", (int)(expiry.Seconds()))
	log.Printf("Fetching new data for '%s' from Horaro", endpoint)

	horaro, err := FetchHoraro(endpoint)
	if err != nil {
		return nil, nil, err
	}

	// Transform Horaro response into a better format and save it in cache
	transformedHoraro := TransformHoraroV1(horaro)
	defer memoryCache.Set(endpoint, &transformedHoraro, cache.DefaultExpiration)

	return &transformedHoraro, &expiration, nil
}

func getHoraroV2(endpoint string) (*TransformedHoraroResponseV2, *time.Duration, error) {
	// Return horaro response if still cached
	response, expiresAt, found := memoryCache.GetWithExpiration(endpoint)
	expiry := expiresAt.Sub(time.Unix(0, 0))
	castedResponse, ok := response.(*TransformedHoraroResponseV2)
	if found && ok {
		return castedResponse, &expiry, nil
	}

	log.Printf("F %d", (int)(expiry.Seconds()))
	log.Printf("Fetching new data for '%s' from Horaro", endpoint)

	horaro, err := FetchHoraro(endpoint)
	if err != nil {
		return nil, nil, err
	}

	// Transform Horaro response into a better format and save it in cache
	transformedHoraro := TransformHoraroV2(horaro)
	defer memoryCache.Set(endpoint, &transformedHoraro, cache.DefaultExpiration)

	return &transformedHoraro, &expiration, nil
}

func upcomingPageHandlerV1(w http.ResponseWriter, r *http.Request) {
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

	horaro, expiry, err := getHoraroV1(*endpoint)
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

	upcoming := UpcomingHoraroV1(*horaro, amount)

	w.WriteHeader(http.StatusOK)
	if expiry != nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", (int)(expiry.Seconds())))
	}
	json.NewEncoder(w).Encode(upcoming)
}

func upcomingPageHandlerV2(w http.ResponseWriter, r *http.Request) {
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

	horaro, expiry, err := getHoraroV2(*endpoint)
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

	upcoming := UpcomingHoraroV2(*horaro, amount)

	w.WriteHeader(http.StatusOK)
	if expiry != nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", (int)(expiry.Seconds())))
	}
	json.NewEncoder(w).Encode(upcoming)
}

func schedulePageHandlerV1(w http.ResponseWriter, r *http.Request) {
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

	horaro, expiry, err := getHoraroV1(*endpoint)
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
	if expiry != nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", (int)(expiry.Seconds())))
	}
	json.NewEncoder(w).Encode(schedule)
}

func schedulePageHandlerV2(w http.ResponseWriter, r *http.Request) {
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

	horaro, expiry, err := getHoraroV2(*endpoint)
	if err != nil {
		log.Printf("Could not find the horaro data from '%s': %s", *endpoint, err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Could not find the Horaro data",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	if expiry != nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", (int)(expiry.Seconds())))
	}
	json.NewEncoder(w).Encode(horaro)
}

func main() {
	router := mux.NewRouter()
	router.SkipClean(true)
	router.HandleFunc("/v1/esa/upcoming/{endpoint:.+}", upcomingPageHandlerV1).Queries("amount", "{amount}").Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v1/esa/upcoming/{endpoint:.+}", upcomingPageHandlerV1).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v1/esa/schedule/{endpoint:.+}", schedulePageHandlerV1).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v2/esa/upcoming/{endpoint:.+}", upcomingPageHandlerV2).Queries("amount", "{amount}").Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v2/esa/upcoming/{endpoint:.+}", upcomingPageHandlerV2).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/v2/esa/schedule/{endpoint:.+}", schedulePageHandlerV2).Methods(http.MethodGet, http.MethodOptions)

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
