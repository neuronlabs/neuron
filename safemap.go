package jsonapi

import (
	"sync"
)

// SafeHashMap is a concurrent safe Set interface implementation based on the hashmap
type SafeHashMap struct {
	values map[interface{}]interface{}
	sync.Mutex
}

func NewSafeHashMap() *SafeHashMap {
	return &SafeHashMap{values: make(map[interface{}]interface{})}
}

// func (s *SafeHashMap) Add(value interface{}) bool {
// 	s.Lock()
// 	defer s.Unlock()
// 	if _, ok := s.values[value]; !ok {
// 		s.values[value] = struct{}{}
// 		return true
// 	}
// 	return false
// }

// func (s *SafeHashMap) AddMany(values ...interface{}) []interface{} {
// 	s.Lock()
// 	defer s.Unlock()

// 	temp := map[interface{}]struct{}{}
// 	for _, value := range values {
// 		if _, ok := s.values[value]; !ok {
// 			s.values[value] = struct{}{}
// 			temp[value] = struct{}{}
// 		}
// 	}

// 	uniqueNotIn := make([]interface{}, len(temp))

// 	i := 0
// 	for uniqueValue := range temp {
// 		uniqueNotIn[i] = uniqueValue
// 		i++
// 	}

// 	return uniqueNotIn
// }

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

func (s *SafeHashMap) Copy() *SafeHashMap {
	copied := NewSafeHashMap()
	for key, value := range s.values {
		copied.values[key] = value
	}
	return copied
}

func (s *SafeHashMap) Length() int {
	s.Lock()
	defer s.Unlock()
	return len(s.values)
}
