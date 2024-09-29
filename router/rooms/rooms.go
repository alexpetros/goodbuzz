package rooms

import (
	"context"
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"sync"

	"github.com/a-h/templ"
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

type channelMap struct {
	sync.RWMutex
	channels map[chan string]struct{}
}

func newChannelMap() *channelMap {
	return &channelMap{
		channels: make(map[chan string]struct{}),
	}
}

func (cm *channelMap) new() chan string {
	eventChan := make(chan string)
	cm.Lock()
	defer cm.Unlock()

	cm.channels[eventChan] = struct{}{}
	return eventChan
}

func (cm *channelMap) delete(eventChan chan string) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.channels, eventChan)
	close(eventChan)
}

func (cm *channelMap) sendToAll(message string) {
	cm.RLock()
	for listener := range cm.channels {
		listener <- message
	}
	cm.RUnlock()
}

type Room struct {
	room_id       int64
	name          string
	buzzer_status BuzzerStatus
	players       *channelMap
	moderators    *channelMap
}

type roomMap struct {
	sync.Mutex
	internal map[int64]*Room
}

var openRooms = roomMap{
	internal: make(map[int64]*Room),
}

func NewRoom(room_id int64, name string) *Room {
	return &Room{
		room_id:       room_id,
		name:          name,
		buzzer_status: Unlocked,
		players:       newChannelMap(),
		moderators:    newChannelMap(),
	}
}

func getOrCreateRoom(room_id int64, name string) *Room {
	openRooms.Lock()
	room := openRooms.internal[room_id]
	if room == nil {
		room = NewRoom(room_id, name)
		openRooms.internal[room_id] = room
	} else if room.name != name {
		room.name = name
	}
	openRooms.Unlock()

	return room
}

func GetRoomsForTournament(ctx context.Context, tournament_id int64) []Room {
	dbRooms := db.GetRoomsForTournament(ctx, tournament_id)
	rooms := make([]Room, 0)
	for _, dbRoom := range dbRooms {
		newRoom := GetRoom(ctx, dbRoom.Id())
		rooms = append(rooms, *newRoom)
	}

	return rooms
}

func GetRoom(ctx context.Context, room_id int64) *Room {
	dbRoom := db.GetRoom(ctx, room_id)
	room := getOrCreateRoom(dbRoom.Id(), dbRoom.Name())
	return room
}

func (r *Room) Id() int64 {
	return r.room_id
}

func (r *Room) Name() string {
	return r.name
}

func (r *Room) Url() string {
	return fmt.Sprintf("/rooms/%d", r.room_id)
}

func (r *Room) PlayerUrl() string {
	return fmt.Sprintf("/rooms/%d/player", r.room_id)
}

func (r *Room) ModeratorUrl() string {
	return fmt.Sprintf("/rooms/%d/moderator", r.room_id)
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
	r.moderators.RLock()
	for listener := range r.moderators.channels {
		listener <- lib.FormatEvent("status", "<span>Waiting<span>")
	}
	r.moderators.RUnlock()

	r.players.RLock()
	for listener := range r.players.channels {
		buzzer := r.GetCurrentBuzzer()
		listener <- lib.FormatEvent("buzzer", buzzer)
		listener <- lib.FormatEvent("log", "<div>Player Buzzed<div>")
	}
	r.players.RUnlock()
}

func (r *Room) Reset() {
	logger.Debug("Sending unlock message")
	r.buzzer_status = Unlocked

	moderatorStatus := lib.FormatEvent("status", "<span>Unlocked<span>")
	r.moderators.sendToAll(moderatorStatus)

	buzzer := lib.ToString(ReadyBuzzer())
	r.players.sendToAll(lib.FormatEvent("buzzer", buzzer))
	r.players.sendToAll(lib.FormatEvent("log", "<div>Buzzer Unlocked<div>"))
}

func (r *Room) GetCurrentStatus() string {
	buzzer := r.GetCurrentBuzzer()
	return lib.FormatEvent("buzzer", buzzer)
}

func (r *Room) SendCurrentStatus(eventChan chan string) {
	eventChan <- r.GetCurrentStatus()
}

func (r *Room) GetCurrentBuzzer() string {
  var buzzer templ.Component

	status  := r.buzzer_status
  if status == Unlocked {
    buzzer = ReadyBuzzer()
  } else if status == Waiting {
    buzzer = WaitingBuzzer()
  } else if status == Locked {
    buzzer = LockedBuzzer()
  }

  return lib.ToString(buzzer)
}


func (r *Room) AddModerator() chan string {
	return r.moderators.new()
}

func (r *Room) AddPlayer() chan string {
	eventChan := r.players.new()
	return eventChan
}

func (r *Room) RemoveModerator(eventChan chan string) {
	r.moderators.delete(eventChan)
}

func (r *Room) RemovePlayer(eventChan chan string) {
	r.players.delete(eventChan)
}
