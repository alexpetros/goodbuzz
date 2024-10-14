package room

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room/users"
	"net/http"
	"slices"
	"strings"
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
	players := room.players.GetUsers()

	for _, player := range players {
		if player.IsLocked() {
			player.Channel() <- LockedBuzzerEvent()
		} else {
			player.Channel() <- currentPlayerBuzzer(buzzerUpdate)
		}
	}

	room.moderators.SendToAll(room.currentModeratorBuzzer(buzzerUpdate))
	if buzzerUpdate.status == Won {
		winner := buzzerUpdate.buzzes[0]
		player, err := room.players.Get(winner.token)
		if err != nil {
			room.log(fmt.Sprintf("(disconnected player) won the buzz!", player.Name()))
		} else {
			room.log(fmt.Sprintf("%s won the buzz!", player.Name()))
		}
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
	return room.buzzer.GetUpdate().status
}

func (room *Room) lockPlayer(token string) {
	player, err := room.players.Get(token)
	if err != nil {
		logger.Info("Attempted to lock player with token %s, who has disconnected", token)
	} else {
		player.Lock()
	}
}

func (room *Room) unlockAll() {
	room.players.Run(func(player *users.Player) {
		player.Unlock()
	})
}

func (room *Room) UnlockPlayer(token string) {
	logger.Debug("Unlocking player %s", token)
	player, err := room.players.Get(token)

	if err != nil {
		logger.Info("Attempted to unlock player %s who has since disconnected", token)
		return
	}

	player.Unlock()
	room.buzzer.SendUpdates()
	room.sendPlayerListUpdates()
}

func (room *Room) SetPlayerName(token string, name string) {
	logger.Debug("Setting %s name to %s", token, name)
	player, err := room.players.Get(token)

	if err != nil {
		logger.Info("Attempted to set name for player %s who has since disconnected", token)
		return
	}

	player.SetName(name)
	room.sendPlayerListUpdates()
}

func (room *Room) BuzzRoom(token string, resetToken string) {
	player, err := room.players.Get(token)
	if err != nil {
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

	buzzerUpdate := room.buzzer.GetUpdate()

	if buzzerUpdate.status == Unlocked {
		logger.Info("room %s was reset with no active buzzes", room.name)
	} else if buzzerUpdate.status == Processing {
		logger.Info("room %s was reset during processing", room.name)
	} else if buzzerUpdate.status == Won {
		winner := buzzerUpdate.buzzes[0]
		logger.Info("Locking player with token %s", winner.token)
		room.lockPlayer(winner.token)
	}

	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for some players")
}

func currentPlayerBuzzer(buzzerUpdate BuzzerUpdate) string {
	if buzzerUpdate.status != Unlocked {
		return LockedBuzzerEvent()
	}

	return ReadyBuzzerEvent(buzzerUpdate.resetToken)
}

func (room *Room) currentModeratorBuzzer(buzzerUpdate BuzzerUpdate) string {
	if buzzerUpdate.status == Won {
		winner := buzzerUpdate.buzzes[0]
		player, err := room.players.Get(winner.token)

		var name string
		if err != nil {
			name = "(disconnected player)"
		} else {
			name = player.Name()
		}

		message := fmt.Sprintf("Locked by %s", name)
		return ModeratorStatusEvent(message)
	} else if buzzerUpdate.status == Processing {
		return ModeratorStatusEvent("Processing...")
	} else {
		return ModeratorStatusEvent("Unlocked")
	}
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
	eventChan <- room.currentModeratorBuzzer(room.buzzer.GetUpdate())
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
	players := room.players.GetUsers()
	slices.SortFunc(players, func(a, b *users.Player) int {
		return strings.Compare(a.Name(), b.Name())
	})

	for _, player := range players {
		player.Channel() <- PlayerListEvent(players, player)
	}

	room.moderators.SendToAll(ModeratorPlayerControlsEvent(players))
}
