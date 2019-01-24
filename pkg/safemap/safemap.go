package safemap

import (
	"sync"
)

// SafeHashMap is a concurrent safe Set interface implementation based on the hashmap
type SafeHashMap struct {
	values map[interface{}]interface{}
	sync.Mutex
}

func New() *SafeHashMap {
	return &SafeHashMap{values: make(map[interface{}]interface{})}
}

func (s *SafeHashMap) Add(key, value interface{}) {
	s.Lock()
	defer s.Unlock()
	s.values[key] = value
}

func (s *SafeHashMap) Contains(value interface{}) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.values[value]
	return ok
}

func (s *SafeHashMap) Get(key interface{}) (interface{}, bool) {
	s.Lock()
	defer s.Unlock()
	value, ok := s.values[key]
	return value, ok
}

// UnsafeAdd adds the value at given key even if the map is locked
func (s *SafeHashMap) UnsafeAdd(key, value interface{}) {
	s.values[key] = value
}

// UnsafeGet gets the value at given key even if the map is locked
func (s *SafeHashMap) UnsafeGet(key interface{}) (interface{}, bool) {
	value, ok := s.values[key]
	return value, ok
}

func (s *SafeHashMap) Copy() *SafeHashMap {
	copied := New()
	for key, value := range s.values {
		copied.values[key] = value
	}
	return copied
}

func (s *SafeHashMap) Values() map[interface{}]interface{} {
	return s.values
}

func (s *SafeHashMap) Length() int {
	s.Lock()
	defer s.Unlock()
	return len(s.values)
}
