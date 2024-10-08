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
	player := room.players.Get(token)
	player.SetName(name)
	room.players.SendToAll(room.CurrentPlayersEvent())
	room.moderators.SendToAll(room.CurrentPlayersEvent())
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

	room.buzzerStatus = Waiting
	player := room.getPlayer(token)
	if player == nil {
		logger.Error("nil Player returned for token %v")
		return
	}
	logMessage := fmt.Sprintf("%s Buzzed", player.Name())

	room.moderators.SendToAll(
		events.ModeratorStatusEvent("Waiting"),
		events.ModeratorLogEvent(logMessage),
	)

	room.players.SendToAll(
		events.LockedBuzzerEvent(),
		events.PlayerLogEvent(logMessage),
	)
}

func (room *Room) Reset() {
	logger.Debug("Sending unlock message")
	room.buzzerStatus = Unlocked

	room.moderators.SendToAll(
		events.ModeratorStatusEvent("Unlocked"),
		events.ModeratorLogEvent("Buzzer Unlocked"),
	)

	room.players.SendToAll(
		events.ReadyBuzzerEvent(),
		events.PlayerLogEvent("Buzzer Unlocked"),
	)
}

func (room *Room) CurrentPlayersEvent() string {
	users := room.players.GetUsers()
	names := make([]string, len(users))
	for i, player := range users {
		names[i] = player.Name()
	}

	return events.ModeratorPlayerListEvent(names)
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
	room.players.SendToAll(room.CurrentPlayersEvent())
	room.moderators.SendToAll(room.CurrentPlayersEvent())

	player.Channel() <- room.CurrentBuzzerEvent()
	player.Channel() <- events.TokenEvent(token)

	return token, closeChan
}

func (room *Room) CreateModerator(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()
	moderator := users.NewModerator(eventChan)
	room.moderators.Insert(token, moderator)

	// Initialize Moderator
	eventChan <- room.CurrentPlayersEvent()
	return token, closeChan
}

func (room *Room) RemoveModerator(token string) {
	room.moderators.CloseAndDelete(token)
}

func (room *Room) RemovePlayer(token string) {
	room.players.CloseAndDelete(token)
	room.players.SendToAll(room.CurrentPlayersEvent())
	room.moderators.SendToAll(room.CurrentPlayersEvent())
}
