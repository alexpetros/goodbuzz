package room

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/events"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room/buzzer"
	"goodbuzz/lib/room/users"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Room struct {
	roomId     int64
	name       string
	logs       []events.Log
	locksCache *LocksCache
	buzzer     *buzzer.Buzzer
	players    *users.UserMap[*users.Player]
	moderators *users.UserMap[struct{}]
}

func (roomMap *RoomMap) newRoom(roomId int64, name string) *Room {
	room := Room{
		roomId:     roomId,
		name:       name,
		logs:       make([]events.Log, 0),
		locksCache:	NewLocksCache(),
		players:    users.NewUserMap[*users.Player](),
		moderators: users.NewUserMap[struct{}](),
	}
	room.buzzer = buzzer.NewBuzzer(room.sendBuzzerUpdates)
	return &room
}

func (room *Room) sendBuzzerUpdates(buzzerUpdate buzzer.BuzzerUpdate) {
	room.UpdateBuzzers(buzzerUpdate)
	room.moderators.SendToAll(room.currentModeratorBuzzer(buzzerUpdate))
	if buzzerUpdate.Status == buzzer.Won {
		winner := buzzerUpdate.Buzzes[0]
		player, ok := room.players.Get(winner.UserToken)
		if !ok {
			room.log(fmt.Sprintf("(disconnected player) won the buzz!"))
		}

		room.log(fmt.Sprintf("%s won the buzz!", player.Name))
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
	return fmt.Sprintf("/rooms/%d/player", room.roomId)
}

func (room *Room) ModeratorUrl() string {
	return fmt.Sprintf("/rooms/%d/moderator", room.roomId)
}

func (room *Room) Status() buzzer.BuzzerStatus {
	return room.buzzer.GetUpdate().Status
}

func (room *Room) LockPlayer(userToken string) {
	room.players.Run(userToken, func(player *users.Player) {
		player.Lock()
	})
}

func (room *Room) UnlockPlayer(userToken string) {
	logger.Debug("Unlocking player %s", userToken)
	room.players.Run(userToken, func(player *users.Player) {
		player.Unlock()
	})

	room.buzzer.SendUpdates()
	room.sendPlayerListUpdates()
}

func (room *Room) UpdateBuzzers(buzzerUpdate buzzer.BuzzerUpdate) {
	updateFunc := func(player *users.Player, eventChan chan string) {
		if player.IsLocked {
			eventChan <- events.LockedBuzzerEvent()
		} else {
			eventChan <- currentPlayerBuzzer(player, buzzerUpdate)
		}
	}
	room.players.RunAll(updateFunc)
}

func (room *Room) SetPlayerName(userToken string, name string) {
	logger.Debug("Setting %s name to %s", userToken, name)

	room.players.Run(userToken, func(player *users.Player) {
		player.SetName(name)
	})

	room.sendPlayerListUpdates()
}

func (room *Room) BuzzRoom(userToken string, resetToken string) {
	room.buzzer.Buzz(userToken, resetToken)

	player, ok := room.players.Get(userToken)
	if !ok {
		logger.Info("unknown player %s buzzed", userToken)
		return
	}
	logMessage := fmt.Sprintf("Player %v buzzed room %v", player.Name, room.Id())
	logger.Debug(logMessage)
	room.log(logMessage)
}

func (room *Room) ResetAll() {
	logger.Debug("Resetting all buzzers")

	room.locksCache.ResetAll()
	room.players.RunAll(func(player *users.Player, eventChan chan string) {
		player.Unlock()
	})

	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for everyone")
}

func (room *Room) ResetSome() {
	logger.Debug("Resetting some buzzers")
	buzzerUpdate := room.buzzer.GetUpdate()

	if buzzerUpdate.Status == buzzer.Unlocked {
		logger.Info("room %s was reset with no active buzzes", room.name)
	} else if buzzerUpdate.Status == buzzer.Processing {
		logger.Info("room %s was reset during processing", room.name)
	} else if buzzerUpdate.Status == buzzer.Won {
		winner := buzzerUpdate.Buzzes[0]
		logger.Info("Locking player with userToken %s", winner.UserToken)
		room.locksCache.LockPlayer(winner.UserToken)
		room.LockPlayer(winner.UserToken)
	}

	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for some players")
}

func currentPlayerBuzzer(player *users.Player, buzzerUpdate buzzer.BuzzerUpdate) string {
	if buzzerUpdate.Status != buzzer.Unlocked || player.IsLocked {
		return events.LockedBuzzerEvent()
	}

	return events.ReadyBuzzerEvent(buzzerUpdate.ResetToken)
}

func (room *Room) currentModeratorBuzzer(buzzerUpdate buzzer.BuzzerUpdate) string {
	if buzzerUpdate.Status == buzzer.Won {
		winner := buzzerUpdate.Buzzes[0]
		player, ok := room.players.Get(winner.UserToken)

		var message string
		if !ok {
			message = fmt.Sprintf("Locked by (disconnected player)")
		} else {
			message = fmt.Sprintf("Locked by %s", player.Name)
		}

		return events.ModeratorStatusEvent(message)
	} else if buzzerUpdate.Status == buzzer.Processing {
		return events.ModeratorStatusEvent("Processing...")
	} else {
		return events.ModeratorStatusEvent("Unlocked")
	}
}

func (room *Room) AttachPlayer(w http.ResponseWriter, r *http.Request, userToken string) {
	nameCookie, err := r.Cookie("name")
	var name string
	if err != nil {
		name = "New Player"
	} else {
		name = nameCookie.Value
	}

	isLocked := room.locksCache.IsLocked(userToken)
	player := users.NewPlayer(name, userToken, isLocked)
	closeChan := room.players.AddUser(w, r, userToken, player)

	// Initialize Player
	room.sendPlayerListUpdates()
	room.players.SendToPlayer(userToken, events.PastLogsEvent(room.logs))
	room.players.SendToPlayer(userToken, currentPlayerBuzzer(player, room.buzzer.GetUpdate()))

	// Wait for the channel to close, and then send everyone else the disconnect update
	<-closeChan
	room.sendPlayerListUpdates()
}

func (room *Room) AttachModerator(w http.ResponseWriter, r *http.Request, userToken string) {
	closeChan := room.moderators.AddUser(w, r, userToken, struct{}{})

	// Initialize Moderator
	room.moderators.SendToPlayer(userToken, events.PastLogsEvent(room.logs))
	room.moderators.SendToPlayer(userToken, events.ModeratorPlayerControlsEvent(room.players.GetAll()))
	room.moderators.SendToPlayer(userToken, room.currentModeratorBuzzer(room.buzzer.GetUpdate()))

	// Wait for the channel to close, and then send everyone else the disconnect update
	<-closeChan
	room.sendPlayerListUpdates()
}

/**
* Functions for pushing updates to the connected clients
* */
func (room *Room) log(message string) {
	timestamp := time.Now().UTC()

	// Add to list of log messages
	log := events.Log{Message: message, Timestamp: timestamp}
	room.logs = append(room.logs, log)

	// Cap the size of the logs array at 100
	if len(room.logs) > 100 {
		room.logs = room.logs[1:]
	}

	logEvent := lib.FormatEventComponent("log", events.LogMessage(log))

	room.moderators.SendToAll(logEvent)
	room.players.SendToAll(logEvent)
}

func (room *Room) sendPlayerListUpdates() {
	players := room.players.GetAll()
	slices.SortFunc(players, func(a, b *users.Player) int {
		return strings.Compare(a.Name, b.Name)
	})

	room.players.RunAll(func(player *users.Player, eventChan chan string) {
		eventChan <- events.PlayerListEvent(players, player)
	})

	room.moderators.SendToAll(events.ModeratorPlayerControlsEvent(players))
}
