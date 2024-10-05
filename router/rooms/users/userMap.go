package users

import (
	"goodbuzz/lib/events"
	"sync"
)

type UserMap[T any] struct {
	sync.RWMutex
	channels map[chan string]T
}

func NewUserMap[T any]() *UserMap[T] {
	return &UserMap[T]{
		channels: make(map[chan string]T),
	}
}

func (um *UserMap[T]) SendToAll(messages ...string) {
	message := events.CombineEvents(messages...)
	um.RLock()
	defer um.RUnlock()

	for listener := range um.channels {
		listener <- message
	}
}

func (um *UserMap[T]) New(user T) chan string {
	eventChan := make(chan string)
	um.Lock()
	defer um.Unlock()

	um.channels[eventChan] = user
	return eventChan
}

func (um *UserMap[T]) Get(eventChan chan string) T {
	um.RLock()
	defer um.RUnlock()
	return um.channels[eventChan]
}

func (um *UserMap[T]) Delete(eventChan chan string) {
	um.Lock()
	defer um.Unlock()
	delete(um.channels, eventChan)
	close(eventChan)
}

func (um *UserMap[T]) NumUsers() int {
	return len(um.channels)
}
