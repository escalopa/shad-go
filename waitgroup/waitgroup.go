//go:build !solution

package waitgroup

type Mutex struct {
	channel chan struct{}
}

func NewMutex() *Mutex {
	return &Mutex{channel: make(chan struct{}, 1)}
}

func (m *Mutex) Lock() {
	m.channel <- struct{}{}
}

func (m *Mutex) Unlock() {
	select {
	case <-m.channel:
	default:
		panic("unlock of unlocked Mutex")
	}
}

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup struct {
	counter int
	channel chan struct{}
	mutex   Mutex
}

// New creates WaitGroup.
func New() *WaitGroup {
	wg := &WaitGroup{
		channel: make(chan struct{}, 1),
		mutex: Mutex{
			channel: make(chan struct{}, 1),
		},
	}
	return wg
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls have returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	wg.mutex.Lock()
	defer wg.mutex.Unlock()

	if wg.counter == 0 && delta > 0 {
		wg.channel <- struct{}{}
	}

	wg.counter += delta
	if wg.counter < 0 {
		panic("negative WaitGroup counter")
	}

	if wg.counter == 0 {
		<-wg.channel
	}
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	wg.channel <- struct{}{}
	<-wg.channel
}
