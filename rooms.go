package main

import "sync"

var roomManager = RoomManager{rooms: map[string]*Room{}}

type Room struct {
	messages  []string
	listeners map[chan string]struct{}
	rw        sync.RWMutex // Locks room contents
}

func (room *Room) MemberCount() int {
	room.rw.RLock()
	defer room.rw.RUnlock()
	return len(room.listeners)
}

func (room *Room) AddUser() chan string {
	room.rw.Lock()
	defer room.rw.Unlock()
	listener := make(chan string)
	room.listeners[listener] = struct{}{}
	return listener
}

// Returns if room is now empty
func (room *Room) RemoveUser(user chan string) bool {
	room.rw.Lock()
	defer room.rw.Unlock()
	delete(room.listeners, user)
	return len(room.listeners) == 0
}

func (room *Room) PostMessage(message string) {
	room.rw.Lock()
	defer room.rw.Unlock()
	room.messages = append(room.messages, message)
	for listener := range room.listeners {
		// TODO: Figure out if this needs to be in a go-routine to avoid blocking
		listener <- message
	}
}

func (room *Room) GetMessages() []string {
	room.rw.RLock()
	defer room.rw.RUnlock()
	return append([]string{}, room.messages...)
}

type RoomManager struct {
	rooms map[string]*Room
	rw    sync.RWMutex // Locks map
}

type RoomInfo struct {
	RoomName    string `json:"room_name"`
	MemberCount int    `json:"member_count"`
}

func (manager *RoomManager) GetRooms() []RoomInfo {
	manager.rw.RLock()

	ret := make([]RoomInfo, 0, len(manager.rooms))
	for roomName, room := range manager.rooms {
		ret = append(ret, RoomInfo{roomName, room.MemberCount()})
	}

	manager.rw.RUnlock()

	return ret
}

func (manager *RoomManager) GetRoomMessages(roomName string) []string {
	manager.rw.RLock()
	room, ok := manager.rooms[roomName]
	manager.rw.RUnlock()
	if ok {
		return append([]string{}, room.messages...)
	}
	return []string{}
}

func (manager *RoomManager) getOrCreateRoom(roomName string) *Room {
	manager.rw.Lock()
	room, ok := manager.rooms[roomName]
	if !ok {
		room = &Room{listeners: map[chan string]struct{}{}}
		manager.rooms[roomName] = room
	}
	manager.rw.Unlock()
	return room
}

// Returns listener for new messages
func (manager *RoomManager) AddUserToRoom(roomName string) chan string {
	room := manager.getOrCreateRoom(roomName)
	return room.AddUser()
}

func (manager *RoomManager) RemoveUserFromRoom(roomName string, userListener chan string) {
	manager.rw.Lock()

	room, ok := manager.rooms[roomName]
	if ok {
		if empty := room.RemoveUser(userListener); empty {
			delete(manager.rooms, roomName)
		}
	}

	manager.rw.Unlock()
}

func (manager *RoomManager) PostInRoom(name, message string) {
	manager.rw.Lock()
	room, ok := manager.rooms[name]
	if ok {
		room.PostMessage(message)
	}
	manager.rw.Unlock()
}
