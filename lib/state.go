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
}

var TEST_ROOMS = map[int]*Room {
  7: {7, Unlocked},
  200: {200, Unlocked},
}

// func GetAllRooms() [2]Room {
// 	return TEST_ROOMS
// }

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

func (r *Room) BuzzRoom() {
  r.buzzer_status = Waiting
}
