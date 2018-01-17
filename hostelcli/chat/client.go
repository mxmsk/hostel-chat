package chat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Server defines requirements for chat server.
type Server interface {
	io.ReadWriter
}

// Client allows to interact with chat server.
type Client struct {
	srv           Server
	subscriptions map[string]string
}

// NewClient returns a new instance of Client for interaction
// with the specified server connection.
func NewClient(srv Server) *Client {
	return &Client{
		srv:           srv,
		subscriptions: make(map[string]string),
	}
}

// AddSubscription instructs Client to subscribe to the specified room.
func (cl *Client) AddSubscription(room string, nick string) {
	cl.subscriptions[room] = nick
}

// Run starts chat loop, allowing to interact with the server
// using the specified terminals streams.
func (cl *Client) Run(in io.Reader, out io.Writer) {
	go func() {
		s := bufio.NewScanner(cl.srv)
		for s.Scan() {
			fmt.Fprintln(out, s.Text())
		}
	}()

	defaultRoom := cl.subscribe()
	for {
		s := bufio.NewScanner(in)
		for s.Scan() {
			ln := s.Text()
			if len(ln) == 0 {
				continue
			}
			room := defaultRoom
			if ln[0] == '/' {
				i := strings.Index(ln, " ")
				if i > 0 {
					room = ln[1:i]
					ln = ln[i+1:]
				}
			}
			fmt.Fprintf(cl.srv, "publish|%s|%s\n", room, ln)
		}
	}
}

func (cl *Client) subscribe() string {
	var defaultRoom string
	subscribe := bytes.NewBufferString("subscribe")
	for room, nick := range cl.subscriptions {
		subscribe.WriteString(fmt.Sprintf("|%s:%s", room, nick))
		defaultRoom = room
	}

	fmt.Fprintln(cl.srv, subscribe)
	return defaultRoom
}
