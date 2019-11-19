package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HoraroResponse is the JSON format of the official Horaro API
type HoraroResponse struct {
	Meta struct {
		Exported time.Time `json:"exported"`
		Hint     string    `json:"hint"`
		API      string    `json:"api"`
		APILink  string    `json:"api-link"`
	} `json:"meta"`
	Schedule struct {
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Timezone    string    `json:"timezone"`
		Start       time.Time `json:"start"`
		StartT      int       `json:"start_t"`
		Website     string    `json:"website"`
		Twitter     string    `json:"twitter"`
		Twitch      string    `json:"twitch"`
		Description string    `json:"description"`
		Setup       string    `json:"setup"`
		SetupT      int       `json:"setup_t"`
		Updated     time.Time `json:"updated"`
		URL         string    `json:"url"`
		Event       struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"event"`
		HiddenColumns []string `json:"hidden_columns"`
		Columns       []string `json:"columns"`
		Items         []struct {
			Length     string      `json:"length"`
			LengthT    int         `json:"length_t"`
			Scheduled  time.Time   `json:"scheduled"`
			ScheduledT int         `json:"scheduled_t"`
			Data       []*string   `json:"data"`
			Options    interface{} `json:"options"`
		} `json:"items"`
	} `json:"schedule"`
}

var defaultTransport = &http.Transport{
	Dial:                (&net.Dialer{KeepAlive: 600 * time.Second}).Dial,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
}
var httpClient = &http.Client{
	Transport: defaultTransport,
	Timeout:   10 * time.Second,
}

// FetchHoraro fetches the full events from horaro
func FetchHoraro(year string) (*HoraroResponse, error) {
	url := fmt.Sprintf("https://horaro.org/esa/%s-one.json", year)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response HoraroResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
