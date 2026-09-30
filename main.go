package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand/v2"
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
	location       string
	catch          string
	pokemons       map[string]models.Pokemon
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
	// endpoint
	url := fmt.Sprintf(
		"https://pokeapi.co/api/v2/location-area?limit=%d&offset=%d",
		cfg.locationLimit,
		cfg.locationOffset,
	)

	// api call
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// body parser
	var body models.LocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Results, nil
}

func getLocation(cfg *config) ([]models.PokemonEncounter, error) {
	url := fmt.Sprintf(
		"https://pokeapi.co/api/v2/location-area/%s",
		cfg.location,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body models.LocationArea
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.PokemonEncounters, nil
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

func exploreHelp(cfg *config) error {
	fmt.Println("Explore location")

	if cfg.locationOffset > 0 {
		cfg.locationOffset -= cfg.locationLimit
	}

	pokes, err := getLocation(cfg)
	if err != nil {
		return err
	}

	for _, poke := range pokes {
		fmt.Println(poke.Pokemon.Name)
	}

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

func catchHelp(cfg *config) error {
	if cfg.catch == "" {
		return fmt.Errorf("Please enter a Pokemon name")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", cfg.catch)

	url := fmt.Sprintf(
		"https://pokeapi.co/api/v2/pokemon/%s",
		cfg.catch,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("pokemon %q not found", cfg.catch)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PokeAPI returned status %s", resp.Status)
	}

	var body models.Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	chance := rand.IntN(1000)
	if chance > body.BaseExperience {
		cfg.pokemons[cfg.catch] = body
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

		// location name
		if len(tokens) == 2 && tokens[0] == "explore" {
			cfg.location = tokens[1]
		}

		// catch a pokemon
		if len(tokens) == 2 && tokens[0] == "catch" {
			cfg.catch = tokens[1]
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
		pokemons:       make(map[string]models.Pokemon),
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
		"explore": {
			name:        "explore",
			description: "explore location",
			callback:    exploreHelp,
		},
		"catch": {
			name:        "catch",
			description: "catch pokemon",
			callback:    catchHelp,
		},
	}

	startRepl(cfg)
}
