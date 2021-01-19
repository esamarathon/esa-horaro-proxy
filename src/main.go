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

var expiration = 10 * time.Minute
var cleanupInterval = 60 * time.Minute
var memoryCache = cache.New(expiration, cleanupInterval)

func getHoraro(endpoint string) (*HoraroResponse, error) {
	response, found := memoryCache.Get(endpoint)
	if found {
		horaro, ok := response.(*HoraroResponse)

		if ok {
			return horaro, nil
		}
	}

	log.Printf("Fetching new data for '%s' from Horaro", endpoint)

	horaro, err := FetchHoraro(endpoint)
	if err != nil {
		return nil, err
	}

	defer memoryCache.Set(endpoint, horaro, cache.DefaultExpiration)

	return horaro, nil
}

func upcomingPageHandler(w http.ResponseWriter, r *http.Request) {
	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get endpoint parameter from URL
	parameter := mux.Vars(r)["endpoint"]
	endpoint, err := FormatHoraroEndpoint(parameter)
	if err != nil {
		log.Printf("Invalid horaro link '%s': %s", parameter, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid Horaro link",
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

	w.Header().Set("Cache-Control", "max-age=600")

	version := mux.Vars(r)["version"]
	if version == "v1" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UpcomingHoraroV1(TransformHoraroV1(horaro), amount))
	} else if version == "v2" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UpcomingHoraroV2(TransformHoraroV2(horaro), amount))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func schedulePageHandler(w http.ResponseWriter, r *http.Request) {
	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")

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

	w.Header().Set("Cache-Control", "max-age=360")

	version := mux.Vars(r)["version"]
	if version == "v1" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OrganizeHoraro(TransformHoraroV1(horaro)))
	} else if version == "v2" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TransformHoraroV2(horaro))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	router := mux.NewRouter()
	router.SkipClean(true)
	router.HandleFunc("/{version:v[12]}/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/{version:v[12]}/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Queries("amount", "{amount}").Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/{version:v[12]}/esa/schedule/{endpoint:.+}", schedulePageHandler).Methods(http.MethodGet, http.MethodOptions)

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
