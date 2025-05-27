package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shahanMMiah/pokedexcli/internal"
)

const LOCATIONURL string = "https://pokeapi.co/api/v2/location-area/"

type Config struct {
	Previous string
	Next     string
	Param    []string
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
	fmt.Println("Welcome to the Pokedex!\nusage:")

	for name, command := range getCommandMap(c) {
		fmt.Printf("%s: %s\n", name, command.description)
	}
	return nil
}

func addCacheRequest(url string, cache *internal.Cache) ([]byte, error) {
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
	cache.Add(url, body)

	return body, nil

}

func getData(url string, cache *internal.Cache) (map[string]interface{}, error) {

	data, exists := cache.Get(url)
	if !exists {
		body, derr := addCacheRequest(url, cache)
		if derr != nil {
			return nil, derr
		}

		data = body

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

	mapData, merr := getData(c.Previous, c.cache)
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
		c.Next = LOCATIONURL

	}

	mapData, merr := getData(c.Next, c.cache)
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

func (c *Config) commandExplore() error {

	fmt.Println("Exploring Pokemon...")
	if len(c.Param) < 1 {
		return fmt.Errorf("no location name specified")
	}

	url := fmt.Sprintf("%s/%s", LOCATIONURL, c.Param[0])
	locData, lerr := getData(url, c.cache)
	if lerr != nil {
		return lerr
	}

	encounters := locData["pokemon_encounters"].([]interface{})

	fmt.Println("Found Pokemon:")
	for ind := range encounters {
		pokemonData := encounters[ind].(map[string]interface{})["pokemon"]

		fmt.Printf(" - %s\n", pokemonData.(map[string]interface{})["name"])

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
		"exit":    {"exit", "Exit the Pokedex", config.commandExit},
		"help":    {"help", "Displays a help message", config.commandHelp},
		"map":     {"map", "Displays next 20 location areas in the Pokemon world", config.commandMapsFoward},
		"mapb":    {"map", "Displays last 20 location areas in the Pokemon world", config.commandMapsBack},
		"explore": {"explore", "Displays list of pokemon in specific location", config.commandExplore},
	}

}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{Previous: "", Next: "", cache: internal.NewCache(10000 * time.Millisecond)}

	for {

		fmt.Print("Pokedex > ")

		scanner.Scan()
		words := cleanInput(scanner.Text())

		if len(words) < 1 {
			continue
		}

		firstWord := words[0]

		if len(words) > 1 {
			config.Param = words[1:]
		}

		fmt.Printf("Your command was: %s \n", firstWord)

		command, ok := getCommandMap(&config)[firstWord]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		callErr := command.callback()
		if callErr != nil {
			fmt.Println(callErr)
		}

		config.Param = make([]string, 0)

	}

}
