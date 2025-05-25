package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/shahanMMiah/pokedexcli/internal"
)

type Config struct {
	Previous string
	Next     string
	cache    *internal.Cache
}

type CliCommands struct {
	name        string
	description string
	callback    func() error
}

func cleanInput(text string) []string {
	sliced := strings.Fields(text)

	for num := range sliced {
		sliced[num] = strings.ToLower(sliced[num])
	}

	return sliced
}
func (c *Config) commandHelp() error {
	fmt.Println("Welcome to the Pokedex!\nusage:\n")

	for name, command := range getCommandMap(c) {
		fmt.Printf("%s: %s\n", name, command.description)
	}
	return nil
}

func getMapData(url string, cache *internal.Cache) (map[string]interface{}, error) {

	data, exists := cache.Get(url)
	if !exists {
		fmt.Printf("Getting request for %s\n", url)
		req, rerr := http.NewRequest("GET", url, nil)

		if rerr != nil {
			return nil, rerr
		}

		res, reserr := http.DefaultClient.Do(req)
		if reserr != nil {
			return nil, reserr
		}

		defer res.Body.Close()
		body, berr := io.ReadAll(res.Body)
		if berr != nil {
			return nil, berr
		}
		data = body
		cache.Add(url, data)

	} else {
		fmt.Printf("Found %s in cache\n", url)
	}

	var mapData map[string]interface{}

	derr := json.Unmarshal(data, &mapData)

	if derr != nil {
		return nil, derr
	}

	return mapData, nil

}
func (c *Config) commandMapsBack() error {

	if c.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	mapData, merr := getMapData(c.Previous, c.cache)
	if merr != nil {
		return merr
	}

	c.Next = c.Previous
	if mapData["previous"] == nil {
		c.Previous = ""
	} else {
		c.Previous = mapData["previous"].(string)
	}

	for _, result := range mapData["results"].([]interface{}) {
		fmt.Println(result.(map[string]interface{})["name"])
	}

	return nil

}

func (c *Config) commandMapsFoward() error {
	if c.Next == "" {
		c.Next = "https://pokeapi.co/api/v2/location-area/"

	}

	mapData, merr := getMapData(c.Next, c.cache)
	if merr != nil {
		return merr
	}

	if mapData["next"] == nil {
		fmt.Println("you're on the last page")
		return nil

	}
	c.Next = mapData["next"].(string)

	if mapData["previous"] != nil {
		c.Previous = mapData["previous"].(string)
	} else {
		c.Previous = ""
	}

	for _, result := range mapData["results"].([]interface{}) {
		fmt.Println(result.(map[string]interface{})["name"])
	}

	return nil
}

func (c *Config) commandExit() error {

	fmt.Println("Closing the Pokedex... Goodbye!")

	os.Exit(0)
	return nil

}

func getCommandMap(config *Config) map[string]CliCommands {

	return map[string]CliCommands{
		"exit": CliCommands{"exit", "Exit the Pokedex", config.commandExit},
		"help": CliCommands{"help", "Displays a help message", config.commandHelp},
		"map":  CliCommands{"map", "Displays next 20 location areas in the Pokemon world", config.commandMapsFoward},
		"mapb": CliCommands{"map", "Displays next 20 location areas in the Pokemon world", config.commandMapsBack},
	}

}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{Previous: "", Next: "", cache: internal.NewCache(5000000000)}

	for {

		fmt.Print("Pokedex > ")

		scanner.Scan()
		firstWord := cleanInput(scanner.Text())[0]

		fmt.Printf("Your command was: %s \n", firstWord)

		command, ok := getCommandMap(&config)[firstWord]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		command.callback()

	}

}
