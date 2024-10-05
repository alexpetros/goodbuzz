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

func (cm *userMap[T]) sendToAll(messages ...string) {
	message := CombineEvents(messages...)
	cm.RLock()
	defer cm.RUnlock()

	for listener := range cm.channels {
		listener <- message
	}
}

func (cm *userMap[T]) new(user T) chan string {
	eventChan := make(chan string)
	cm.Lock()
	defer cm.Unlock()

	cm.channels[eventChan] = user
	return eventChan
}

func (cm *userMap[T]) get(eventChan chan string) T {
	cm.RLock()
	defer cm.RUnlock()
	return cm.channels[eventChan]
}

func (cm *userMap[T]) delete(eventChan chan string) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.channels, eventChan)
	close(eventChan)
}
