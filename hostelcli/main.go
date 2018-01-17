package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hostel-chat/hostelcli/chat"
	"io/ioutil"
	"net"
	"os"
)

func main() {
	svrAddr := flag.String("server", "127.0.0.1:5000", "Hostel server address [host[:port]]")
	flag.Parse()

	c := chat.Config{}
	f, err := ioutil.ReadFile("config.json")
	if err == nil {
		json.Unmarshal(f, &c)
	}

	fmt.Println("Connecting to:", *svrAddr, "...")
	conn, err := net.Dial("tcp", *svrAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected. You can now start chatting.")

	cl := chat.NewClient(conn, c)
	cl.Run(os.Stdin, os.Stdout)
}
