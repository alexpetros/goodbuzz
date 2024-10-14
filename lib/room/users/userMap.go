package users

import (
	"errors"
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"net/http"
	"sync"
)

type User interface {
	Channel() chan string
}

func CreateUser(w http.ResponseWriter, r *http.Request) (chan string, chan struct{}) {
	notify := r.Context().Done()
	eventChan := make(chan string)
	closeChan := make(chan struct{})

	go func() {
		<-notify
		closeChan <- struct{}{}
	}()

	// Continuously send data to the client
	go func() {
		for {
			data := <-eventChan
			// Upon receiving the zero value (""), the channel is closed, so break the loop
			if data == "" {
				break
			}

			//logger.Debug("Sending data to moderator in room %d:\n%s", room.Id(), data)
			_, err2 := fmt.Fprintf(w, data)
			if err2 != nil {
				//logger.Error("Failed to send data to moderator in room %d:\n%s", room.Id(), data)
			}

			if w != nil && w.(http.Flusher) != nil {
				w.(http.Flusher).Flush()
			} else {
				logger.Warn("write to socket after connection closed")
			}
		}
	}()

	return eventChan, closeChan
}

type UserMap[T User] struct {
	sync.RWMutex
	users map[string]T
}

func NewUserMap[T User]() *UserMap[T] {
	return &UserMap[T]{
		users: make(map[string]T),
	}
}

func (um *UserMap[T]) SendToAll(messages ...string) {
	message := lib.CombineEvents(messages...)
	um.RLock()
	defer um.RUnlock()

	for _, user := range um.users {
		user.Channel() <- message
	}
}

func (um *UserMap[T]) Insert(token string, user T) {
	um.Lock()
	defer um.Unlock()
	um.users[token] = user
}

func (um *UserMap[T]) Get(token string) (T, error) {
	um.RLock()
	defer um.RUnlock()
	user, ok := um.users[token]

	if ok {
		return user, nil
	} else {
		return user, errors.New("Resource was not found")
	}
}

func (um *UserMap[T]) Run(fn func(T)) {
	um.RLock()
	defer um.RUnlock()
	for _, user := range um.users {
		fn(user)
	}
}

func (um *UserMap[T]) CloseAndDelete(token string) {
	um.Lock()
	defer um.Unlock()
	user := um.users[token]
	close(user.Channel())
	delete(um.users, token)
}

func (um *UserMap[T]) NumUsers() int {
	um.RLock()
	defer um.RUnlock()
	return len(um.users)
}

func (um *UserMap[T]) GetUsers() []T {
	um.RLock()
	defer um.RUnlock()

	res := make([]T, 0, len(um.users))
	for _, value := range um.users {
		res = append(res, value)
	}

	return res
}
