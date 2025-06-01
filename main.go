package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shahanMMiah/pokedexcli/internal"
)

const BASEURL string = "https://pokeapi.co/api/v2/"

type Config struct {
	Previous string
	Next     string
	Param    []string
	cache    *internal.Cache
	Pokedex  map[string]Pokemon
}

type CliCommands struct {
	name        string
	description string
	callback    func() error
}

type Pokemon struct {
	past_types               []interface{}
	species                  map[string]interface{}
	base_experience          float64
	held_items               []interface{}
	id                       float64
	stats                    []interface{}
	weight                   float64
	forms                    []interface{}
	is_default               bool
	sprites                  map[string]interface{}
	name                     string
	past_abilities           []interface{}
	types                    []interface{}
	abilities                []interface{}
	game_indices             []interface{}
	height                   float64
	order                    float64
	cries                    map[string]interface{}
	location_area_encounters string
	moves                    []interface{}
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
		c.Next = fmt.Sprintf("%s/%s/", BASEURL, "location-area")

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

	url := fmt.Sprintf("%s/%s", fmt.Sprintf("%s/%s/", BASEURL, "location-area"), c.Param[0])
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

func makePokemon(data map[string]interface{}) Pokemon {

	return Pokemon{
		past_types:               data["past_types"].([]interface{}),
		species:                  data["species"].(map[string]interface{}),
		base_experience:          data["base_experience"].(float64),
		held_items:               data["held_items"].([]interface{}),
		id:                       data["id"].(float64),
		stats:                    data["stats"].([]interface{}),
		weight:                   data["weight"].(float64),
		forms:                    data["forms"].([]interface{}),
		is_default:               data["is_default"].(bool),
		sprites:                  data["sprites"].(map[string]interface{}),
		name:                     data["name"].(string),
		past_abilities:           data["past_abilities"].([]interface{}),
		types:                    data["types"].([]interface{}),
		abilities:                data["abilities"].([]interface{}),
		game_indices:             data["game_indices"].([]interface{}),
		height:                   data["height"].(float64),
		order:                    data["order"].(float64),
		cries:                    data["cries"].(map[string]interface{}),
		location_area_encounters: data["location_area_encounters"].(string),
		moves:                    data["moves"].([]interface{}),
	}

}
func (c *Config) commandCatchPokemon() error {

	if len(c.Param) < 1 {
		return fmt.Errorf("no pokemon name specified")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", c.Param[0])

	url := fmt.Sprintf("%s/%s", fmt.Sprintf("%s/%s/", BASEURL, "pokemon"), c.Param[0])
	pokeData, lerr := getData(url, c.cache)

	if lerr != nil {
		return lerr
	}

	base_exp := pokeData["base_experience"]
	var min int64 = 50
	var max int64 = 310
	bigVal, randerr := rand.Int(rand.Reader, big.NewInt(max-min))

	if randerr != nil {
		return randerr
	}

	randVal := float64(bigVal.Int64() + min)
	if randVal >= base_exp.(float64) {

		fmt.Printf("%s was caught!\n", c.Param[0])

		c.Pokedex[c.Param[0]] = makePokemon(pokeData)
	} else {
		fmt.Printf("%s escaped\n", c.Param[0])
	}

	return nil

}

func (c *Config) commandInspectPokemon() error {

	if len(c.Param) < 1 {
		return fmt.Errorf("%v not pokemon specified", c.Param[0])

	}

	pokemenon, exists := c.Pokedex[c.Param[0]]
	if !exists {
		return fmt.Errorf("%v not found on pokedex", c.Param[0])
	}
	pokemonstats := pokemenon.stats

	stats := []any{pokemenon.name, pokemenon.height, pokemenon.weight}
	for _, statData := range pokemonstats {
		statVal := statData.(map[string]interface{})["base_stat"]
		stats = append(stats, statVal)

	}
	statsLn := fmt.Sprintf(
		"Name: %v\nHeight: %v\nWeight: %v\nStats:\n-hp: %v\n-attack: %v\n-deffense: %v\n-special-attack: %v\n-special-defense: %v\n-speed: %v\nTypes:\n",
		stats...,
	)

	for _, pType := range pokemenon.types {
		typ := pType.(map[string]interface{})["type"]
		statsLn += fmt.Sprintf("- %v\n", typ.(map[string]interface{})["name"])
	}

	fmt.Println(statsLn)
	return nil

}

func (c *Config) commandExit() error {

	fmt.Println("Closing the Pokedex... Goodbye!")

	os.Exit(0)
	return nil

}

func (c *Config) commandPokedex() error {

	prntLn := "Your Pokedex:\n"
	for name := range c.Pokedex {
		prntLn += fmt.Sprintf("- %v\n", name)
	}

	fmt.Print(prntLn)
	return nil
}

func getCommandMap(config *Config) map[string]CliCommands {

	return map[string]CliCommands{
		"exit":    {"exit", "Exit the Pokedex", config.commandExit},
		"help":    {"help", "Displays a help message", config.commandHelp},
		"map":     {"map", "Displays next 20 location areas in the Pokemon world", config.commandMapsFoward},
		"mapb":    {"map", "Displays last 20 location areas in the Pokemon world", config.commandMapsBack},
		"explore": {"explore", "Displays list of pokemon in specific location", config.commandExplore},
		"catch":   {"catch", "attemped to catch a specific pokemon", config.commandCatchPokemon},
		"inspect": {"inspect", "inspect data of catched pokemon", config.commandInspectPokemon},
		"pokedex": {"pokedex", "show names of pokemon in pokedex", config.commandPokedex},
	}

}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{Previous: "", Next: "", cache: internal.NewCache(100000 * time.Millisecond), Pokedex: make(map[string]Pokemon, 0)}

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
