package class

import (
	"errors"
	"sync"
)

// Minor is a 3 digit mid level error classification.
type Minor struct {
	value uint16
	major Major

	own bool
}

// Description gets the minor's description.
func (m Minor) Description() string {
	if !m.valid() {
		return ""
	}
	return m.container().description(m)
}

// InBounds checks if the index value is in the possible 15-bit range.
func (m Minor) InBounds() bool {
	return m.inBounds()
}

// Indexes returns minor's registered indexes.
func (m Minor) Indexes() []Index {
	if !m.valid() {
		return nil
	}

	indexContainer := m.container().indexContainer(m)

	indexes := make([]Index, indexContainer.nextID)
	for i := uint16(0); i < indexContainer.nextID; i++ {
		indexes[i] = Index{value: i, minor: m, own: true}
	}

	return indexes
}

// MustRegisterIndex registers and returns index for given minor value.
// Panics if the index name already exists or the minor is not valid.
func (m Minor) MustRegisterIndex(name string, description ...string) Index {
	idx, err := m.RegisterIndex(name, description...)
	if err != nil {
		panic(err)
	}
	return idx
}

// Name gets the minors registered name.
func (m Minor) Name() string {
	if !m.valid() {
		return ""
	}
	return m.container().name(m)
}

// Major gets the minor's root Major.
func (m Minor) Major() Major {
	return m.major
}

// Valid checks if the Minor is valid.
func (m Minor) Valid() bool {
	return m.valid()
}

// Value gets the minor's uint16 value.
func (m Minor) Value() uint16 {
	return m.value
}

// RegisterIndex registers the index for given Minor.
func (m Minor) RegisterIndex(name string, description ...string) (Index, error) {
	if !m.valid() {
		return Index{}, errors.New("invalid minor provided")
	}

	// get the minor container
	minorCtr := m.container()

	// get the indexContainer for given minor
	indexContainer := minorCtr.indexContainer(m)
	id, err := indexContainer.new(m, name, description...)
	if err != nil {
		return Index{}, err
	}
	return id, nil
}

func (m Minor) container() *minorsContainer {
	return majors.minors[m.major.containerIndex()]
}

func (m Minor) inBounds() bool {
	return m.value>>minorBitSize == 0 && m.value != 0
}

func (m Minor) containerIndex() uint16 {
	return m.value - 1
}

func (m Minor) valid() bool {
	return m.inBounds() && m.own
}

type minorsContainer struct {
	uniqueNames  map[string]struct{}
	names        []string
	descriptions []string
	indices      []*indexContainer

	nextID uint16
	lock   sync.Mutex
}

func (m *minorsContainer) description(id Minor) string {
	m.expandIfRequired(id.containerIndex())
	return m.descriptions[id.containerIndex()]
}

func (m *minorsContainer) expandIfRequired(id uint16) {
	size := len(m.names)

	// check if the given 'id' is not in the bounds
	// of the current name and description slice
	if int(id) < size-1 {
		return
	}

	size *= 2

	temp := m.names
	m.names = make([]string, size)
	copy(m.names, temp)

	temp = m.descriptions
	m.descriptions = make([]string, size)
	copy(m.descriptions, temp)

	tempIndices := m.indices
	m.indices = make([]*indexContainer, size)
	copy(m.indices, tempIndices)
}

func (m *minorsContainer) indexContainer(id Minor) *indexContainer {
	m.expandIfRequired(id.containerIndex())

	indexContainer := m.indices[id.containerIndex()]
	if indexContainer == nil {
		indexContainer = newIndexContainer()
		m.indices[id.containerIndex()] = indexContainer
	}

	return indexContainer
}

func (m *minorsContainer) name(id Minor) string {
	m.expandIfRequired(id.containerIndex())
	return m.names[id.containerIndex()]
}

func (m *minorsContainer) new(major Major, name string, description ...string) (Minor, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// check uniqueness of the name
	_, exists := m.uniqueNames[name]
	if exists {
		return Minor{}, errors.New("minor name already exists")
	}

	next := m.next(major)

	m.names[next.containerIndex()] = name
	m.uniqueNames[name] = struct{}{}
	if len(description) >= 1 {
		m.descriptions[next.containerIndex()] = description[0]
	}

	return next, nil
}

func (m *minorsContainer) next(major Major) Minor {
	// increase the id counter
	defer func() {
		m.nextID++
	}()
	value := m.nextID
	m.expandIfRequired(m.nextID)

	return Minor{value: value, major: major, own: true}
}

func newMinorsContainer() *minorsContainer {
	return &minorsContainer{
		uniqueNames:  make(map[string]struct{}),
		names:        make([]string, 10),
		descriptions: make([]string, 10),
		indices:      make([]*indexContainer, 10),
		nextID:       1,
	}
}
