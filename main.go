package main

import (
	"fmt"
	"strings"
)

func cleanInput(text string) []string {
	tokens := strings.Fields(text)

	for i, token := range tokens {
		tokens[i] = strings.ToLower(token)
	}

	return tokens
}

func main() {
	fmt.Println("Hello, World!")
}
