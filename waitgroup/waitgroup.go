//go:build !solution

package waitgroup

var panicErr = &struct{}{}

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup struct {
	counter int

	wait    chan struct{}  // for synchronization of Wait
	done    chan *struct{} // for synchronization of Add
	channel chan int
}

// New creates WaitGroup.
func New() *WaitGroup {
	wg := &WaitGroup{
		channel: make(chan int),
		done:    make(chan *struct{}),
		wait:    make(chan struct{}),
	}
	close(wg.wait) // Wait should not block initially
	wg.run()
	return wg
}

func (wg *WaitGroup) run() {
	go func() {
		for v := range wg.channel {
			wg.counter += v
			if wg.counter < 0 {
				wg.counter -= v
				wg.done <- panicErr
				continue
			}
			wg.updateWait()
			wg.done <- nil
		}
	}()
}

func (wg *WaitGroup) updateWait() {
	var isClosed bool
	select {
	case <-wg.wait:
		isClosed = true
	default:
	}

	// case when the counter was zero(and currently not) and the wait channel was closed
	if wg.counter != 0 && isClosed {
		wg.wait = make(chan struct{})
	}

	// case when the counter is zero and the wait channel is open
	if wg.counter == 0 && !isClosed {
		close(wg.wait)
	}
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
	wg.channel <- delta
	if res := <-wg.done; res == panicErr {
		panic("negative WaitGroup counter")
	}
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	<-wg.wait // wait for the counter to be zero
}
