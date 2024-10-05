package users

import (
	"goodbuzz/lib"
	"sync"
)

type User interface {
	Channel() chan string
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

func (um *UserMap[T]) Get(token string) T {
	um.RLock()
	defer um.RUnlock()
	return um.users[token]
}

func (um *UserMap[T]) CloseAndDelete(token string) {
	um.Lock()
	defer um.Unlock()
	user := um.users[token]
	close(user.Channel())
	delete(um.users, token)
}

func (um *UserMap[T]) NumUsers() int {
	return len(um.users)
}
