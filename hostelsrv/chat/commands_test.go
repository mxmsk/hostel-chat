package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeCommand_CorrectArgs_UserSubscribed(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	outgoing := make(chan<- message)
	expectedSub1 := subscriber{
		nick:     "nick1",
		outgoing: outgoing,
	}
	expectedSub2 := subscriber{
		nick:     "nick2",
		outgoing: outgoing,
	}

	cmd := NewSubscribeCommand(hub)
	cmd.Handle("id1", "room1:nick1|room2:nick2", outgoing)

	assert.Equal(t, expectedSub1, hub.getSubscribers("room1")["id1"])
	assert.Equal(t, expectedSub2, hub.getSubscribers("room2")["id1"])
}

func TestSubscribeCommand_CorrectArgs_HistoryToOutgoing(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	hub.CreateRoom("room3")
	hub.AppendRoomHistory("room1", historyItem{nick: "nick8", msg: "msg1"})
	hub.AppendRoomHistory("room2", historyItem{nick: "nick10", msg: "msg2"})
	hub.AppendRoomHistory("room2", historyItem{nick: "nick8", msg: "msg3"})
	hub.AppendRoomHistory("room3", historyItem{nick: "nick10", msg: "msg4"})
	outgoing := make(chan message, 3)

	cmd := NewSubscribeCommand(hub)
	cmd.Handle("id1", "room1:nick1|room2:nick2", outgoing)

	assert.Equal(t, message("nick8@room1: msg1"), <-outgoing)
	assert.Equal(t, message("nick10@room2: msg2"), <-outgoing)
	assert.Equal(t, message("nick8@room2: msg3"), <-outgoing)
}

func TestSubscribeCommand_HasUnknownRooms_UnknownToOutgoing(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room3")
	outgoing := make(chan message, 2)

	cmd := NewSubscribeCommand(hub)
	cmd.Handle("id1", "room1:nick1|room2:nick1|room3:nick1|room4:nick1", outgoing)

	assert.Contains(t, hub.getSubscribers("room1"), identity("id1"))
	assert.Contains(t, hub.getSubscribers("room3"), identity("id1"))

	assert.Len(t, outgoing, 2)
	assert.Contains(t, <-outgoing, "Cannot subscribe to unknown room: room2.")
	assert.Contains(t, <-outgoing, "Cannot subscribe to unknown room: room4.")
}

func TestSubscribeCommand_HasDuplicateNicks_DuplicatesToOutgoing(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	hub.CreateRoom("room3")
	hub.CreateRoom("room4")
	hub.SubscribeToRoom("id2", "room1", subscriber{nick: "nick1"})
	hub.SubscribeToRoom("id2", "room3", subscriber{nick: "nick3"})
	outgoing := make(chan message, 2)

	cmd := NewSubscribeCommand(hub)
	cmd.Handle("id1", "room1:nick1|room2:nick2|room3:nick3|room4:nick4", outgoing)

	assert.Contains(t, hub.getSubscribers("room2"), identity("id1"))
	assert.Contains(t, hub.getSubscribers("room4"), identity("id1"))

	assert.Len(t, outgoing, 2)
	assert.Contains(t, <-outgoing, "User nick1 already joined room1.")
	assert.Contains(t, <-outgoing, "User nick3 already joined room3.")
}

func TestSubscribeCommand_InvalidArgs_ErrorToOutgoing(t *testing.T) {
	testCases := []struct {
		args  string
		reply string
	}{
		{args: "", reply: "Room name is missing."},
		{args: "room1", reply: "Nickname for room1 is missing."},
		{args: "room1:nick1|room2", reply: "Nickname for room2 is missing."},
		{args: ":nick1", reply: "Room name is missing."},
		{args: "room1:nick1|:nick2", reply: "Room name is missing."},
	}

	for _, testCase := range testCases {
		hub := NewHub(128)
		hub.CreateRoom("room1")
		outgoing := make(chan message, 1)

		cmd := NewSubscribeCommand(hub)
		cmd.Handle("id", testCase.args, outgoing)

		assert.Len(t, outgoing, 1)
		assert.Contains(t, <-outgoing, testCase.reply)
		assert.Len(t, hub.getSubscribers("room1"), 0)
	}
}

