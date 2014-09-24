package main

import (
	"fmt"
	"strings"
)

func TrimInput(input string) string {
	return strings.TrimRight(input, "\r\n")
}

func Confirm() bool {
	var answer string
	_, err := fmt.Scanln(&answer)
	if err != nil {
		return false
	}

	yeses := []string{"y", "yes"}

	for _, yes := range yeses {
		if strings.ToLower(answer) == yes {
			return true
		}
	}

	return false
}
