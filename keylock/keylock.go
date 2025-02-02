//go:build !solution

package keylock

import (
	"sort"
	"sync"

	"github.com/puzpuzpuz/xsync/v3"
)

type KeyLock struct {
	lock  *xsync.MapOf[string, chan struct{}]
	check sync.RWMutex
}

func (l *KeyLock) create(key string) {
	l.check.Lock()
	if _, ok := l.lock.Load(key); !ok {
		l.lock.Store(key, make(chan struct{}, 1))
	}
	l.check.Unlock()
}

func New() *KeyLock {
	return &KeyLock{lock: xsync.NewMapOf[string, chan struct{}]()}
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	keys = l.sort(keys)

	var last int
	unlock = func() {
		var ch chan struct{}
		for i := range last {
			ch, _ = l.lock.Load(keys[i])
			<-ch
		}
	}

	var ch chan struct{}
	for _, key := range keys {
		l.create(key)

		ch, _ = l.lock.Load(key)
		select {
		case ch <- struct{}{}:
			last++
		case <-cancel:
			unlock()
			return true, func() {}
		}
	}

	return false, unlock
}

func (l *KeyLock) sort(keys []string) []string {
	kk := make([]string, len(keys))
	copy(kk, keys)
	sort.Strings(kk)
	return kk
}
