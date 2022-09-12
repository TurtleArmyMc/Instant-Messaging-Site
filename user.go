package main

type Connection struct {
	outgoing chan<- *Message
	Id       uint
	User     UId
}

type UId uint

type Session2UId struct {
	ids    map[string]UId
	nextId UId
}

func NewSession2UId() Session2UId {
	return Session2UId{ids: map[string]UId{}}
}

func (s2u *Session2UId) UId(session string) UId {
	id, ok := s2u.ids[session]
	if !ok {
		s2u.nextId++
		id = s2u.nextId
		s2u.ids[session] = id
	}
	return id
}