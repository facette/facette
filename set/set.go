package set

import (
	"fmt"
	"sync"
)

// Set represents a set instance.
type Set struct {
	l sync.RWMutex
	m map[interface{}]struct{}
}

// New creates a new set instance.
func New(items ...interface{}) *Set {
	s := &Set{m: make(map[interface{}]struct{})}
	if len(items) > 0 {
		s.Add(items...)
	}
	return s
}

// Add inserts new items into the set.
func (s *Set) Add(items ...interface{}) {
	s.l.Lock()
	defer s.l.Unlock()

	for _, v := range items {
		s.m[v] = struct{}{}
	}
}

// Has returns whether or not items are present in the set.
func (s *Set) Has(items ...interface{}) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	ok := true
	for _, v := range items {
		_, ok = s.m[v]
		if !ok {
			break
		}
	}
	return ok
}

// Len returns the number of items in the set.
func (s *Set) Len() int {
	s.l.RLock()
	defer s.l.RUnlock()

	return len(s.m)
}

// Remove removes items from the set.
func (s *Set) Remove(items ...interface{}) {
	s.l.Lock()
	defer s.l.Unlock()

	for _, v := range items {
		delete(s.m, v)
	}
}

// Slice returns a slice of all set items.
func (s *Set) Slice() []interface{} {
	var result []interface{}

	s.l.RLock()
	defer s.l.RUnlock()

	for v := range s.m {
		result = append(result, v)
	}

	return result
}

// StringSlice returns a strings slice representation of all set items.
func StringSlice(s *Set) []string {
	var result []string
	for _, v := range s.Slice() {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
}
