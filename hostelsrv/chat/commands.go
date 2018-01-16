package chat

import (
	"fmt"
	"strings"
)

func publicMsg(nick string, room string, msg message) message {
	return message(fmt.Sprintf("%s@%s: %s", nick, room, msg))
}

// SubscribeCommand lets clients to subscribe to specific chat rooms.
type SubscribeCommand struct {
	hub *Hub
}

// NewSubscribeCommand creates a new instance of SubscribeCommand.
func NewSubscribeCommand(hub *Hub) *SubscribeCommand {
	return &SubscribeCommand{hub}
}

// Handle handles SubscribeCommand
func (cmd *SubscribeCommand) Handle(user identity, args string, outgoing chan<- message) {
	rnPairs := strings.Split(args, "|")
	if !cmd.validateRoomNickPairs(rnPairs, outgoing) {
		return
	}
	for _, pair := range rnPairs {
		rn := strings.SplitN(pair, ":", 2)
		room, nick := rn[0], rn[1]
		subscriber := subscriber{
			nick:     nick,
			outgoing: outgoing,
		}
		if err := cmd.hub.SubscribeToRoom(user, room, subscriber); err != nil {
			outgoing <- message(err.Error() + ".")
			continue
		}
		history := cmd.hub.getRoomHistory(room)
		for _, item := range history {
			outgoing <- publicMsg(item.nick, room, item.msg)
		}
	}
}

func (cmd *SubscribeCommand) validateRoomNickPairs(rnPairs []string, outgoing chan<- message) bool {
	for _, pair := range rnPairs {
		rn := strings.SplitN(pair, ":", 2)
		if rn[0] == "" {
			outgoing <- message("Room name is missing.")
			return false
		}
		if len(rn) == 1 || rn[1] == "" {
			outgoing <- message("Nickname for " + rn[0] + " is missing.")
			return false
		}
	}
	return true
}

// PublishCommand lets clients to publish message to rooms which
// they are subscribed to.
type PublishCommand struct {
	hub    *Hub
	msgCap int
}

// NewPublishCommand creates a new instance of PublishCommand.
func NewPublishCommand(hub *Hub, msgCap int) *PublishCommand {
	return &PublishCommand{
		hub:    hub,
		msgCap: msgCap,
	}
}

// Handle handles PublishCommand
func (cmd *PublishCommand) Handle(user identity, args string, outgoing chan<- message) {
	rm := strings.SplitN(args, "|", 2)
	if !cmd.validateRoomMsgPair(rm, outgoing) {
		return
	}
	target, msg := rm[0], message(rm[1])
	subs := cmd.hub.getSubscribers(target)
	if _, subscribed := subs[user]; !subscribed {
		outgoing <- message("You are not subscribed to " + target + ".")
		return
	}
	for id, sub := range subs {
		if id != user {
			sub.outgoing <- publicMsg(subs[user].nick, target, msg)
		}
	}
	cmd.hub.AppendRoomHistory(target, historyItem{
		nick: subs[user].nick,
		msg:  msg,
	})
}

func (cmd *PublishCommand) validateRoomMsgPair(rm []string, outgoing chan<- message) bool {
	if rm[0] == "" {
		outgoing <- message("Target room name is missing.")
		return false
	}
	if len(rm) == 1 || strings.TrimSpace(rm[1]) == "" {
		outgoing <- message("Message is empty.")
		return false
	}
	if len(rm[1]) > cmd.msgCap {
		outgoing <- message("Message is too long.")
		return false
	}
	return true
}
