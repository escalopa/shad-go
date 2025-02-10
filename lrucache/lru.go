//go:build !solution

package lrucache

import (
	"container/list"
	"sync"
)

type entry struct {
	key int
	val int
}

type cache struct {
	l  *list.List
	m  map[int]*list.Element
	c  int
	mu sync.RWMutex
}

func (c *cache) Get(key int) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if e, ok := c.m[key]; ok {
		c.l.MoveToFront(e)
		return e.Value.(entry).val, true
	}
	return 0, false
}

func (c *cache) Set(key, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.m[key]; ok {
		c.l.MoveToFront(e)
		e.Value = entry{key: key, val: value}
		return
	}
	// handle eviction
	if c.l.Len() == c.c {
		e := c.l.Back()
		delete(c.m, e.Value.(entry).key)
		c.l.Remove(e)
	}
	e := entry{key: key, val: value}
	c.m[key] = c.l.PushFront(e)
}

func (c *cache) Range(f func(key int, value int) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for e := c.l.Back(); e != nil; e = e.Prev() {
		ee := e.Value.(entry)
		if !f(ee.key, ee.val) {
			break
		}
	}
}

func (c *cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.l.Init()
	c.m = make(map[int]*list.Element)
}

func New(cap int) Cache {
	if cap <= 0 {
		return dummyCache{}
	}

	return &cache{
		l: list.New(),
		m: make(map[int]*list.Element),
		c: cap,
	}
}

type dummyCache struct{}

func (dummyCache) Get(_ int) (int, bool)             { return 0, false }
func (dummyCache) Set(_, _ int)                      {}
func (dummyCache) Range(_ func(key, value int) bool) {}
func (dummyCache) Clear()                            {}
