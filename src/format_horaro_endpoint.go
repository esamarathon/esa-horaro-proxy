package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// FormatHoraroEndpoint validates the users supplied endpoint and throws an error if it doesn't exist
func FormatHoraroEndpoint(parameter string) (*string, error) {
	if strings.HasSuffix(parameter, ".json") == false {
		parameter += ".json"
	}

	// If parameter isn't a URL
	if regexp.MustCompile("^[^\\/]+$").MatchString(parameter) {
		horaroURL := fmt.Sprintf("https://horaro.org/esa/%s", parameter)
		return &horaroURL, nil
	}

	endpoint, err := url.Parse(parameter)
	if err != nil {
		return nil, errors.New("Can not parse URL")
	}

	if endpoint.Hostname() != "horaro.org" {
		return nil, errors.New("Can not fetch from different domain than Horaro")
	}

	if endpoint.Scheme != "https" {
		return nil, errors.New("Can only fetch from HTTPS")
	}

	return &parameter, nil
}
