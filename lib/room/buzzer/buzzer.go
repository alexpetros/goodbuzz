package buzzer

import (
	"goodbuzz/lib/logger"
	"time"

	"github.com/google/uuid"
)

const BUZZER_DELAY = 500 * time.Millisecond

type BuzzerStatus int

type BuzzerUpdate struct {
	Status     BuzzerStatus
	// This will be "" if there's no winner
	// Yes that's a little gross but go's enum system is... yeesh
	WinnerToken		 string
	ResetToken string
}

const (
	Unlocked   BuzzerStatus = 0
	Processing              = 1
	Won                     = 2
)

type Buzz struct {
	UserToken  string
	ResetToken string
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
	winnerToken := ""
	if len(buzzer.buzzes) > 0 {
		winnerToken = buzzer.buzzes[0].UserToken
	}

	return BuzzerUpdate {
		Status: buzzer.buzzerStatus,
		WinnerToken: winnerToken,
		ResetToken: buzzer.resetToken,
	}
}

func (buzzer *Buzzer) doUpdates(data *Buzz) {
	// This is the sign to reset
	if data.ResetToken == "" {
		buzzer.buzzerStatus = Unlocked
		buzzer.resetToken = uuid.NewString()
		buzzer.buzzes = make([]*Buzz, 0)
		return
	}

	// Ignore buzzes that don't match the reset userToken
	if data.ResetToken != buzzer.resetToken {
		logger.Info("reset userToken %s does not match room reset userToken %s", data.UserToken, buzzer.resetToken)
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
