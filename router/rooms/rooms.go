package rooms

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
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
	players      *users.UserMap[*player]
	moderators   *users.UserMap[*moderator]
}

type player struct {
	name    string
	channel chan string
}

func (player player) Channel() chan string {
	return player.channel
}

type moderator struct {
	channel chan string
}

func (moderator moderator) Channel() chan string {
	return moderator.channel
}

var openRooms = newRoomMap()

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

func (r *Room) getPlayer(token string) *player {
	return r.players.Get(token)
}

func (r *Room) SetPlayerName(token string, name string) {
	player := r.players.Get(token)
	player.name = name
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())
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
func (r *Room) BuzzRoom() {
	r.buzzerStatus = Waiting

	r.moderators.SendToAll(
		events.ModeratorStatusEvent("Waiting"),
		events.ModeratorLogEvent("Player Buzzed"),
	)

	r.players.SendToAll(
		events.PlayerBuzzerEvent(LockedBuzzer()),
		events.PlayerLogEvent("Player Buzzed"),
	)
}

func (r *Room) Reset() {
	logger.Debug("Sending unlock message")
	r.buzzerStatus = Unlocked

	r.moderators.SendToAll(
		events.ModeratorStatusEvent("Unlocked"),
		events.ModeratorLogEvent("Buzzer Unlocked"),
	)

	r.players.SendToAll(
		events.PlayerBuzzerEvent(ReadyBuzzer()),
		events.PlayerLogEvent("Buzzer Unlocked"),
	)
}

func (r *Room) CurrentPlayersEvent() string {

	users := r.players.GetUsers()
	names := make([]string, len(users))
	for i, player := range users {
		names[i] = player.name
	}

	return events.ModeratorPlayerListEvent(names)
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

func (room *Room) CreatePlayer(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()
	tokenInput := TokenInput(token)
	player := player{"Test", eventChan}
	room.players.Insert(token, &player)

	// Initialize Player
	room.players.SendToAll(room.CurrentPlayersEvent())
	room.moderators.SendToAll(room.CurrentPlayersEvent())

	player.channel <- events.PlayerBuzzerEvent(room.GetCurrentBuzzer())
	player.channel <- events.TokenEvent(tokenInput)

	return token, closeChan
}

func (room *Room) CreateModerator(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()
	moderator := moderator{eventChan}
	room.moderators.Insert(token, &moderator)

	// Initialize Moderator
	eventChan <- room.CurrentPlayersEvent()
	return token, closeChan
}

func (r *Room) RemoveModerator(token string) {
	r.moderators.CloseAndDelete(token)
}

func (r *Room) RemovePlayer(token string) {
	r.players.CloseAndDelete(token)
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())
}
