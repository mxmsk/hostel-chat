package main

import (
	"flag"
	"hostel-chat/hostelsrv/chat"
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
	log.Println("Listening on port", *port)

	chatSvc := initChatService()
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

func initChatService() *chat.Service {
	hub := chat.NewHub(128)
	commands := map[string]chat.Command{
		"subscribe": chat.NewSubscribeCommand(hub),
		"publish":   chat.NewPublishCommand(hub, 254),
	}
	hub.CreateRoom("A")
	hub.CreateRoom("B")
	hub.CreateRoom("C")
	return chat.NewService(commands, hub)
}
