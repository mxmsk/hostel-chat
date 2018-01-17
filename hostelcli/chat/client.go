package chat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Server defines requirements for a chat server.
type Server interface {
	io.ReadWriter
}

// Config defines configuration of client.
type Config struct {
	Subscriptions []string
}

// Client allows to interact with chat server.
type Client struct {
	s Server
	c Config
}

// NewClient returns a new instance of Client for interaction
// with the specified server connection.
func NewClient(s Server, c Config) *Client {
	return &Client{
		s: s,
		c: c,
	}
}

// Run starts chat loop, allowing to interact with the server
// using the specified terminals streams.
func (cl *Client) Run(in io.Reader, out io.Writer) {
	go func() {
		s := bufio.NewScanner(cl.s)
		for s.Scan() {
			fmt.Fprintln(out, s.Text())
		}
	}()

	subscribe := bytes.NewBufferString("subscribe")
	for _, subscription := range cl.c.Subscriptions {
		subscribe.WriteString("|" + subscription)
	}
	fmt.Fprintln(cl.s, subscribe)

	for {
		s := bufio.NewScanner(in)
		for s.Scan() {
			//fmt.Fprintln(cl.s, s.Text())
			fmt.Fprintln(cl.s, "publish|A|"+s.Text())
		}
	}
}
