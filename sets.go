package jsonapi

import (
	"sync"
)

type Set interface {
	// Adds the value to the set. If the value was not found within given set already
	// the function reutrns true. otherwise it should return false.
	Add(value interface{}) bool

	// AddMany adds mulitple values to the set. The unique values that were not present in the set
	// are returned
	AddMany(values ...interface{}) (uniquesNotIn []interface{})

	// Contains check if given value exists within given Set.
	Contains(value interface{}) bool

	// Length returns the length of given set
	Length() int
}

// HashSet is a concurrent safe Set interface implementation based on the hashmap
type HashSet struct {
	values map[interface{}]struct{}
	sync.Mutex
}

func NewHashSet() *HashSet {
	return &HashSet{values: make(map[interface{}]struct{})}
}

func (s *HashSet) Add(value interface{}) bool {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.values[value]; !ok {
		s.values[value] = struct{}{}
		return true
	}
	return false
}

func (s *HashSet) AddMany(values ...interface{}) []interface{} {
	s.Lock()
	defer s.Unlock()

	temp := map[interface{}]struct{}{}
	for _, value := range values {
		if _, ok := s.values[value]; !ok {
			s.values[value] = struct{}{}
			temp[value] = struct{}{}
		}
	}

	uniqueNotIn := make([]interface{}, len(temp))

	i := 0
	for uniqueValue := range temp {
		uniqueNotIn[i] = uniqueValue
		i++
	}

	return uniqueNotIn
}

func (s *HashSet) Contains(value interface{}) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.values[value]
	return ok
}

func (s *HashSet) Length() int {
	s.Lock()
	defer s.Unlock()
	return len(s.values)
}
