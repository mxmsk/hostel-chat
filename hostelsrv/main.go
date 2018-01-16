package main

import (
	"flag"
	"hostel/hostelsrv/chat"
	"log"
	"net"
)

var chatCommands = map[string]chat.Command{
	"subscribe": &chat.SubscribeCommand{},
	"publish":   &chat.PublishCommand{},
}

func main() {
	port := flag.String("port", "5000", "Port to listen requests on")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalln("Can't start server:", err)
	}

	chatSvc := initChatService()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go chatSvc.HandleClient(conn)
	}
}

func initChatService() *chat.Service {
	hub := chat.NewHub(128)
	commands := map[string]chat.Command{
		"subscribe": chat.NewSubscribeCommand(hub),
		"publish":   chat.NewPublishCommand(hub, 254),
	}
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	return chat.NewService(commands, hub)
}
