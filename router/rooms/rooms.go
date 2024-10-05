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
	"io"
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
	players      *users.UserMap[player]
	moderators   *users.UserMap[moderator]
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

func (r *Room) getPlayer(token string) player {
	player := r.players.Get(token)
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

func (r *Room) CreatePlayer(w io.Writer, notify <-chan struct{}) chan struct{} {
	// Note: do not to send any events to the channel in this function, it will deadlock
	// because the function that calls hasn't set up any listeners yet
	token := uuid.New().String()
	eventChan := make(chan string)
	player := player{"Test", eventChan}
	r.players.Insert(token, player)

	// Listen for client close and delete channel when it happens
	closeChan := make(chan struct{})
	go func() {
		<-notify
		fmt.Printf("Player disconnected from room %d\n", r.Id())
		r.RemovePlayer(token)
		closeChan <- struct{}{}
	}()

	// Continuously send data to the client
	go func() {
		for {
			data := <-eventChan
			// This is what's received from a closed channel
			if data == "" {
				break
			}

			logger.Debug("Sending data to player in room %d:\n%s", r.Id(), data)
			_, err := fmt.Fprintf(w, data)
			if err != nil {
				logger.Error("Failed to send data to player in room %d:\n%s", r.Id(), data)
			}
			w.(http.Flusher).Flush()
		}
	}()

	// Initialize Player
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())

	tokenInput := TokenInput(token)

	player.channel <- events.PlayerBuzzerEvent(r.GetCurrentBuzzer())
	player.channel <- events.TokenEvent(tokenInput)

	return closeChan
}

func (r *Room) CreateModerator(w io.Writer, notify <-chan struct{}) chan struct{} {
	token := uuid.New().String()
	eventChan := make(chan string)
	moderator := moderator{eventChan}
	r.moderators.Insert(token, moderator)

	closeChan := make(chan struct{})
	go func() {
		<-notify
		fmt.Printf("Moderator disconnected from room %d", r.Id())
		r.RemoveModerator(token)
		closeChan <- struct{}{}
	}()

	// Continuously send data to the client
	go func() {
		for {
			data := <-eventChan
			if data == "" {
				break
			}

			logger.Debug("Sending data to moderator in room %d:\n%s", r.Id(), data)
			_, err2 := fmt.Fprintf(w, data)
			if err2 != nil {
				logger.Error("Failed to send data to moderatorr in room %d:\n%s", r.Id(), data)
			}
			w.(http.Flusher).Flush()
		}
	}()

	// Initialize Moderator
	eventChan <- r.CurrentPlayersEvent()
	return closeChan
}

func (r *Room) RemoveModerator(token string) {
	r.moderators.CloseAndDelete(token)
}

func (r *Room) RemovePlayer(token string) {
	r.players.CloseAndDelete(token)
	r.players.SendToAll(r.CurrentPlayersEvent())
	r.moderators.SendToAll(r.CurrentPlayersEvent())
}
