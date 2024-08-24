package rooms

import (
	"context"
	"fmt"
	"goodbuzz/lib"
	"strconv"
)

type BuzzerStatus int

const (
	Unlocked BuzzerStatus = 0
	Waiting               = 1
	Locked                = 2
)

func (s BuzzerStatus) String() string {
	switch s {
	case Unlocked:
		return "Unlocked"
	case Waiting:
		return "Waiting"
	case Locked:
		return "Locked"
	}

	return "Unknown"
}

type Room struct {
	room_id       int64
	buzzer_status BuzzerStatus
	players       map[chan string]struct{}
	moderators    map[chan string]struct{}
}

var TEST_ROOMS = map[int]*Room{
	7:   NewRoom(7),
	200: NewRoom(200),
}

func NewRoom(room_id int64) *Room {
	return &Room{
		room_id,
		Unlocked,
		make(map[chan string]struct{}),
		make(map[chan string]struct{}),
	}
}

func GetRoom(room_id string) *Room {
	id, error := strconv.Atoi(room_id)
	if error != nil {
		return nil
	}

	room := TEST_ROOMS[id]
	return room
}

func (r *Room) Id() int64 {
	return r.room_id
}

func (r *Room) Status() BuzzerStatus {
	return r.buzzer_status
}

func (r *Room) StatusString() string {
	return r.buzzer_status.String()
}

// TODO need a way to ignore buzzes that came in before the reset
func (r *Room) BuzzRoom() {
	r.buzzer_status = Waiting
	for listener := range r.moderators {
		listener <- lib.FormatEvent("status", "<span>Waiting<span>")
	}
}

func (r *Room) Reset() {
	r.buzzer_status = Unlocked
	for listener := range r.moderators {
		listener <- lib.FormatEvent("status", "<span>Unlocked<span>")
	}

	for listener := range r.players {
		fmt.Printf("Sending unlock message")
		buzzer := lib.ToString(BuzzerButton(false))
		listener <- lib.FormatEvent("log", "<div>Buzzer Unlocked<div>")
		listener <- lib.FormatEvent("log", buzzer)
	}
}

func (r *Room) AddModerator() chan string {
	// Create a channel and add it to the room's list of channels
	eventChan := make(chan string)
	r.moderators[eventChan] = struct{}{}
	return eventChan
}

func (r *Room) AddPlayer() chan string {
	// Create a channel and add it to the room's list of channels
	eventChan := make(chan string)
	r.players[eventChan] = struct{}{}
	return eventChan
}

func (r *Room) RemoveListener(listener chan string) {
	delete(r.moderators, listener)
}
