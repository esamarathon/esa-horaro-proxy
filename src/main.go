package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%d", h.Sum32())
}

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

func getHoraroApi(endpoint string) (*string, error) {
	response, found := memoryCache.Get(endpoint)
	if found {
		horaro, ok := response.(*string)

		if ok {
			return horaro, nil
		}
	}

	horaro, err := FetchHoraroApi(endpoint)
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

	eTag := `"` + hash(horaro.Schedule.Updated.UTC().Format(time.RFC3339Nano)) + `"`
	w.Header().Set("Etag", eTag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

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

	eTag := `"` + hash(horaro.Schedule.Updated.UTC().Format(time.RFC3339Nano)) + `"`
	w.Header().Set("Etag", eTag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

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

// Special use-case, does not transform the data, just proxies the api.
func apiProxy(w http.ResponseWriter, r *http.Request) {
	// Ignore Options request from CORS
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get endpoint parameter from URL
	parameter := mux.Vars(r)["endpoint"]
	endpoint, err := ParseHoraroUrl(parameter)
	if err != nil {
		log.Printf("Invalid horaro link '%s': %s", parameter, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid Horaro link: '%s'", err.Error()),
		})
		return
	}

	horaro, err := getHoraroApi(*endpoint)
	if err != nil {
		log.Printf("Could not find the horaro data from '%s': %s", *endpoint, err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Could not find the Horaro data",
		})
	}

	// cache for 5 minutes
	w.Header().Set("Cache-Control", "public, max-age=300")

	eTag := `"` + hash(*horaro) + `"`
	w.Header().Set("Etag", eTag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, eTag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, *horaro)
}

func customCorsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force the headers on the response, seemed to be missing during my testing
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		headers.Set("Access-Control-Allow-Headers", "*")

		h.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()
	router.SkipClean(true)
	router.HandleFunc("/{version:v[12]}/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/{version:v[12]}/esa/upcoming/{endpoint:.+}", upcomingPageHandler).Queries("amount", "{amount}").Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/{version:v[12]}/esa/schedule/{endpoint:.+}", schedulePageHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api_proxy/{endpoint:.+}", apiProxy).Methods(http.MethodGet, http.MethodOptions)

	handler := CaselessMatcher(router)
	handler = customCorsMiddleware(handler)

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
