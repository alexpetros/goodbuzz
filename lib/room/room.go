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
	resetToken   string
	logs         []Log
	buzzes       []string
	buzzerStatus BuzzerStatus
	players      *users.UserMap[*users.Player]
	moderators   *users.UserMap[*users.Moderator]
}

func (roomMap *RoomMap) newRoom(roomId int64, name string) *Room {
	return &Room{
		roomId:       roomId,
		name:         name,
		resetToken:   uuid.NewString(),
		logs:					make([]Log, 0),
		buzzes:       make([]string, 0),
		buzzerStatus: Unlocked,
		players:      users.NewUserMap[*users.Player](),
		moderators:   users.NewUserMap[*users.Moderator](),
	}
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
	return room.buzzerStatus
}

func (room *Room) StatusString() string {
	return room.buzzerStatus.String()
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
	room.sendBuzzerUpdates()
	room.sendPlayerListUpdates()
}

func (room *Room) SetPlayerName(token string, name string) {
	logger.Debug("Setting %s name to %s", token, name)
	player := room.players.Get(token)
	player.SetName(name)
	room.sendPlayerListUpdates()
}

func (room *Room) BuzzRoom(token string, resetToken string) {
	logger.Debug("player %s buzzed with resetToken %s", token, resetToken)

	if resetToken != room.resetToken {
		logger.Info("reset token %s does not match room reset token %s", token, resetToken)
		return
	}

	player := room.players.Get(token)
	if player == nil {
		logger.Error("nil player returned for token %v")
		return
	}

	room.buzzerStatus = Locked
	room.buzzes = append(room.buzzes, token)
	logMessage := fmt.Sprintf("%s puzzed", player.Name())

	room.sendBuzzerUpdates()
	room.log(logMessage)
}

func (room *Room) ResetAll() {
	logger.Debug("Resetting all buzzers")
	room.buzzerStatus = Unlocked
	room.buzzes = make([]string, 0)
	room.unlockAll()

	room.resetToken = uuid.NewString()
	room.sendBuzzerUpdates()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for everyone")
}

func (room *Room) ResetSome() {
	logger.Debug("Resetting some buzzers")
	room.buzzerStatus = Unlocked

	if len(room.buzzes) < 1 {
		logger.Info("room %s was reset with no active buzzes", room.name)
	} else {
		token := room.buzzes[0]
		room.lockPlayer(token)
	}

	room.buzzes = make([]string, 0)

	room.resetToken = uuid.NewString()
	room.sendBuzzerUpdates()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for some players")
}

func (room *Room) CurrentBuzzerEvent() string {
	var buzzer string

	status := room.buzzerStatus
	if status == Unlocked {
		buzzer = ReadyBuzzerEvent(room.resetToken)
	} else if status == Waiting {
		buzzer = WaitingBuzzerEvent()
	} else if status == Locked {
		buzzer = LockedBuzzerEvent()
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
	player := users.NewPlayer(name, token, eventChan)

	room.players.Insert(token, player)

	// Initialize Player
	room.sendPlayerListUpdates()
	player.Channel() <- TokenEvent(token)
	player.Channel() <- PastLogsEvent(room.logs)
	player.Channel() <- room.CurrentBuzzerEvent()

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
	log := Log { message, timestamp }
	room.logs = append(room.logs, log)

	// Cap the size of the logs array at 100
	if len(room.logs) > 100 {
		room.logs = room.logs[1:]
	}

	logEvent := lib.FormatEventComponent("log", LogMessage(log))

	room.moderators.SendToAll(logEvent)
	room.players.SendToAll(logEvent)
}

func (room *Room) sendBuzzerUpdates() {
	for _, player := range room.players.GetUsers() {
		if player.IsLocked() {
			player.Channel() <- LockedBuzzerEvent()
		} else {
			player.Channel() <- room.CurrentBuzzerEvent()
		}
	}

	statusString := room.buzzerStatus.String()
	room.moderators.SendToAll(ModeratorStatusEvent(statusString))
}

func (room *Room) sendPlayerListUpdates() {
	users := room.players.GetUsers()

	room.players.SendToAll(PlayerListEvent(users))
	room.moderators.SendToAll(ModeratorPlayerControlsEvent(users))
}
