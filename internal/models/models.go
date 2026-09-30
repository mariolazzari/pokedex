package models

type Location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []Location `json:"results"`
}

type LocationArea struct {
	Name              string             `json:"name"`
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEncounter struct {
	Pokemon Pokemon `json:"pokemon"`
}

type Pokemon struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	BaseExperience int    `josn:"base_experience"`
}
