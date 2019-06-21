package class

import (
	"errors"
	"sync"
)

// Index is a 4 digit lowest level error classification.
// It is the most precise division - i.e.:
// 'major' Repository
//	'minor' filter builder
//	 'index' unsupported operator.
type Index struct {
	value uint16
	minor Minor
	own   bool
}

// Class gets the index related class.
func (i Index) Class() Class {
	if !i.valid() {
		return Class(0)
	}
	return Class(uint32(i.minor.major)<<(32-majorBitSize) | uint32(i.minor.value)<<(indexBitSize) | uint32(i.value))
}

// InBounds checks if the index value is in the possible 15-bit range.
func (i Index) InBounds() bool {
	return i.inBounds()
}

// Name gets the index stored name.
func (i Index) Name() string {
	if !i.valid() {
		return ""
	}
	return i.container().name(i)
}

// Minor returns index related Minor.
func (i Index) Minor() Minor {
	return i.minor
}

// Valid checks if the provided index is valid.
func (i Index) Valid() bool {
	return i.valid()
}

// Value gets the index uint16 value.
func (i Index) Value() uint16 {
	return i.value
}

func (i Index) container() *indexContainer {
	if !i.valid() {
		return nil
	}

	return i.minor.container().indexContainer(i.minor)
}

func (i Index) inBounds() bool {
	return i.value>>indexBitSize == 0 || i.value&maxIndexValue != 0
}

func (i Index) containerIndex() uint16 {
	return i.value - 1
}

func (i Index) valid() bool {
	return i.inBounds() && i.own
}

type indexContainer struct {
	uniqueNames  map[string]struct{}
	names        []string
	descriptions []string
	nextID       uint16
	lock         sync.Mutex
}

func (i *indexContainer) description(id Index) string {
	i.expandIfRequired(id.containerIndex())
	return i.descriptions[id.containerIndex()]
}

func (i *indexContainer) expandIfRequired(id uint16) {
	size := len(i.names)

	// check if the given 'id' is not in the bounds
	// of the current name and description slice
	if int(id) < size-1 {
		return
	}

	size *= 2

	temp := i.names
	i.names = make([]string, size)
	copy(i.names, temp)

	temp = i.descriptions
	i.descriptions = make([]string, size)
	copy(i.descriptions, temp)
}

func (i *indexContainer) name(id Index) string {
	i.expandIfRequired(id.containerIndex())
	return i.names[id.containerIndex()]
}

func (i *indexContainer) new(minor Minor, name string, description ...string) (Index, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	// check uniqueness of the name
	_, exists := i.uniqueNames[name]
	if exists {
		return Index{}, errors.New("index name already exists")
	}

	next := i.next(minor)

	i.names[next.containerIndex()] = name
	i.uniqueNames[name] = struct{}{}

	if len(description) == 1 {
		i.descriptions[next.containerIndex()] = description[0]
	}

	return next, nil
}

func (i *indexContainer) next(minor Minor) Index {
	// increase the id counter
	defer func() {
		i.nextID++
	}()

	index := Index{value: i.nextID, minor: minor, own: true}
	i.expandIfRequired(index.containerIndex())

	return index
}

func newIndexContainer() *indexContainer {
	return &indexContainer{
		uniqueNames:  make(map[string]struct{}),
		names:        make([]string, 8),
		descriptions: make([]string, 8),
		nextID:       1,
	}
}
