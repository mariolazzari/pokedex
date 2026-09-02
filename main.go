package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type config struct {
	commands map[string]cliCommand
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
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
	cfg := &config{}

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
	}

	startRepl(cfg)
}
