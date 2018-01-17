package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"strings"
)

// Config defines configuration of chat client.
type Config struct {
	Server        string
	Subscriptions []string
}

// Parse loads config from CLI and file where CLI args have priority.
func (c *Config) Parse() error {
	var cliServer string
	var cliSubs string
	flag.StringVar(&cliServer, "server", "", "Hostel server address [host[:port]]")
	flag.StringVar(&cliSubs, "subs", "", "Room subscriptions [room:nick1|room2:nick1...]")
	flag.Parse()

	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(f, &c); err != nil {
		return err
	}

	if cliServer != "" {
		c.Server = cliServer
	}
	if cliSubs != "" {
		c.Subscriptions = c.Subscriptions[:0]
		for _, sub := range strings.Split(cliSubs, "|") {
			c.Subscriptions = append(c.Subscriptions, sub)
		}
	}
	return nil
}

func roomNickPair(subscription string) (string, string) {
	pair := strings.SplitN(subscription, ":", 2)
	if len(pair) == 1 {
		return pair[0], ""
	}
	return pair[0], pair[1]
}
