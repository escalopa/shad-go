//go:build !solution

package once

// Once describes an object that will perform exactly one action.
type Once struct {
	ch    chan struct{}
	queue chan struct{}
}

// New creates Once.
func New() *Once {
	return &Once{
		ch:    make(chan struct{}, 1),
		queue: make(chan struct{}, 1),
	}
}

// Do calls the function f if and only if Do is being called for the
// first time for this instance of Once. In other words, given
//
//	once := New()
//
// if once.Do(f) is called multiple times, only the first call will invoke f,
// even if f has a different value in each invocation. A new instance of
// Once is required for each function to execute.
//
// Do is intended for initialization that must be run exactly once.
//
// Because no call to Do returns until the one call to f returns, if f causes
// Do to be called, it will deadlock.
//
// If f panics, Do considers it to have returned; future calls of Do return
// without calling f.
func (o *Once) Do(f func()) {
	o.queue <- struct{}{}
	defer func() { <-o.queue }()
	o.doSlow(f)
}

func (o *Once) doSlow(f func()) {
	select {
	case o.ch <- struct{}{}:
		f()
	default:
	}
}
