package rooms

import (
	"context"
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
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
  name          string
	buzzer_status BuzzerStatus
	players       map[chan string]struct{}
	moderators    map[chan string]struct{}
}

var rooms = map[int64]*Room{}

func NewRoom(room_id int64, name string) *Room {
	return &Room{
		room_id,
    name,
		Unlocked,
		make(map[chan string]struct{}),
		make(map[chan string]struct{}),
	}
}

func getOrCreateRoom(room_id int64, name string) *Room {
  room := rooms[room_id]
  if room == nil {
    room = NewRoom(room_id, name)
    rooms[room_id] = room
  }

  return room
}

func GetRoom(ctx context.Context, room_id string) *Room {
	id, error := strconv.ParseInt(room_id, 10, 64)
	if error != nil {
		return nil
	}

	dbRoom := db.GetRoom(ctx, id)
  room := getOrCreateRoom(dbRoom.Id(), dbRoom.Name())
  return room
}

func (r *Room) Id() int64 {
	return r.room_id
}

func (r *Room) Name() string {
	return r.name
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

	for listener := range r.players {
		buzzer := lib.ToString(BuzzerButton(true))
		listener <- lib.FormatEvent("log", buzzer)
		listener <- lib.FormatEvent("log", "<div>Player Buzzed<div>")
	}
}

func (r *Room) Reset() {
	r.buzzer_status = Unlocked
	for listener := range r.moderators {
		listener <- lib.FormatEvent("status", "<span>Unlocked<span>")
	}

	for listener := range r.players {
		fmt.Println("Sending unlock message")
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

func (r *Room) RemoveModerator(listener chan string) {
	delete(r.moderators, listener)
}

func (r *Room) RemovePlayer(listener chan string) {
	delete(r.players, listener)
}
