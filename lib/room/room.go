package room

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room/users"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	roomId       int64
	name         string
	logs         []Log
	buzzer			 *Buzzer
	players      *users.UserMap[*users.Player]
	moderators   *users.UserMap[*users.Moderator]
}

func (roomMap *RoomMap) newRoom(roomId int64, name string) *Room {
	room := Room{
		roomId:       roomId,
		name:         name,
		logs:         make([]Log, 0),
		players:      users.NewUserMap[*users.Player](),
		moderators:   users.NewUserMap[*users.Moderator](),
	}

	buzzer := NewBuzzer(room.sendBuzzerUpdates)
	room.buzzer = buzzer

	return &room
}

func (room *Room) sendBuzzerUpdates(buzzerUpdate BuzzerUpdate) {
	for _, player := range room.players.GetUsers() {
		if player.IsLocked() {
			player.Channel() <- LockedBuzzerEvent()
		} else {
			player.Channel() <- currentPlayerBuzzer(buzzerUpdate)
		}
	}

	moderatorStatusEvent := ModeratorStatusEvent(buzzerUpdate.status.String())
	room.moderators.SendToAll(moderatorStatusEvent)
}

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
	return room.buzzer.GetUpdate().status
}

func (room *Room) lockPlayer(token string) {
	room.players.Get(token).Lock()
}

func (room *Room) unlockAll() {
	room.players.Run(func(player *users.Player) {
		player.Unlock()
	})
}

func (room *Room) UnlockPlayer(token string) {
	logger.Debug("Unlocking player %s", token)
	room.players.Get(token).Unlock()
	room.buzzer.SendUpdates()
	room.sendPlayerListUpdates()
}

func (room *Room) SetPlayerName(token string, name string) {
	logger.Debug("Setting %s name to %s", token, name)
	player := room.players.Get(token)
	player.SetName(name)
	room.sendPlayerListUpdates()
}

func (room *Room) BuzzRoom(token string, resetToken string) {
	player := room.players.Get(token)
	if player == nil {
		logger.Error("nil player returned for token %v", token)
		return
	}

	room.buzzer.Buzz(token, resetToken)

	logMessage := fmt.Sprintf("Player %v buzzed room %v", player.Name(), room.Id())
	logger.Debug(logMessage)
	room.log(logMessage)
}

func (room *Room) ResetAll() {
	logger.Debug("Resetting all buzzers")
	room.unlockAll()
	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for everyone")
}

func (room *Room) ResetSome() {
	logger.Debug("Resetting some buzzers")

	winningToken := room.buzzer.GetUpdate().winner

	if winningToken == "" {
		logger.Info("room %s was reset with no active buzzes", room.name)
	} else {
		logger.Info("Locking player with token %s", winningToken)
		room.lockPlayer(winningToken)
	}

	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for some players")
}

func currentPlayerBuzzer(buzzerUpdate BuzzerUpdate) string {
	var playerBuzzer string
	if buzzerUpdate.status == Unlocked {
		playerBuzzer = ReadyBuzzerEvent(buzzerUpdate.resetToken)
	} else if buzzerUpdate.status == Waiting {
		playerBuzzer = WaitingBuzzerEvent()
	} else if buzzerUpdate.status == Locked {
		playerBuzzer = LockedBuzzerEvent()
	}

	return playerBuzzer
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
	player := users.NewPlayer(name, token, eventChan)

	room.players.Insert(token, player)

	// Initialize Player
	room.sendPlayerListUpdates()
	player.Channel() <- TokenEvent(token)
	player.Channel() <- PastLogsEvent(room.logs)
	player.Channel() <- currentPlayerBuzzer(room.buzzer.GetUpdate())

	return token, closeChan
}

func (room *Room) CreateModerator(w http.ResponseWriter, r *http.Request) (string, chan struct{}) {
	eventChan, closeChan := users.CreateUser(w, r)

	token := uuid.New().String()
	moderator := users.NewModerator(eventChan)
	room.moderators.Insert(token, moderator)

	// Initialize Moderator
	eventChan <- PastLogsEvent(room.logs)
	eventChan <- ModeratorPlayerControlsEvent(room.players.GetUsers())
	eventChan <- ModeratorStatusEvent(room.Status().String())
	return token, closeChan
}

func (room *Room) RemoveModerator(token string) {
	room.moderators.CloseAndDelete(token)
}

func (room *Room) RemovePlayer(token string) {
	room.players.CloseAndDelete(token)
	room.sendPlayerListUpdates()
}

/**
* Functions for pushing updates to the connected clients
* */
func (room *Room) log(message string) {
	timestamp := time.Now().UTC()

	// Add to list of log messages
	log := Log{message, timestamp}
	room.logs = append(room.logs, log)

	// Cap the size of the logs array at 100
	if len(room.logs) > 100 {
		room.logs = room.logs[1:]
	}

	logEvent := lib.FormatEventComponent("log", LogMessage(log))

	room.moderators.SendToAll(logEvent)
	room.players.SendToAll(logEvent)
}

func (room *Room) sendPlayerListUpdates() {
	users := room.players.GetUsers()

	room.players.SendToAll(PlayerListEvent(users))
	room.moderators.SendToAll(ModeratorPlayerControlsEvent(users))
}
