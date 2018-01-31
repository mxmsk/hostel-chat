package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHubCreateRoom_GivenName_RoomAdded(t *testing.T) {
	hub := NewHub(128)

	err1 := hub.CreateRoom("room1")
	err2 := hub.CreateRoom("room2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Contains(t, hub.rooms, "room1")
	assert.Contains(t, hub.rooms, "room2")
}

func TestHubCreateRoom_NewRoom_RoomInitialized(t *testing.T) {
	hub := NewHub(128)

	hub.CreateRoom("room1")
	room := hub.rooms["room1"]

	assert.Equal(t, "room1", room.name)
	assert.NotNil(t, room.subscribers)
	assert.NotNil(t, room.history)
}

func TestHubCreateRoom_DuplicateRoom_ErrorReturned(t *testing.T) {
	hub := NewHub(128)

	err1 := hub.CreateRoom("room1")
	err2 := hub.CreateRoom("room1")

	assert.NoError(t, err1)
	assert.EqualError(t, err2, "Attempt to create duplicate room: room1")
}

func TestHubSubscribeToRoom_RoomsExist_UserSubscribedToRoom(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	sub := subscriber{
		nick:     "nick1",
		outgoing: make(chan<- message),
	}

	err := hub.SubscribeToRoom("id1", "room1", sub)
	assert.NoError(t, err)

	assert.Equal(t, sub, hub.rooms["room1"].subscribers["id1"])
	assert.Len(t, hub.rooms["room2"].subscribers, 0)
}

func TestHubSubscribeToRooms_RoomDoesntExist_ErrorReturned(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room2")
	hub.CreateRoom("room3")
	sub := subscriber{nick: "nick1"}

	err := hub.SubscribeToRoom("id1", "room1", sub)

	assert.EqualError(t, err, "Cannot subscribe to unknown room: room1")
}

func TestHubSubscribeToRooms_DuplicateNicks_ErrorReturned(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.rooms["room1"].subscribers = map[identity]subscriber{
		"id1": subscriber{nick: "nick1"},
	}
	sub1 := subscriber{nick: "nick1"}
	sub2 := subscriber{nick: "NiCK1"}

	err := hub.SubscribeToRoom("id2", "room1", sub1)
	assert.EqualError(t, err, "User nick1 already joined room1")

	err = hub.SubscribeToRoom("id2", "room1", sub2)
	assert.EqualError(t, err, "User nick1 already joined room1")
}

func TestHubGetSubscribers_RoomExists_SubscribersReturned(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	sub11 := subscriber{nick: "nick11", outgoing: make(chan<- message)}
	sub12 := subscriber{nick: "nick12", outgoing: make(chan<- message)}
	sub21 := subscriber{nick: "nick21", outgoing: make(chan<- message)}
	sub22 := subscriber{nick: "nick22", outgoing: make(chan<- message)}

	hub.rooms["room1"].subscribers = map[identity]subscriber{
		"id1": sub11,
		"id2": sub12,
	}
	hub.rooms["room2"].subscribers = map[identity]subscriber{
		"id1": sub21,
		"id2": sub22,
	}

	subscribers1 := hub.getSubscribers("room1")
	subscribers2 := hub.getSubscribers("room2")

	assert.Len(t, subscribers1, 2)
	assert.Len(t, subscribers2, 2)
	assert.Equal(t, sub11, subscribers1["id1"])
	assert.Equal(t, sub12, subscribers1["id2"])
	assert.Equal(t, sub21, subscribers2["id1"])
	assert.Equal(t, sub22, subscribers2["id2"])
}

func TestHubGetSubscribers_RoomDoesntExist_EmptyReturned(t *testing.T) {
	hub := NewHub(128)
	subscribers := hub.getSubscribers("room1")
	assert.Empty(t, subscribers)
}

func TestHubAppendRoomHistory_RoomExists_ExtendsHistory(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	item1 := historyItem{nick: "nick1", msg: "msg1"}
	item2 := historyItem{nick: "nick1", msg: "msg2"}

	err1 := hub.AppendRoomHistory("room1", item1)
	err2 := hub.AppendRoomHistory("room1", item2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	room := hub.rooms["room1"]
	assert.Equal(t, item2, room.history.Prev().Value)
	assert.Equal(t, item1, room.history.Prev().Prev().Value)
}

func TestHubAppendRoomHistory_RoomDoesntExist_ErrorReturned(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")

	err := hub.AppendRoomHistory("room2", historyItem{nick: "nick1", msg: "msg2"})

	assert.EqualError(t, err, "Cannot save history for unknown room: room2")
}

func TestHubNewHub_GivenCapacity_ExpectCorrectHistoryRingLen(t *testing.T) {
	for i := 2; i < 4; i++ {
		hub := NewHub(i)
		hub.CreateRoom("room1")
		assert.Equal(t, i, hub.rooms["room1"].history.Len())
	}
}

func TestHubgetRoomHistory_RoomExists_HistoryItemsReturned(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	item1 := historyItem{nick: "nick1", msg: "msg1"}
	item2 := historyItem{nick: "nick1", msg: "msg2"}
	item3 := historyItem{nick: "nick1", msg: "msg3"}
	item4 := historyItem{nick: "nick1", msg: "msg4"}

	hub.rooms["room1"].history.Value = item1
	hub.rooms["room1"].history.Next().Value = item2
	hub.rooms["room2"].history.Value = item3
	hub.rooms["room2"].history.Next().Value = item4

	history1 := hub.getRoomHistory("room1")
	history2 := hub.getRoomHistory("room2")

	assert.Len(t, history1, 2)
	assert.Len(t, history1, 2)
	assert.Contains(t, history1, item1)
	assert.Contains(t, history1, item2)
	assert.Contains(t, history2, item3)
	assert.Contains(t, history2, item4)
}

func TestHubgetRoomHistory_RoomDoesntExist_EmptyReturned(t *testing.T) {
	hub := NewHub(128)
	history := hub.getRoomHistory("room1")
	assert.Empty(t, history)
}

func TestHubUnsubscribe_GivenUser_UserRemovedFromAllRooms(t *testing.T) {
	hub := NewHub(128)
	hub.CreateRoom("room1")
	hub.CreateRoom("room2")
	hub.CreateRoom("room3")

	hub.SubscribeToRoom("id1", "room1", subscriber{nick: "nick1"})
	hub.SubscribeToRoom("id1", "room2", subscriber{nick: "nick1"})
	hub.SubscribeToRoom("id2", "room2", subscriber{nick: "nick2"})
	hub.SubscribeToRoom("id2", "room3", subscriber{nick: "nick2"})

	hub.Unsubscribe("id1")

	assert.NotContains(t, hub.rooms["room1"].subscribers, identity("id1"))
	assert.NotContains(t, hub.rooms["room2"].subscribers, identity("id1"))
	assert.Contains(t, hub.rooms["room2"].subscribers, identity("id2"))

	hub.Unsubscribe("id2")

	assert.NotContains(t, hub.rooms["room2"].subscribers, identity("id2"))
	assert.NotContains(t, hub.rooms["room3"].subscribers, identity("id2"))
}
