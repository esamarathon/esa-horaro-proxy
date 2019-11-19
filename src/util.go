package main

import "strings"

// IndexOf gets the index of an element in a list
func IndexOf(element string, data []string, caseSensitive bool) int {
	for i, v := range data {
		if element == v {
			return i
		}
	}

	return -1
}

// IndexOfCaseInsensitive gets the index of an element in a list ignoring casing
func IndexOfCaseInsensitive(element string, data []string) int {
	for i, v := range data {
		if strings.EqualFold(element, v) {
			return i
		}
	}

	return -1
}
