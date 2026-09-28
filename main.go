package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mariolazzari/pokedex/internal/models"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	commands       map[string]cliCommand
	locationLimit  int
	locationOffset int
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)

	return nil
}

func getLocations(cfg *config) ([]models.Location, error) {
	url := fmt.Sprintf(
		"https://pokeapi.co/api/v2/location-area?limit=%d&offset=%d",
		cfg.locationLimit,
		cfg.locationOffset,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body LocationResponse

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Results, nil
}

func mapHelp(cfg *config) error {
	fmt.Println("Show next 20 locations")

	locations, err := getLocations(cfg)
	if err != nil {
		return err
	}

	for _, location := range locations {
		fmt.Println(location.Name)
	}

	cfg.locationOffset += cfg.locationLimit

	return nil
}

func mapbHelp(cfg *config) error {
	fmt.Println("Show previous 20 locations")

	if cfg.locationOffset > 0 {
		cfg.locationOffset -= cfg.locationLimit
	}

	locations, err := getLocations(cfg)
	if err != nil {
		return err
	}

	for _, location := range locations {
		fmt.Println(location.Name)
	}

	return nil
}

func cleanInput(text string) []string {
	tokens := strings.Fields(text)

	for i, token := range tokens {
		tokens[i] = strings.ToLower(token)
	}

	return tokens
}

func startRepl(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}

		tokens := cleanInput(scanner.Text())

		if len(tokens) == 0 {
			continue
		}

		cmd, ok := cfg.commands[tokens[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

func main() {
	cfg := &config{
		locationLimit:  20,
		locationOffset: 0,
	}

	cfg.commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Pokedex help",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "display the next 20 locations",
			callback:    mapHelp,
		},
		"mapb": {
			name:        "mapb",
			description: "display the previous 20 locations",
			callback:    mapbHelp,
		},
	}

	startRepl(cfg)
}
