package class

import (
	"errors"
	"strings"
)

const (
	majorBitSize = 7
	minorBitSize = 10
	indexBitSize = 32 - majorBitSize - minorBitSize

	maxIndexValue = (2 << (indexBitSize - 1)) - 1
	maxMinorValue = (2 << (minorBitSize - 1)) - 1
	maxMajorValue = (2 << (majorBitSize - 1)) - 1

	majorMinorMask = uint32((2<<(majorBitSize+minorBitSize-1) - 1) << indexBitSize)
)

func init() {
	registerClasses()
}

func registerClasses() {
	registerConfigClasses()
	registerEncodingClasses()
	registerInternalClasses()
	registerModelClasses()
	registerRepositoryClasses()
	registerQueryClasses()
	registerCommonClasses()
	registerLanguageClasses()
}

// Class is the neuron error classification model.
// it is composed of the major, minor and index subclassifications.
// Each subclassifaction is a different length number, where
// major is composed of 7, minor 10 and index of 15 bits.
// Example:
//  44205263 in a binary form is 00000010101000101000010011001111 which decomposes into:
//	0000001 - major (7 bit) - 1
//
//		   0101000101 - minor (10 bit) - 10
//
//					 000010011001111 - index (15 bit) - 1231
// Major should be a global scope division like 'Repository', 'Marshaler', 'Controller' etc.
// Minor should divide the 'major' into subclasses like the Repository filter builders, marshaler - invalid field etc.
// Index is the most precise classification - i.e. Repository - filter builder - unsupported operator.
type Class uint32

// Index is a four digit number unique within given minor and major.
func (c Class) Index() Index {
	return c.index()
}

// IsMajor checks if the given class is composed of provided major 'm'.
func (c Class) IsMajor(m Major) bool {
	return c.major() == m
}

// Major is a single digit major classification.
func (c Class) Major() Major {
	return c.major()
}

// Minor is a double digit minor classification unique within given major.
func (c Class) Minor() Minor {
	return c.minor()
}

// MjrMnrMasked returns the class value masked by the major and minor value only.
func (c Class) MjrMnrMasked() uint32 {
	return uint32(c) & majorMinorMask
}

// String implements fmt.Stringer interface.
func (c Class) String() string {
	var names []string

	major := c.Major().Name()
	names = append(names, strings.Split(major, " ")...)
	minor := c.minor()

	if !minor.InBounds() {
		return strings.Join(names, "")
	}
	names = append(names, strings.Split(minor.Name(), " ")...)

	index := c.Index()
	if index.Valid() {
		names = append(names, strings.Split(index.Name(), " ")...)
	}

	return strings.Join(names, "")
}

func (c Class) index() Index {
	return Index{value: uint16(c.indexValue()), minor: c.minor(), own: true}
}

func (c Class) indexValue() int {
	return int(c & maxIndexValue)
}

func (c Class) major() Major {
	return Major(c.majorValue())
}

func (c Class) minor() Minor {
	return Minor{value: uint16(c.minorValue()), major: c.major(), own: true}
}

func (c Class) majorValue() int {
	return int(c >> (32 - majorBitSize))
}

func (c Class) minorValue() int {
	return int(c) >> (indexBitSize) & maxMinorValue
}

// MustNewMinorClass creates new minor class for provided argument.
// If the minor value is not vlaid the function panics.
func MustNewMinorClass(minor Minor) Class {
	c, err := NewMinorClass(minor)
	if err != nil {
		panic(err)
	}
	return c
}

// NewClass gets new class from the provided 'minor' and 'index'.
// If any of the arguments is not valid or out of bands the function returns an error.
func NewClass(index Index) (Class, error) {
	minor := index.Minor()

	if !minor.Major().InBounds() {
		return Class(0), errors.New("provided invalid major")
	}

	if !minor.valid() {
		return Class(0), errors.New("provided invalid minor")
	}

	if !index.valid() {
		return Class(0), errors.New("provided invalid index")
	}

	return Class(uint32(minor.major)<<(32-majorBitSize) | uint32(minor.value)<<(32-minorBitSize-majorBitSize) | uint32(index.value)), nil
}

// NewMinorClass gets the class from provided 'minor'.
// The function gets minor's major and gets the major/minor class.
func NewMinorClass(minor Minor) (Class, error) {
	if !minor.Major().InBounds() {
		return Class(0), errors.New("provided invalid major")
	}

	if !minor.valid() {
		return Class(0), errors.New("provided invalid minor")
	}

	return Class(uint32(minor.major)<<(32-majorBitSize) | uint32(minor.value)<<(32-minorBitSize-majorBitSize)), nil
}
