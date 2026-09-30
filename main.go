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
	inspect        string
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

func getPokemon(cfg *config) (models.Pokemon, error) {

	url := fmt.Sprintf(
		"https://pokeapi.co/api/v2/pokemon/%s",
		cfg.catch,
	)

	resp, err := http.Get(url)
	if err != nil {
		return models.Pokemon{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.Pokemon{}, fmt.Errorf("pokemon %q not found", cfg.catch)
	}

	if resp.StatusCode != http.StatusOK {
		return models.Pokemon{}, fmt.Errorf("PokeAPI returned status %s", resp.Status)
	}

	var body models.Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return models.Pokemon{}, err
	}

	return body, nil
}

func catchHelp(cfg *config) error {
	if cfg.catch == "" {
		return fmt.Errorf("Please enter a Pokemon name")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", cfg.catch)

	body, err := getPokemon(cfg)
	if err != nil {
		return err
	}

	chance := rand.IntN(1000)
	if chance > body.BaseExperience {
		cfg.pokemons[cfg.catch] = body
		fmt.Printf("%s was caught!\n", cfg.catch)
	} else {
		fmt.Printf("%s was not caught!\n", cfg.catch)
	}

	return nil
}

func inspectHelp(cfg *config) error {
	if cfg.inspect == "" {
		return fmt.Errorf("please enter a Pokemon name")
	}

	pokemon, ok := cfg.pokemons[cfg.inspect]
	if !ok {
		return fmt.Errorf("%s has not been caught yet", cfg.inspect)
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokemonType.Type.Name)
	}

	return nil
}

func pokedexHelp(cfg *config) error {

	if len(cfg.pokemons) == 0 {
		return fmt.Errorf("No pokemon caught")
	}

	fmt.Println("Your Pokedex:")

	for p := range cfg.pokemons {
		fmt.Printf("- %s\n", p)
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

		// inspect a pokemon
		if len(tokens) == 2 && tokens[0] == "inspect" {
			cfg.inspect = tokens[1]
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
		}, "inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon",
			callback:    inspectHelp,
		}, "pokedex": {
			name:        "pokedex",
			description: "Caught Pokemons",
			callback:    pokedexHelp,
		},
	}

	startRepl(cfg)
}
