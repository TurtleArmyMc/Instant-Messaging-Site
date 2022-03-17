package main

import "sync"

type Room struct {
	messages             []*Message
	listeners            map[chan Message]struct{}
	idComponentGenerator RandomIdComponentGenerator
	rw                   sync.RWMutex // Locks room contents
}

func NewRoom() *Room {
	return &Room{
		listeners:            map[chan Message]struct{}{},
		idComponentGenerator: NewRandomIdComponentGenerator(),
	}
}

func (room *Room) MemberCount() int {
	room.rw.RLock()
	defer room.rw.RUnlock()
	return len(room.listeners)
}

func (room *Room) AddUser() chan Message {
	room.rw.Lock()
	defer room.rw.Unlock()
	listener := make(chan Message)
	room.listeners[listener] = struct{}{}
	return listener
}

// Returns if room is now empty
func (room *Room) RemoveUser(user chan Message) bool {
	room.rw.Lock()
	defer room.rw.Unlock()
	delete(room.listeners, user)
	return len(room.listeners) == 0
}

func (room *Room) PostMessage(content string) {
	room.rw.Lock()
	defer room.rw.Unlock()
	message := NewMessage(content, room.idComponentGenerator.IdComponent())
	room.messages = append(room.messages, &message)
	for listener := range room.listeners {
		// TODO: Figure out if this needs to be in a go-routine to avoid blocking
		listener <- message
	}
}

func (room *Room) GetMessages() []*Message {
	room.rw.RLock()
	defer room.rw.RUnlock()
	return append([]*Message{}, room.messages...)
}
