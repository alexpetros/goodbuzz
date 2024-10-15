package room

import (
	"goodbuzz/lib/logger"
	"time"

	"github.com/google/uuid"
)

const BUZZER_DELAY = 500 * time.Millisecond

type BuzzerStatus int

type BuzzerUpdate struct {
	status     BuzzerStatus
	resetToken string
	buzzes     []Buzz
}

const (
	Unlocked   BuzzerStatus = 0
	Processing              = 1
	Won                     = 2
)

type Buzz struct {
	token      string
	resetToken string
}

type Buzzer struct {
	buzzes         []*Buzz
	resetToken     string
	readChannel    chan struct{}
	writeChannel   chan *Buzz
	updateChannel  chan BuzzerUpdate
	updateCallback func(BuzzerUpdate)
	buzzerStatus   BuzzerStatus
}

func NewBuzzer(updateCallback func(BuzzerUpdate)) *Buzzer {
	buzzer := &Buzzer{
		buzzes:         make([]*Buzz, 0),
		resetToken:     uuid.NewString(),
		readChannel:    make(chan struct{}),
		writeChannel:   make(chan *Buzz),
		updateChannel:  make(chan BuzzerUpdate),
		updateCallback: updateCallback,
		buzzerStatus:   Unlocked,
	}

	go func() {
		for {
			// This is read/write locking code
			// The buzzer status will only be retrieved when no one else is editing it
			// I don't really think this is much simpler than mutexes, but it's fine I guess
			select {
			case <-buzzer.readChannel:
				buzzer.updateChannel <- buzzer.makeUpdateSnapshot()
			case data := <-buzzer.writeChannel:
				buzzer.doUpdates(data)
			}
		}
	}()

	return buzzer
}

func (buzzer *Buzzer) makeUpdateSnapshot() BuzzerUpdate {
	buzzes := make([]Buzz, len(buzzer.buzzes))
	for i, buzz := range buzzer.buzzes {
		buzzes[i] = Buzz{buzz.token, buzz.resetToken}
	}

	return BuzzerUpdate{buzzer.buzzerStatus, buzzer.resetToken, buzzes}
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

	// If the buzzer isn't locked yet, append the buzz
	if buzzer.buzzerStatus != Won {
		logger.Debug("Buzzer isn't locked yet, adding new buzz data")
		buzzer.buzzes = append(buzzer.buzzes, data)
	}

	// If the buzzer isn't on a timer yet, start the processing timer
	if buzzer.buzzerStatus == Unlocked {
		logger.Debug("Buzzer isn't processing yet, starting processing")
		go buzzer.StartProcessing()
	}

}

func (buzzer *Buzzer) StartProcessing() {
	buzzer.buzzerStatus = Processing
	buzzer.SendUpdates()

	time.Sleep(BUZZER_DELAY)

	logger.Debug("Processing finished, locking buzzer")
	logger.Debug("Buzzes: %v", buzzer.buzzes)
	buzzer.buzzerStatus = Won
	buzzer.SendUpdates()
}

func (buzzer *Buzzer) SendUpdates() {
	buzzer.readChannel <- struct{}{}
	buzzerUpdate := <-buzzer.updateChannel
	buzzer.updateCallback(buzzerUpdate)
}

func (buzzer *Buzzer) GetUpdate() BuzzerUpdate {
	buzzer.readChannel <- struct{}{}
	return <-buzzer.updateChannel
}

func (buzzer *Buzzer) Buzz(token string, resetToken string) {
	buzzer.writeChannel <- &Buzz{token, resetToken}
	buzzer.SendUpdates()
}

func (buzzer *Buzzer) Reset() {
	buzzer.writeChannel <- &Buzz{"", ""}
	buzzer.SendUpdates()
}
