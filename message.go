package main

import (
	"encoding/json"
	"math/rand"
	"time"
)

// A message id must fit within 53 bits, since that's Javascript's max int size.
// The Unix timestamp should be <= 34 bits for the next half millennium,
// so this id will fit within this constraint.
type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"time"` // Must be UTC
	Id        uint64    `json:"id`
}

func NewMessage(content string, randomIdComponent uint16) Message {
	now := time.Now().UTC().Round(time.Second)
	id := uint64(now.Unix()<<16) | uint64(randomIdComponent)
	return Message{content, now, id}
}

func (message *Message) JSON() (string, error) {
	marshaled, err := json.Marshal(*message)
	return string(marshaled), err
}

// The rationale behind using a random number in a message's id is to make it
// difficult to tell if any messages have been deleted. This does make it
// impossible to sort messages sent in the same second chronologically based
// on just their id, but messages are stored chronologically in a room's list
// of messages.
type RandomIdComponentGenerator struct {
	randNum             *rand.Rand
	lastGeneratedAt     time.Time
	generatedThisSecond map[uint16]bool
}

// Allows up to 65,536 messages in a room per second.
// Validating random numbers are unique is probably not necessary but done anyway.
func (generator *RandomIdComponentGenerator) IdComponent() uint16 {
	randomIdComponent := uint16(generator.randNum.Uint32())
	if now := time.Now().UTC().Round(time.Second); !now.Equal(generator.lastGeneratedAt) {
		generator.lastGeneratedAt = now
		generator.generatedThisSecond = map[uint16]bool{randomIdComponent: true}
		return randomIdComponent
	}
	// Ensure that the number is unique
	for generator.generatedThisSecond[randomIdComponent] {
		randomIdComponent = uint16(generator.randNum.Uint32())
	}
	generator.generatedThisSecond[randomIdComponent] = true
	return randomIdComponent
}

func NewRandomIdComponentGenerator() RandomIdComponentGenerator {
	return RandomIdComponentGenerator{randNum: rand.New(rand.NewSource(time.Now().Unix()))}
}

type C2S_CreateMessageRequest struct {
	Content string `json:"content" form:"content" binding:"required"`
}