func TestPulishCommand_CorrectArgs_MessagePublished(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	outgoing1 := make(chan message, 2)
	outgoing2 := make(chan message, 2)
	outgoing3 := make(chan message, 1)
	hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1", outgoing: outgoing1})
	hub.SubscribeToRoom("id1", "room2", subscriber{nick: "nick1", outgoing: outgoing1})
	hub.SubscribeToRoom("id2", "room1", subscriber{nick: "nick2", outgoing: outgoing2})
	hub.SubscribeToRoom("id2", "room2", subscriber{nick: "nick2", outgoing: outgoing2})
	hub.SubscribeToRoom("id3", "room2", subscriber{nick: "nick3", outgoing: outgoing3})

	cmd := NewPublishCommand(hub, 254)
	cmd.Handle("id1", "room1|msg1", make(chan message))
	cmd.Handle("id2", "room2|msg2", make(chan message))
	cmd.Handle("id3", "room2|msg3", make(chan message))

	assert.Equal(t, message("nick2@room2: msg2"), <-outgoing1)
	assert.Equal(t, message("nick3@room2: msg3"), <-outgoing1)
	assert.Equal(t, message("nick1@room1: msg1"), <-outgoing2)
	assert.Equal(t, message("nick3@room2: msg3"), <-outgoing2)
	assert.Equal(t, message("nick2@room2: msg2"), <-outgoing3)
}

func TestPulishCommand_RoomNotSubscribed_UnknownToOutgoing(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	outgoing := make(chan message, 1)
	hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1"})

	cmd := NewPublishCommand(hub, 254)
	cmd.Handle("id1", "room2|msg1", outgoing)

	assert.Len(t, outgoing, 1)
	assert.Contains(t, <-outgoing, "You are not subscribed to room2.")
}

func TestPulishCommand_InvalidArgs_ErrorToOutgoing(t *testing.T) {
	testCases := []struct {
		args  string
		reply string
	}{
		{args: "", reply: "Target room name is missing."},
		{args: "|", reply: "Target room name is missing."},
		{args: "room1", reply: "Message is empty."},
		{args: "room1|", reply: "Message is empty."},
		{args: "room1| ", reply: "Message is empty."},
	}

	for _, testCase := range testCases {
		hub := NewHub(128)
		hub.CreateRoom("room1")
		hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1"})
		outgoing := make(chan message, 1)

		cmd := NewPublishCommand(hub, 254)
		cmd.Handle("id", testCase.args, outgoing)

		assert.Len(t, outgoing, 1)
		assert.Contains(t, <-outgoing, testCase.reply)
	}
}

func TestPulishCommand_MessageExceedsCapacity_ErrorToOutgoing(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	outgoing1 := make(chan message, 1)
	outgoing2 := make(chan message, 1)
	hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1", outgoing: outgoing1})
	hub.SubscribeToRoom("id2", "room1", subscriber{nick: "nick2", outgoing: outgoing2})

	cmd := NewPublishCommand(hub, 4)

	cmd.Handle("id1", "room1|mmm", outgoing1)
	assert.Equal(t, message("nick1@room1: mmm"), <-outgoing2)

	cmd.Handle("id1", "room1|mmmm", outgoing1)
	assert.Equal(t, message("nick1@room1: mmmm"), <-outgoing2)

	cmd.Handle("id1", "room1|mmmmm", outgoing1)
	assert.Equal(t, message("Message is too long."), <-outgoing1)

	cmd.Handle("id1", "room1|mmmmmm", outgoing1)
	assert.Equal(t, message("Message is too long."), <-outgoing1)
}

func TestPulishCommand_MessagePublished_RoomHistoryAppended(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1", outgoing: make(chan message)})
	hub.SubscribeToRoom("id1", "room2", subscriber{nick: "nick1", outgoing: make(chan message)})
	expectedItem := historyItem{nick: "nick1", msg: "msg1"}

	cmd := NewPublishCommand(hub, 254)
	cmd.Handle("id1", "room1|msg1", make(chan message))

	assert.Equal(t, expectedItem, hub.rooms["room1"].history.Front().Value)
	assert.Equal(t, 0, hub.rooms["room2"].history.Len())
}
