package room

import (
	"sync"
)

type LocksCache struct {
	sync.RWMutex
	cache map[string]struct{}
}

func NewLocksCache() *LocksCache {
	return &LocksCache {
		cache: make(map[string]struct{}),
	}
}

func (lc *LocksCache) IsLocked(userToken string) bool {
	lc.RLock()
	defer lc.RUnlock()

	_, ok := lc.cache[userToken]
	return ok
}

func (lc *LocksCache) LockPlayer(userToken string) {
	lc.Lock()
	defer lc.Unlock()
	lc.cache[userToken] = struct{}{}
}

func (lc *LocksCache) UnlockPlayer(userToken string) {
	lc.Lock()
	defer lc.Unlock()
	delete(lc.cache, userToken)
}

func (lc *LocksCache) ResetAll() {
	lc.Lock()
	defer lc.Unlock()
	lc.cache = make(map[string]struct{})
}

