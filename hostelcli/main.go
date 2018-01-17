package main

import (
	"fmt"
	"hostel-chat/hostelcli/chat"
	"log"
	"net"
	"os"
)

func main() {
	c := Config{}
	if err := c.Parse(); err != nil {
		log.Fatalln("Config error:", err)
	}
	log.Println("Use config:", c)

	log.Println("Connecting to:", c.Server, "...")
	conn, err := net.Dial("tcp", c.Server)
	if err != nil {
		log.Fatalln(err)
	}

	cl := chat.NewClient(conn)
	for _, sub := range c.Subscriptions {
		room, nick := roomNickPair(sub)
		cl.AddSubscription(room, nick)
	}

	fmt.Println("Connected! You can now start chatting.")
	cl.Run(os.Stdin, os.Stdout)
}
