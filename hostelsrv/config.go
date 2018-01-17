package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"strings"
)

// Config defines configuration of chat server.
type Config struct {
	Port  uint
	Rooms []string
}

// Parse loads config from CLI and file where CLI args have priority.
func (c *Config) Parse() error {
	var cliPort uint
	var cliRooms string
	flag.UintVar(&cliPort, "port", 0, "Port to listen requests on")
	flag.StringVar(&cliRooms, "rooms", "", "List of rooms [room1|room2|..|roomN]")
	flag.Parse()

	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(f, &c); err != nil {
		return err
	}

	if cliPort != 0 {
		c.Port = cliPort
	}
	if cliRooms != "" {
		c.Rooms = c.Rooms[:0]
		for _, room := range strings.Split(cliRooms, "|") {
			c.Rooms = append(c.Rooms, room)
		}
	}
	return nil
}
