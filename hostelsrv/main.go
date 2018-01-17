package main

import (
	"fmt"
	"hostel-chat/hostelsrv/chat"
	"log"
	"net"
)

var chatCommands = map[string]chat.Command{
	"subscribe": &chat.SubscribeCommand{},
	"publish":   &chat.PublishCommand{},
}

func main() {
	c := Config{}
	if err := c.Parse(); err != nil {
		log.Fatalln("Config error:", err)
	}
	log.Println("Use config:", c)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		log.Fatalln("Can't start server:", err)
	}
	log.Println("Listening on port", c.Port)

	chatSvc := initChatService(c)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		log.Println("New connection", conn.RemoteAddr())
		go chatSvc.HandleClient(conn)
	}
}

func initChatService(c Config) *chat.Service {
	hub := chat.NewHub(128)
	commands := map[string]chat.Command{
		"subscribe": chat.NewSubscribeCommand(hub),
		"publish":   chat.NewPublishCommand(hub, 254),
	}
	for _, room := range c.Rooms {
		hub.CreateRoom(room)
	}
	return chat.NewService(commands, hub)
}
