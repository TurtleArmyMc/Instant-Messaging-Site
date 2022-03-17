package main

import "sync"

var roomManager = RoomManager{rooms: map[string]*Room{}}

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

func (manager *RoomManager) GetRoomMessages(roomName string) []*Message {
	manager.rw.RLock()
	room, ok := manager.rooms[roomName]
	manager.rw.RUnlock()
	if ok {
		return append([]*Message{}, room.messages...)
	}
	return []*Message{}
}

func (manager *RoomManager) getOrCreateRoom(roomName string) *Room {
	manager.rw.Lock()
	room, ok := manager.rooms[roomName]
	if !ok {
		room = NewRoom()
		manager.rooms[roomName] = room
	}
	manager.rw.Unlock()
	return room
}

// Returns listener for new messages
func (manager *RoomManager) AddUserToRoom(roomName string) chan Message {
	room := manager.getOrCreateRoom(roomName)
	return room.AddUser()
}

func (manager *RoomManager) RemoveUserFromRoom(roomName string, userListener chan Message) {
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
