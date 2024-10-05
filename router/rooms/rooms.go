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
	players      *users.UserMap[player]
	moderators   *users.UserMap[moderator]
}

type player struct {
	name  string
	token string
}
type moderator struct{}

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

func (r *Room) getPlayer(eventChan chan string) player {
	player := r.players.Get(eventChan)
	return player
}

//func (r *Room) SetPlayerName(token string, name string) {
//	player := r.players.Get(eventChan)
//}

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
	names := make([]string, r.players.NumUsers())
	for i := 0; i < r.players.NumUsers(); i++ {
		names[i] = fmt.Sprintf("Player %d", i+1)
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

func (r *Room) AddModerator() chan string {
	moderator := moderator{}
	return r.moderators.New(moderator)
}

func (r *Room) AddPlayer() chan string {
	// Note: do not to send any events to the channel in this function, it will deadlock
	// because the function that calls hasn't set up any listeners yet
	token := uuid.New().String()
	player := player{"Test", token}
	return r.players.New(player)
}

func (r *Room) RemoveModerator(eventChan chan string) {
	r.moderators.Delete(eventChan)
}

func (r *Room) RemovePlayer(eventChan chan string) {
	r.players.Delete(eventChan)
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())
}

func (r *Room) InitializePlayer(eventChan chan string) {
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())

	eventChan <- events.PlayerBuzzerEvent(r.GetCurrentBuzzer())
	player := r.getPlayer(eventChan)
	tokenInput := TokenInput(player.token)
	eventChan <- events.TokenEvent(tokenInput)
}

func (r *Room) InitializeModerator(eventChan chan string) {
	eventChan <- r.CurrentPlayersEvent()
}
