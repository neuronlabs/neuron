package safemap

import (
	"sync"
)

// SafeHashMap is a concurrent safe Set interface implementation based on the hashmap.
type SafeHashMap struct {
	values map[interface{}]interface{}
	sync.Mutex
}

// New creates new safe hashmap
func New() *SafeHashMap {
	return &SafeHashMap{values: make(map[interface{}]interface{})}
}

// Set sets the value at given _key
func (s *SafeHashMap) Set(key, value interface{}) {
	s.Lock()
	defer s.Unlock()
	s.values[key] = value
}

// Contains checks if given hash map has a given value
func (s *SafeHashMap) Contains(value interface{}) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.values[value]
	return ok
}

// Get gets the safe hashmap value
func (s *SafeHashMap) Get(key interface{}) (interface{}, bool) {
	s.Lock()
	defer s.Unlock()
	value, ok := s.values[key]
	return value, ok
}

// Map returns the hash value map.
func (s *SafeHashMap) Map() map[interface{}]interface{} {
	return s.values
}

// UnsafeSet adds the value at given key even if the map is locked
func (s *SafeHashMap) UnsafeSet(key, value interface{}) {
	s.values[key] = value
}

// UnsafeGet gets the value at given key even if the map is locked
func (s *SafeHashMap) UnsafeGet(key interface{}) (interface{}, bool) {
	value, ok := s.values[key]
	return value, ok
}

// Copy copies the safe map
func (s *SafeHashMap) Copy() *SafeHashMap {
	copied := New()
	for key, value := range s.values {
		copied.values[key] = value
	}
	return copied
}

// Values gets the hashmap values
func (s *SafeHashMap) Values() map[interface{}]interface{} {
	return s.values
}

// Length gets the hash map length
func (s *SafeHashMap) Length() int {
	s.Lock()
	defer s.Unlock()
	return len(s.values)
}
