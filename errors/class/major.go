package class

import (
	"errors"
	"sync"
)

var majors = newMajors()

// Major is a 7 bit top level error classification.
type Major uint8

// Description gets the major registered description.
func (m Major) Description() string {
	return majors.description(m)
}

// InBounds checks if the major value is not greater than the allowed size.
func (m Major) InBounds() bool {
	return (m >> majorBitSize) == 0
}

// Minors gets the registered minors for given major 'm'.
func (m Major) Minors() []Minor {
	// check if the major is in bounds
	if !m.InBounds() {
		return nil
	}

	// get the minor container that contains current ID
	container := majors.minorContainer(m)

	// create minors slice
	minors := make([]Minor, container.nextID)

	for i := uint16(0); i < container.nextID-1; i++ {
		minors[i] = Minor{value: i + 1, major: m, own: true}
	}

	return minors
}

// MustRegisterMinor registers the minor classification for
// given Major 'm', 'name' - unique name for given Major and
// optional 'description'. Panics when the major is invalid.
func (m Major) MustRegisterMinor(name string, description ...string) Minor {
	minor, err := m.RegisterMinor(name, description...)
	if err != nil {
		panic(err)
	}
	return minor
}

// Name returns the major registered name.
func (m Major) Name() string {
	return majors.name(m)
}

// RegisterMinor registers the minor classification for
// given Major 'm', 'name' - unique name for given Major and
// optional 'description'.
// If the major is not valid - out of bands - function throws an error.
func (m Major) RegisterMinor(name string, description ...string) (Minor, error) {
	if !m.InBounds() {
		return Minor{}, errors.New("major out of bounds")
	}

	minorCtr := majors.minorContainer(m)

	// create new minor for given container
	minor, err := minorCtr.new(m, name, description...)
	if err != nil {
		return Minor{}, err
	}

	return minor, nil
}

func (m Major) containerIndex() uint8 {
	return uint8(m) - 1
}

// RegisterMajor registers new major error classification with provided
// 'name', and optional 'description' for the major.
// Returns an error if there are already a maximum number of the Major's
// which is maximum uint8 value - 0xff.
func RegisterMajor(name string, descriptions ...string) (Major, error) {
	return majors.new(name, descriptions...)
}

// MustRegisterMajor regsiters new major error classification with provided
// 'name' and optional 'description' for the major.
// Panics when the major already exists.
func MustRegisterMajor(name string, descriptions ...string) Major {
	m, err := RegisterMajor(name, descriptions...)
	if err != nil {
		panic(err)
	}
	return m
}

type majorsContainer struct {
	uniqueNames  map[string]struct{}
	names        []string
	descriptions []string
	minors       []*minorsContainer

	currentID uint16
	idLock    sync.Mutex
}

func (m *majorsContainer) description(major Major) string {
	return m.descriptions[major.containerIndex()]
}

func (m *majorsContainer) minorContainer(v Major) *minorsContainer {
	// get the minor container
	minorCtr := majors.minors[v.containerIndex()]
	if minorCtr == nil {
		// if the container doesn't exists
		minorCtr = newMinorsContainer()
		majors.minors[v.containerIndex()] = minorCtr
	}

	return minorCtr
}

func (m *majorsContainer) name(major Major) string {
	return m.names[major.containerIndex()]
}

func (m *majorsContainer) new(name string, description ...string) (Major, error) {
	m.idLock.Lock()
	defer m.idLock.Unlock()

	_, exists := m.uniqueNames[name]
	if exists {
		return Major(0), errors.New("major name already registered")
	}

	major := m.next()

	// check if the value is int bounds of uint8
	if major > maxMajorValue {
		return major, errors.New("too many major registered")
	}

	m.names[major.containerIndex()] = name
	if len(description) == 1 {
		m.descriptions[major.containerIndex()] = description[0]
	}
	m.uniqueNames[name] = struct{}{}

	return major, nil
}

func (m *majorsContainer) next() Major {
	defer func() {
		m.currentID++
	}()
	return Major(m.currentID)
}

func (m *majorsContainer) reset() {
	majors := newMajors()
	m.currentID = majors.currentID
	m.names = majors.names
	m.descriptions = majors.descriptions
	m.minors = majors.minors
	m.uniqueNames = majors.uniqueNames
}

func newMajors() *majorsContainer {
	return &majorsContainer{
		names:        make([]string, maxMajorValue+1),
		descriptions: make([]string, maxMajorValue+1),
		minors:       make([]*minorsContainer, maxMajorValue+1),
		uniqueNames:  make(map[string]struct{}),
		currentID:    1,
	}
}
