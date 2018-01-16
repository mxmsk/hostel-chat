package chat

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

// Command provides interface for an action accepted by chat service.
type Command interface {
	Handle(user identity, args string, outgoing chan<- message)
}

// Unsubscriber lets signal that user should be unsubscribed from chat.
type Unsubscriber interface {
	Unsubscribe(user identity)
}

// Client defines requirements for a chat client.
type Client interface {
	io.ReadWriteCloser
}

type message string
type identity string

// Service encapsulates features of chat server.
type Service struct {
	unsubscriber Unsubscriber
	commands     map[string]Command
}

// NewService creates new instance of chat service with
// the list of supported commands.
func NewService(commands map[string]Command, unsubscriber Unsubscriber) *Service {
	return &Service{
		unsubscriber: unsubscriber,
		commands:     commands,
	}
}

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// HandleClient starts handling chat commands from the specified client.
func (s *Service) HandleClient(cl Client) {
	defer cl.Close()

	user := identity(randToken())
	disconnect := make(chan struct{}, 1)
	defer func() {
		close(disconnect)
		s.unsubscriber.Unsubscribe(user)
	}()

	incoming := make(chan message)
	outgoing := make(chan message)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer close(outgoing)
		s.handleIncoming(user, incoming, outgoing, disconnect)
	}()
	go func() {
		defer wg.Done()
		s.handleOutgoing(cl, outgoing)
	}()

	scanner := bufio.NewScanner(cl)
	for scanner.Scan() {
		incoming <- message(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Println("Error reading input:", err)
	}
	disconnect <- struct{}{}
	wg.Wait()
}

func (s *Service) handleIncoming(user identity, incoming <-chan message,
	outgoing chan<- message, disconnect <-chan struct{}) {

	for {
		select {
		case <-disconnect:
			return
		case m := <-incoming:
			na := strings.SplitN(string(m), "|", 2)
			name, args := na[0], ""
			if len(na) > 1 {
				args = na[1]
			}
			if cmd, ok := s.commands[name]; ok {
				func() {
					defer func() {
						if r := recover(); r != nil {
							outgoing <- "Unexpected server error!"
							log.Println("Command", name, "paniced:", r)
						}
					}()
					cmd.Handle(user, args, outgoing)
				}()
			} else {
				outgoing <- message("Unknown command: " + name + ".")
			}
		}
	}
}

func (s *Service) handleOutgoing(w io.Writer, outgoing <-chan message) {
	for m := range outgoing {
		fmt.Fprintln(w, m)
	}
}
