package lib

import "strconv"

type Room struct {
	room_id       int64
	buzzer_status string
}

var TEST_ROOMS = [2]Room{
	{7, "Room 1"},
	{200, "Room 2"},
}

func GetAllRooms() [2]Room {
	return TEST_ROOMS
}

func GetRoom(room_id string) *Room {
	id, error := strconv.Atoi(room_id)
	if error != nil {
		return nil
	}

	room := Room{int64(id), "Unlocked"}
	return &room
}

func (r Room) Id() int64 {
	return r.room_id
}

func (r Room) Status() string {
	return r.buzzer_status
}
