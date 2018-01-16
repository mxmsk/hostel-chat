package chat

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

// Hub represents chat database.
// For simplicity, we don't introduce abstraction of DAO.
// Instead, a hub is in-memory dataset that someday (maybe)
// will be managed by data-access service.
type Hub struct {
	rooms          map[string]*room
	roomHistoryCap int
}

type subscriber struct {
	nick     string
	outgoing chan<- message
}

type room struct {
	name        string
	subscribers map[identity]subscriber
	sm          sync.RWMutex
	history     *list.List
	hm          sync.Mutex
}

type historyItem struct {
	nick string
	msg  message
}

// NewHub creates a new hub, the storage of chat rooms.
func NewHub(roomHistoryCap int) *Hub {
	return &Hub{
		rooms:          make(map[string]*room),
		roomHistoryCap: roomHistoryCap,
	}
}

// CreateRoom adds to hub a new room with the specified name.
func (hub *Hub) CreateRoom(roomName string) error {
	if _, exists := hub.rooms[roomName]; exists {
		return fmt.Errorf("Attempt to create duplicate room: %s", roomName)
	}
	hub.rooms[roomName] = &room{
		name:        roomName,
		subscribers: make(map[identity]subscriber),
		history:     list.New(),
	}
	return nil
}

// SubscribeToRoom subsribes the specified user to room by assigning
// corresponding nick.
func (hub *Hub) SubscribeToRoom(user identity, roomName string, sub subscriber) error {
	if room, ok := hub.rooms[roomName]; ok {
		room.sm.Lock()
		defer room.sm.Unlock()
		for _, roomSub := range room.subscribers {
			if strings.EqualFold(roomSub.nick, sub.nick) {
				return fmt.Errorf("User %s already joined %s", roomSub.nick, roomName)
			}
		}
		room.subscribers[user] = sub
		return nil
	}
	return fmt.Errorf("Cannot subscribe to unknown room: %s", roomName)
}

func (hub *Hub) getSubscribers(roomName string) map[identity]subscriber {
	if room, ok := hub.rooms[roomName]; ok {
		room.sm.RLock()
		defer room.sm.RUnlock()
		result := make(map[identity]subscriber, len(room.subscribers))
		for k, v := range room.subscribers {
			result[k] = v
		}
		return result
	}
	return nil
}

// AppendRoomHistory extends history of a given room with the specified item.
func (hub *Hub) AppendRoomHistory(roomName string, item historyItem) error {
	if room, ok := hub.rooms[roomName]; ok {
		room.hm.Lock()
		defer room.hm.Unlock()
		h := room.history
		if h.Len() == hub.roomHistoryCap {
			h.Remove(h.Front())
		}
		h.PushBack(item)
		return nil
	}
	return fmt.Errorf("Cannot save history for unknown room: %s", roomName)
}

func (hub *Hub) getRoomHistory(roomName string) []historyItem {
	var history []historyItem
	if room, ok := hub.rooms[roomName]; ok {
		room.hm.Lock()
		defer room.hm.Unlock()
		history = make([]historyItem, 0, room.history.Len())
		h := room.history.Front()
		for h != nil {
			history = append(history, h.Value.(historyItem))
			h = h.Next()
		}
	}
	return history
}

// Unsubscribe removes user with the specified id from all rooms.
func (hub *Hub) Unsubscribe(user identity) {
	for _, room := range hub.rooms {
		room.sm.Lock()
		delete(room.subscribers, user)
		room.sm.Unlock()
	}
}
