//go:build !solution

package pubsub

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrClosed = errors.New("pubsub is closed")
)

type ID string

//////////////////////////////////
// MySubscription
//////////////////////////////////

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	queue   chan interface{}
	handler MsgHandler
	done    chan struct{}
	wg      sync.WaitGroup
}

func newSubscriber(handler MsgHandler) *MySubscription {
	return &MySubscription{
		queue:   make(chan interface{}, 100),
		handler: handler,
		done:    make(chan struct{}),
	}
}

func (s *MySubscription) send(msg interface{}) {
	select {
	case s.queue <- msg:
		s.wg.Add(1)
	case <-s.done: // if subscription is closed then ignore the message
	}
}

func (s *MySubscription) run() {
	for {
		select {
		case msg := <-s.queue:
			if isClosed(s.done) {
				return
			}
			s.handler(msg)
			s.wg.Done()
		case <-s.done:
			return
		}
	}
}

func (s *MySubscription) Unsubscribe() {
	s.wg.Wait()
	tryClose(s.done)
}

//////////////////////////////////
// Queue
//////////////////////////////////

type (
	Queue struct {
		listener map[ID]*MySubscription
		lock     sync.RWMutex
		done     chan struct{}
	}
)

func NewQueue() *Queue {
	return &Queue{
		listener: make(map[ID]*MySubscription),
		done:     make(chan struct{}),
	}
}

func (q *Queue) Subscribe(sub *MySubscription) {
	q.lock.Lock()
	defer q.lock.Unlock()

	subID := ID(uuid.New().String())
	q.listener[subID] = sub
	go sub.run()

	// unsubscribe when the subscription is done
	go func() {
		<-sub.done
		q.lock.Lock()
		defer q.lock.Unlock()
		delete(q.listener, subID)
	}()
}

func (q *Queue) Publish(msg interface{}) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	for _, sub := range q.listener {
		sub.send(msg)
	}
}

func (q *Queue) close() {
	q.lock.Lock()
	defer q.lock.Unlock()

	if isClosed(q.done) {
		return
	}

	for key, sub := range q.listener {
		sub.Unsubscribe()
		delete(q.listener, key)
	}

	tryClose(q.done)
}

//////////////////////////////////
// MyPubSub
//////////////////////////////////

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	subjects map[string]*Queue
	lock     sync.RWMutex
	done     chan struct{}
}

func NewPubSub() PubSub {
	return &MyPubSub{
		subjects: make(map[string]*Queue),
		done:     make(chan struct{}),
	}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	if isClosed(p.done) {
		return nil, ErrClosed
	}

	sub := newSubscriber(cb)
	p.getQueue(subj).Subscribe(sub)
	return sub, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	if isClosed(p.done) {
		return ErrClosed
	}

	p.getQueue(subj).Publish(msg)
	return nil
}

func (p *MyPubSub) getQueue(subj string) *Queue {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subjects[subj]; !ok {
		p.subjects[subj] = NewQueue()
	}

	return p.subjects[subj]
}

func (p *MyPubSub) Close(ctx context.Context) error {
	f := func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		for key, q := range p.subjects {
			q.close()
			delete(p.subjects, key)
		}

		tryClose(p.done)
	}

	go f()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return nil
	}
}

//////////////////////////////////
// channel helper
//////////////////////////////////

var tryCloseMutex = sync.Mutex{}

func tryClose(done chan struct{}) {
	tryCloseMutex.Lock()
	defer tryCloseMutex.Unlock()

	select {
	case <-done:
	default:
		close(done)
	}
}

func isClosed(done chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
