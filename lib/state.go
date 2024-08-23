package lib

import (
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
  listeners     map[chan string]struct{}
}

var TEST_ROOMS = map[int]*Room {
  7: {7, Unlocked, make(map[chan string]struct{})},
  200: {200, Unlocked, make(map[chan string]struct{})},
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
  for listener := range r.listeners {
    sse := FormatEvent("status", "<span>Waiting<span>")
    listener <- sse
  }
}

func (r *Room) Reset() {
  r.buzzer_status = Unlocked
  for listener := range r.listeners {
    sse := FormatEvent("status", "<span>Unlocked<span>")
    listener <- sse
  }
}

func (r *Room) AddListener() chan string {
  // Create a channel and add it to the room's list of channels
  eventChan := make(chan string)
  r.listeners[eventChan] = struct{}{}

  return eventChan
}

func (r *Room) RemoveListener(listener chan string) {
  delete(r.listeners, listener)
}
