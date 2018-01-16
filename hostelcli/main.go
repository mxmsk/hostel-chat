package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	svrAddr := flag.String("server", "127.0.0.1:5000", "Hostel server address [host[:port]]")
	flag.Parse()

	fmt.Println(*svrAddr)
	conn, err := net.Dial("tcp", *svrAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			fmt.Fprintln(conn, s.Text())
		}
	}
}
