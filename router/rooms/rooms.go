package rooms

import (
	"context"
	"fmt"
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

func (cm *channelMap) sendToAll(messages ...string) {
	message := CombineEvents(messages...)
	cm.RLock()
	for listener := range cm.channels {
		listener <- message
	}
	cm.RUnlock()
}

type Room struct {
	roomId       int64
	name         string
	buzzerStatus BuzzerStatus
	players      *channelMap
	moderators   *channelMap
}

type roomMap struct {
	sync.Mutex
	internal map[int64]*Room
}

var openRooms = roomMap{
	internal: make(map[int64]*Room),
}

func NewRoom(roomId int64, name string) *Room {
	return &Room{
		roomId:       roomId,
		name:         name,
		buzzerStatus: Unlocked,
		players:      newChannelMap(),
		moderators:   newChannelMap(),
	}
}

func (r *Room) Id() int64 {
	return r.roomId
}

func (r *Room) Name() string {
	return r.name
}

func (r *Room) Url() string {
	return fmt.Sprintf("/rooms/%d", r.roomId)
}

func (r *Room) PlayerUrl() string {
	return fmt.Sprintf("/rooms/%d/player", r.roomId)
}

func (r *Room) ModeratorUrl() string {
	return fmt.Sprintf("/rooms/%d/moderator", r.roomId)
}

func (r *Room) Status() BuzzerStatus {
	return r.buzzerStatus
}

func (r *Room) StatusString() string {
	return r.buzzerStatus.String()
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

func GetRoomsForTournament(ctx context.Context, tournamentId int64) []Room {
	dbRooms := db.GetRoomsForTournament(ctx, tournamentId)
	rooms := make([]Room, 0)
	for _, dbRoom := range dbRooms {
		newRoom := GetRoom(ctx, dbRoom.Id())
		rooms = append(rooms, *newRoom)
	}

	return rooms
}

func GetRoom(ctx context.Context, roomId int64) *Room {
	dbRoom := db.GetRoom(ctx, roomId)
	room := getOrCreateRoom(dbRoom.Id(), dbRoom.Name())
	return room
}

// TODO need a way to ignore buzzes that came in before the reset
func (r *Room) BuzzRoom() {
	r.buzzerStatus = Waiting

	r.moderators.sendToAll(
		ModeratorStatusEvent("Waiting"),
		ModeratorLogEvent("Player Buzzed"),
	)

	r.players.sendToAll(
		r.PlayerInitializeEvent(),
		PlayerLogEvent("Player Buzzed"),
	)
}

func (r *Room) Reset() {
	logger.Debug("Sending unlock message")
	r.buzzerStatus = Unlocked

	r.moderators.sendToAll(
		ModeratorStatusEvent("Unlocked"),
		ModeratorLogEvent("Buzzer Unlocked"),
	)

	r.players.sendToAll(
		PlayerBuzzerEvent(ReadyBuzzer()),
		PlayerLogEvent("Buzzer Unlocked"),
	)
}

func (r *Room) CurrentPlayersEvent() string {
	names := make([]string, len(r.players.channels))
	for i := 0; i < len(r.players.channels); i++ {
		names[i] = fmt.Sprintf("Player %d", i+1)
	}

	return ModeratorPlayerListEvent(names)
}

func (r *Room) GetCurrentBuzzer() templ.Component {
	var buzzer templ.Component

	status := r.buzzerStatus
	if status == Unlocked {
		buzzer = ReadyBuzzer()
	} else if status == Waiting {
		buzzer = WaitingBuzzer()
	} else if status == Locked {
		buzzer = LockedBuzzer()
	}

	return buzzer
}

func (r *Room) AddModerator() chan string {
	return r.moderators.new()
}

func (r *Room) AddPlayer() chan string {
	// Note: if you try to send any events in this function, it will deadlock
	// because the function that calls hasn't set up any listeners yet
	return r.players.new()
}

func (r *Room) RemoveModerator(eventChan chan string) {
	r.moderators.delete(eventChan)
}

func (r *Room) RemovePlayer(eventChan chan string) {
	r.players.delete(eventChan)
	r.players.sendToAll(r.CurrentPlayersEvent())
	r.moderators.sendToAll(r.CurrentPlayersEvent())
}

func (r *Room) PlayerInitializeEvent() string {
	r.players.sendToAll(r.CurrentPlayersEvent())
	r.moderators.sendToAll(r.CurrentPlayersEvent())

	buzzer := PlayerBuzzerEvent(r.GetCurrentBuzzer())
	return buzzer
}

func (r *Room) ModeratorInitializeEvent() string {
	players := r.CurrentPlayersEvent()
	return players
}
