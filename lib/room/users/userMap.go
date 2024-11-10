package users

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"net/http"
	"sync"
)

type user[T any] struct {
	data          T
	eventChan     chan string
	interruptChan chan struct{}
}

type UserMap[T any] struct {
	sync.RWMutex
	users map[string]user[T]
}

func NewUserMap[T any]() *UserMap[T] {
	return &UserMap[T]{
		users: make(map[string]user[T]),
	}
}

func (um *UserMap[T]) AddUser(w http.ResponseWriter, r *http.Request, userToken string, data T) <-chan struct{} {
	// Set the response header to indicate SSE content type
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	eventChan := make(chan string)
	interruptChan := make(chan struct{})
	closeChan := make(chan struct{})

	um.Lock()
	defer um.Unlock()
	newUser := user[T]{data, eventChan, interruptChan}
	um.users[userToken] = newUser

	// Remove the channel if the connection closes
	go func() {
		select {
		case <-r.Context().Done():
			um.RemoveUser(userToken)
			closeChan <- struct{}{}
		case <-newUser.interruptChan:
			// Send the close event, and then remove the the user
			newUser.eventChan <- lib.FormatEventString("close", "kicked")
			um.RemoveUser(userToken)
			closeChan <- struct{}{}
		}
	}()

	// Continuously send data to the client
	go func() {
		// Remove the channel if the loop exists (an exception, probably)
		defer func() {
			um.RemoveUser(userToken)
			closeChan <- struct{}{}
		}()

		rc := http.NewResponseController(w)
		for {
			data := <-eventChan

			// No need to handle the errors here, panics will just break the loop and remove the user
			fmt.Fprintf(w, data)
			rc.Flush()
		}
	}()

	return closeChan
}

func (um *UserMap[T]) SendToUser(userToken string, message string) {
	um.RLock()
	defer um.RUnlock()

	um.users[userToken].eventChan <- message
}

func (um *UserMap[T]) SendToAll(messages ...string) {
	message := lib.CombineEvents(messages...)
	um.RLock()
	defer um.RUnlock()

	for _, user := range um.users {
		user.eventChan <- message
	}
}

func (um *UserMap[T]) Run(userToken string, fn func(data T)) {
	um.RLock()
	defer um.RUnlock()

	user, ok := um.users[userToken]

	if ok {
		fn(user.data)
	}

}

func (um *UserMap[T]) RunAll(fn func(data T, eventChan chan string)) {
	um.RLock()
	defer um.RUnlock()
	for _, user := range um.users {
		fn(user.data, user.eventChan)
	}
}

func (um *UserMap[T]) RemoveUser(token string) {
	um.Lock()
	defer um.Unlock()
	delete(um.users, token)
}

func (um *UserMap[T]) NumUsers() int {
	um.RLock()
	defer um.RUnlock()
	return len(um.users)
}

func (um *UserMap[T]) KickUser(userToken string) {
	um.RLock()
	defer um.RUnlock()
	user, ok := um.users[userToken]
	if ok {
		user.interruptChan <- struct{}{}
	} else {
		logger.Info("Attempted to kick user %s who was not present", userToken)
	}
}

func (um *UserMap[T]) KickAll() {
	um.RLock()
	defer um.RUnlock()

	for _, user := range um.users {
		user.interruptChan <- struct{}{}
	}
}

func (um *UserMap[T]) HasUser(userToken string) bool {
	_, ok := um.Get(userToken)
	return ok
}

func (um *UserMap[T]) Get(userToken string) (T, bool) {
	um.RLock()
	defer um.RUnlock()
	user, ok := um.users[userToken]
	return user.data, ok
}

func (um *UserMap[T]) GetAll() []T {
	um.RLock()
	defer um.RUnlock()

	items := make([]T, 0)
	for _, user := range um.users {
		items = append(items, user.data)
	}

	return items
}
