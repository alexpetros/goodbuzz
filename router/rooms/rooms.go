package rooms

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms/events"
	"goodbuzz/router/rooms/users"
	"net/http"
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
	roomId       int64
	name         string
	buzzerStatus BuzzerStatus
	players      *users.UserMap[*users.Player]
	moderators   *users.UserMap[*users.Moderator]
}

var openRooms = newRoomMap()

func (room *Room) Id() int64 {
	return room.roomId
}

func (room *Room) Name() string {
	return room.name
}

func (room *Room) Url() string {
	return fmt.Sprintf("/rooms/%d", room.roomId)
}

func (room *Room) PlayerUrl() string {
	return fmt.Sprintf("/rooms/%d/Player", room.roomId)
}

func (room *Room) ModeratorUrl() string {
	return fmt.Sprintf("/rooms/%d/Moderator", room.roomId)
}

func (room *Room) Status() BuzzerStatus {
	return room.buzzerStatus
}

func (room *Room) StatusString() string {
	return room.buzzerStatus.String()
}

func (room *Room) getPlayer(token string) *users.Player {
	return room.players.Get(token)
}

func (room *Room) SetPlayerName(token string, name string) {
	logger.Debug("Setting %s name to %s", token, name)
	player := room.players.Get(token)
	player.SetName(name)
	room.sendPlayerListUpdates()
}

func (room *Room) sendLogUpdates(message string) {
	room.moderators.SendToAll(events.ModeratorLogEvent(message))
	room.players.SendToAll(events.PlayerLogEvent(message))
}

func (room *Room) sendBuzzerUpdates() {
	for _, player := range room.players.GetUsers() {
		if player.IsLocked() {
			player.Channel() <- events.LockedBuzzerEvent()
		} else {
			player.Channel() <- room.CurrentBuzzerEvent()
		}
	}

	statusString := room.buzzerStatus.String()
	room.moderators.SendToAll(events.ModeratorStatusEvent(statusString))
}

func (room *Room) sendPlayerListUpdates() {
	users := room.players.GetUsers()

	room.players.SendToAll(events.PlayerListEvent(users))
	room.moderators.SendToAll(events.ModeratorPlayerControlsEvent(users))
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
	// TODO handle case where this is nil
	dbRoom := db.GetRoom(ctx, roomId)
	room := openRooms.getOrCreateRoom(dbRoom.Id(), dbRoom.Name())
	return room
}

// TODO need a way to ignore buzzes that came in before the reset
func (room *Room) BuzzRoom(token string) {
	logger.Debug("Buzzing room for Player with token: %s", token)

	player := room.getPlayer(token)
	if player == nil {
		logger.Error("nil Player returned for token %v")
		return
	}
	room.buzzerStatus = Locked
	logMessage := fmt.Sprintf("%s Buzzed", player.Name())

	room.sendBuzzerUpdates()
	room.sendLogUpdates(logMessage)
}

func (room *Room) Reset() {
	logger.Debug("Sending unlock message")
	room.buzzerStatus = Unlocked

	room.sendBuzzerUpdates()
	room.sendLogUpdates("Buzzer Unlocked")
}

func (room *Room) CurrentBuzzerEvent() string {
	var buzzer string

	status := room.buzzerStatus
	if status == Unlocked {
		buzzer = events.ReadyBuzzerEvent()
	} else if status == Waiting {
		buzzer = events.WaitingBuzzerEvent()
	} else if status == Locked {
		buzzer = events.LockedBuzzerEvent()
	}

	return buzzer
}

func (room *Room) CreatePlayer(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()

	nameCookie, err := r.Cookie("name")
	var name string
	if err != nil {
		name = "New Player"
	} else {
		name = nameCookie.Value
	}
	player := users.NewPlayer(name, eventChan)

	room.players.Insert(token, player)

	// Initialize Player
	room.sendPlayerListUpdates()
	player.Channel() <- events.TokenEvent(token)
	player.Channel() <- room.CurrentBuzzerEvent()

	return token, closeChan
}

func (room *Room) CreateModerator(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()
	moderator := users.NewModerator(eventChan)
	room.moderators.Insert(token, moderator)

	// Initialize Moderator
	eventChan <- events.ModeratorPlayerControlsEvent(room.players.GetUsers())
	return token, closeChan
}

func (room *Room) RemoveModerator(token string) {
	room.moderators.CloseAndDelete(token)
}

func (room *Room) RemovePlayer(token string) {
	room.players.CloseAndDelete(token)
	room.sendPlayerListUpdates()
}
