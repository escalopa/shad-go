//go:build !solution

package rwmutex

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// If a goroutine holds a RWMutex for reading and another goroutine might
// call Lock, no goroutine should expect to be able to acquire a read lock
// until the initial read lock is released. In particular, this prohibits
// recursive read locking. This is to ensure that the lock eventually becomes
// available; a blocked Lock call excludes new readers from acquiring the
// lock.

type Mutex struct {
	channel chan struct{}
}

func (m *Mutex) Lock() {
	m.channel <- struct{}{}
}

func (m *Mutex) Unlock() {
	<-m.channel
}

type RWMutex struct {
	counter int

	readChan  chan struct{}
	writeChan chan struct{}

	mutex Mutex
}

// New creates *RWMutex.
func New() *RWMutex {
	rw := &RWMutex{
		readChan:  make(chan struct{}, 1),
		writeChan: make(chan struct{}, 1),
		mutex: Mutex{
			channel: make(chan struct{}, 1),
		},
	}
	return rw
}

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
func (rw *RWMutex) RLock() {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	if rw.counter == 0 {
		rw.readChan <- struct{}{}
	}
	rw.counter++
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	if rw.counter == 0 {
		panic("unlock of unlocked RWMutex")
	}
	rw.counter--
	if rw.counter == 0 {
		<-rw.readChan
	}
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	rw.writeChan <- struct{}{}
	rw.readChan <- struct{}{}
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked RWMutex is not associated with a particular
// goroutine. One goroutine may RLock (Lock) a RWMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (rw *RWMutex) Unlock() {
	select {
	case <-rw.writeChan:
	default:
		panic("unlock of unlocked RWMutex")
	}
	<-rw.readChan
}
