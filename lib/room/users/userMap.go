package users

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"net/http"
	"sync"
)

type item [T any] struct {
	data T
	eventChan chan string
}

type UserMap [T any] struct {
	sync.RWMutex
	users map[string]item[T]
}

func NewUserMap[T any]() *UserMap[T] {
	return &UserMap[T]{
		users: make(map[string]item[T]),
	}
}

func (um *UserMap[T]) AddUser(w http.ResponseWriter, r *http.Request, userToken string, data T) <- chan struct {} {
	// Set the response header to indicate SSE content type
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	eventChan := make(chan string)
	closeChan := make(chan struct{})

	um.Lock()
	um.users[userToken] = item[T] { data, eventChan }
	um.Unlock()

	// Remove the channel if the connection closes
	go func() {
		<- r.Context().Done()
		um.RemoveUser(userToken)
		closeChan <- struct{}{}
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

			_, err2 := fmt.Fprintf(w, data)
			if err2 != nil {
				logger.Error("error writing data:\n%v", err2)
			}

			err := rc.Flush()
			if err != nil {
				logger.Error("error flushing writer:\n%v", err)
			}
		}
	}()

	return closeChan
}

func (um *UserMap[T]) SendToPlayer(userToken string, message string) {
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

	item := um.users[userToken].data
	fn(item)
}

func (um *UserMap[T]) RunAll(fn func(data T, eventChan chan string)) {
	um.RLock()
	defer um.RUnlock()
	for _, item := range um.users {
		fn(item.data, item.eventChan)
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
	for _, item := range um.users {
		items = append(items, item.data)
	}

	return items
}
