package main

import (
	"strconv"
)

func parseInt(s string) (uint32, error) {

	num, err := strconv.Atoi(s)

	return uint32(num), err
}

func mapStrings(slice []string, fn func(string) string) []string {
	var newSlice []string
	for _, item := range slice {
		newSlice = append(newSlice, fn(item))
	}
	return newSlice
}