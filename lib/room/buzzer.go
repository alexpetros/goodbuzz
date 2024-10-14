package room

import (
	"goodbuzz/lib/logger"

	"github.com/google/uuid"
)

type BuzzerStatus int

type BuzzerUpdate struct {
	status BuzzerStatus
	resetToken string
	winner string
}

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

type Buzz struct {
	token      string
	resetToken string
}

type Buzzer struct {
	buzzes       []*Buzz
	resetToken   string
	readChannel  	chan struct{}
	writeChannel  chan *Buzz
	updateChannel   chan BuzzerUpdate
	updateCallback func (BuzzerUpdate)
	buzzerStatus BuzzerStatus
}

func NewBuzzer(updateCallback func(BuzzerUpdate)) *Buzzer {
	buzzer := &Buzzer {
		buzzes:       make([]*Buzz, 0),
		resetToken: 	uuid.NewString(),
		readChannel:  	make(chan struct{}),
		writeChannel: 	make(chan *Buzz),
		updateChannel:  		make(chan BuzzerUpdate),
		updateCallback: updateCallback,
		buzzerStatus: Unlocked,
	}

	go func() {
		for {
			// This is read/write locking code
			// The buzzer status will only be retrieved when no one else is editing it
			// I don't really think this is much simpler than mutexes, but it's fine I guess
			select {
			case <- buzzer.readChannel:
				buzzer.updateChannel <- buzzer.makeUpdateSnapshot()
			case data := <- buzzer.writeChannel:
				buzzer.doUpdates(data)
			}
		}
	}()

	return buzzer
}

func (buzzer *Buzzer) makeUpdateSnapshot() BuzzerUpdate {
	winner := ""
	if len(buzzer.buzzes) > 0 {
		winner = buzzer.buzzes[0].token
	}

	return BuzzerUpdate { buzzer.buzzerStatus, buzzer.resetToken, winner }
}

func (buzzer *Buzzer) doUpdates(data *Buzz) {
	// This is the sign to reset
	if data.resetToken == "" {
		buzzer.buzzerStatus = Unlocked
		buzzer.resetToken = uuid.NewString()
		buzzer.buzzes = make([]*Buzz, 0)
		return
	}

	// Ignore buzzes that don't match the reset token
	if data.resetToken != buzzer.resetToken {
		logger.Info("reset token %s does not match room reset token %s", data.resetToken, buzzer.resetToken)
		return
	}

	// Append buzzes that match all these conditions
	buzzer.buzzerStatus = Locked
	buzzer.buzzes = append(buzzer.buzzes, data)
}

func (buzzer *Buzzer) SendUpdates() {
	buzzer.readChannel <- struct{}{}
	buzzerUpdate := <- buzzer.updateChannel
	buzzer.updateCallback(buzzerUpdate)
}

func (buzzer *Buzzer) GetUpdate() BuzzerUpdate {
	buzzer.readChannel <- struct{}{}
	return <- buzzer.updateChannel
}

func (buzzer *Buzzer) Buzz(token string, resetToken string) {
	buzzer.writeChannel <- &Buzz{token, resetToken}
	buzzer.SendUpdates()
}

func (buzzer *Buzzer) Reset() {
	buzzer.writeChannel <- &Buzz{"", ""}
	buzzer.SendUpdates()
}
