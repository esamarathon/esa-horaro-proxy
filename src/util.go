package main

import "strings"

// IndexOfCaseInsensitive gets the index of an element in a list ignoring casing
func IndexOfCaseInsensitive(element string, data []string) int {
	for i, v := range data {
		if strings.EqualFold(element, v) {
			return i
		}
	}

	return -1
}
