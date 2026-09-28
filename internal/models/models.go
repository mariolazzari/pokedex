package models

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type Config struct {
	commands       map[string]cliCommand
	locationLimit  int
	locationOffset int
}

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
