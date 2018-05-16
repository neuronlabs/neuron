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

func NewSafeArraySet() *SafeArraySet {
	return &SafeArraySet{ArraySet: *newArraySet()}
}

// SafeArraySet is a concurrent safe Set implementation based on the array.
type SafeArraySet struct {
	ArraySet
	sync.RWMutex
}

func (s *SafeArraySet) Add(value interface{}) bool {
	s.Lock()
	defer s.Unlock()
	return s.ArraySet.Add(value)
}

func (s *SafeArraySet) Contains(value interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	return s.ArraySet.Contains(value)
}

func (s *SafeArraySet) Length() int {
	s.RLock()
	defer s.RUnlock()
	return s.ArraySet.Length()
}

func (s *SafeArraySet) AddMany(values ...interface{}) []interface{} {
	s.Lock()
	defer s.Unlock()

	return s.ArraySet.AddMany(values...)
}

// ArraySet is an implementation of the Set based on the array.
// It implements Set interface.
// It is a good choice if the value set is relatively small and the
// equality checks are not expensive. i.e. set of 10-50 integers.
type ArraySet struct {
	values []interface{}
}

func NewArraySet() *ArraySet {
	return newArraySet()
}

func newArraySet() *ArraySet {
	return &ArraySet{values: make([]interface{}, 0)}
}

func (s *ArraySet) Add(value interface{}) bool {
	if !s.Contains(value) {
		s.values = append(s.values, value)
		return true
	}
	return false
}

func (s *ArraySet) Contains(value interface{}) bool {
	for _, element := range s.values {
		if element == value {
			return true
		}
	}
	return false
}

func (s *ArraySet) Length() int {
	return len(s.values)
}

func (s *ArraySet) AddMany(values ...interface{}) []interface{} {
	temp := newArraySet()
	for _, value := range values {
		if addedCorrectly := s.Add(value); addedCorrectly {
			temp.Add(value)
		}
	}
	return temp.values

}

// HashSet is a concurrent safe Set interface implementation based on the hashmap
type HashSet struct {
	values map[interface{}]struct{}
	sync.RWMutex
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
	s.RLock()
	defer s.RUnlock()
	_, ok := s.values[value]
	return ok
}

func (s *HashSet) Length() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.values)
}
