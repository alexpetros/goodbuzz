package rooms

import (
	"sync"
)

type userMap[T any] struct {
	sync.RWMutex
	channels map[chan string]T
}

func newUserMap[T any]() *userMap[T] {
	return &userMap[T]{
		channels: make(map[chan string]T),
	}
}

func (um *userMap[T]) sendToAll(messages ...string) {
	message := CombineEvents(messages...)
	um.RLock()
	defer um.RUnlock()

	for listener := range um.channels {
		listener <- message
	}
}

func (um *userMap[T]) new(user T) chan string {
	eventChan := make(chan string)
	um.Lock()
	defer um.Unlock()

	um.channels[eventChan] = user
	return eventChan
}

func (um *userMap[T]) get(eventChan chan string) T {
	um.RLock()
	defer um.RUnlock()
	return um.channels[eventChan]
}

func (um *userMap[T]) delete(eventChan chan string) {
	um.Lock()
	defer um.Unlock()
	delete(um.channels, eventChan)
	close(eventChan)
}
