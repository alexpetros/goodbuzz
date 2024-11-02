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
	Id      int64
	Name        string
	Description string
	logs        []events.Log
	locksCache  *LocksCache
	buzzer      *buzzer.Buzzer
	players     *users.UserMap[*users.Player]
	moderators  *users.UserMap[struct{}]
}

type roomUpdate struct {
	buzzerStatus buzzer.BuzzerStatus
	// nilable because there will be no winner if the status isn't "Won"
	winner     *users.Player
	resetToken string
}

func (roomMap *RoomMap) newRoom(roomId int64, name string, description string) *Room {
	room := Room{
		Id:      roomId,
		Name:        name,
		Description: description,
		logs:        make([]events.Log, 0),
		locksCache:  NewLocksCache(),
		players:     users.NewUserMap[*users.Player](),
		moderators:  users.NewUserMap[struct{}](),
	}
	room.buzzer = buzzer.NewBuzzer(room.sendBuzzerUpdates)
	return &room
}

func (room *Room) convertUpdate(buzzerUpdate buzzer.BuzzerUpdate) roomUpdate {
	var winner *users.Player
	if buzzerUpdate.Status == buzzer.Won {
		winnerToken := buzzerUpdate.WinnerToken
		player, ok := room.players.Get(winnerToken)

		// Ensures we never have a nil winner, if there's a winner
		if !ok {
			winner = users.NewPlayer("(disconnected player)", buzzerUpdate.WinnerToken, false)
		} else {
			winner = player
		}
	}

	update := roomUpdate{buzzerUpdate.Status, winner, buzzerUpdate.ResetToken}
	return update
}

func (room *Room) getUpdate() roomUpdate {
	buzzerUpdate := room.buzzer.GetUpdate()
	return room.convertUpdate(buzzerUpdate)
}

func (room *Room) sendBuzzerUpdates(buzzerUpdate buzzer.BuzzerUpdate) {
	update := room.convertUpdate(buzzerUpdate)

	// Update all the player buzzers
	updateFunc := func(player *users.Player, eventChan chan string) {
		if player.IsLocked {
			eventChan <- events.LockedOutBuzzerEvent()
		} else {
			eventChan <- currentPlayerBuzzer(player, update)
		}
	}
	room.players.RunAll(updateFunc)

	// Update all the moderator statuses
	room.moderators.SendToAll(room.currentModeratorBuzzer(update))

	// Log update if there's a winner
	if update.buzzerStatus == buzzer.Won {
		room.log(fmt.Sprintf("%s won the buzz!", update.winner.Name))
	}
}

func (room *Room) SetName(name string) {
	room.Name = name
}

func (room *Room) SetDescription(description string) {
	room.Description = description
}

func (room *Room) Url() string {
	return fmt.Sprintf("/rooms/%d", room.Id)
}

func (room *Room) EditUrl() string {
	return fmt.Sprintf("/rooms/%d/edit", room.Id)
}

func (room *Room) PlayerUrl() string {
	return fmt.Sprintf("/rooms/%d/player", room.Id)
}

func (room *Room) ModeratorUrl() string {
	return fmt.Sprintf("/rooms/%d/moderator", room.Id)
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

func (room *Room) KickPlayer(userToken string) {
	logger.Debug("Kicking player %s", userToken)
	room.players.KickUser(userToken)
}

func (room *Room) KickAll() {
	logger.Debug("Kicking everyone out room %s (id: %d)", room.Name, room.Id)
	room.players.KickAll()
	room.moderators.KickAll()
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
	logMessage := fmt.Sprintf("Player %v buzzed room %v", player.Name, room.Id)
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
	update := room.getUpdate()

	if update.buzzerStatus == buzzer.Unlocked {
		logger.Info("room %s was reset with no active buzzes", room.Name)
	} else if update.buzzerStatus == buzzer.Processing {
		logger.Info("room %s was reset during processing", room.Name)
	} else if update.buzzerStatus == buzzer.Won {
		logger.Info("Locking player with userToken %s", update.winner.Token)
		room.locksCache.LockPlayer(update.winner.Token)
		room.LockPlayer(update.winner.Token)
	}

	room.buzzer.Reset()
	room.sendPlayerListUpdates()
	room.log("Buzzer unlocked for some players")
}

func currentPlayerBuzzer(player *users.Player, update roomUpdate) string {

	if player.IsLocked {
		return events.LockedOutBuzzerEvent()
	}

	if update.buzzerStatus == buzzer.Won {
		if update.winner.Token == player.Token {
			return events.YouWonBuzzerEvent()
		} else {
			return events.OtherPlayerWonBuzzerEvent(update.winner)
		}
	}

	if update.buzzerStatus == buzzer.Processing {
		return events.ProcessingBuzzerEvent()
	}

	return events.ReadyBuzzerEvent(update.resetToken)
}

func (room *Room) currentModeratorBuzzer(update roomUpdate) string {
	if update.buzzerStatus == buzzer.Won {
		return events.LockedStatusEvent(update.winner)
	} else if update.buzzerStatus == buzzer.Processing {
		return events.ProcessingStatusEvent()
	} else {
		return events.UnlockedStatusEvent()
	}
}

func (room *Room) IsPlayerAlreadyConnected(userToken string) bool {
	return room.players.HasUser(userToken)
}

func (room *Room) AttachPlayer(w http.ResponseWriter, r *http.Request, userToken string, name string) {
	isLocked := room.locksCache.IsLocked(userToken)
	player := users.NewPlayer(name, userToken, isLocked)
	closeChan := room.players.AddUser(w, r, userToken, player)

	// Initialize Player
	room.sendPlayerListUpdates()
	room.players.SendToUser(userToken, events.PastLogsEvent(room.logs))
	room.players.SendToUser(userToken, currentPlayerBuzzer(player, room.getUpdate()))

	// Wait for the channel to close, and then send everyone else the disconnect update
	<-closeChan
	room.sendPlayerListUpdates()
	// Wait two seconds to give the connection time to close gracefully, if necessary
	time.Sleep(2 * time.Second)
}

func (room *Room) AttachModerator(w http.ResponseWriter, r *http.Request, userToken string) {
	closeChan := room.moderators.AddUser(w, r, userToken, struct{}{})

	// Initialize Moderator
	room.moderators.SendToUser(userToken, events.PastLogsEvent(room.logs))
	room.moderators.SendToUser(userToken, events.ModeratorPlayerControlsEvent(room.Id, room.players.GetAll()))
	room.moderators.SendToUser(userToken, room.currentModeratorBuzzer(room.getUpdate()))

	// Wait for the channel to close, and then send everyone else the disconnect update
	<-closeChan
	room.sendPlayerListUpdates()
	time.Sleep(2 * time.Second)
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

	room.moderators.SendToAll(events.ModeratorPlayerControlsEvent(room.Id, players))
}
