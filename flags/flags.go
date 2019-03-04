package flags

// Container is the container for any enum type flags.
// The stored flag can be
type Container uint64

func New() *Container {
	c := Container(uint64(0))
	return &c
}

// Copy returns the new copy of the container
func (f *Container) Copy() *Container {
	var cp uint64
	cp = uint64(*f)
	cpC := Container(cp)
	return &cpC
}

// Set sets the provided flag value.
func (f *Container) Set(flag uint, value bool) {
	f.set(flag, value)
	return
}

// SetFirst if any of the provided flag containers has a value for provided flag
// then the first one in the order will be written.
// Otherwise nothing happens
func (f *Container) SetFirst(flag uint, containers ...*Container) (found bool) {
	for _, c := range containers {
		ok := c.isSet(flag)
		if ok {
			// if found set the first and break the loop
			f.Set(flag, c.value(flag))
			found = true
			break
		}
	}
	return
}

// SetFrom sets the flag value from the 'root' Container.
// If the value was not set within the root container, the value would in
// 'f' container would not be set.
func (f *Container) SetFrom(flag uint, root *Container) (found bool) {
	var value bool
	value, found = root.Get(flag)
	if found {
		f.set(flag, value)
	}
	return
}

// Get returns the value and the boolean if the flag were already set.
func (f Container) Get(flag uint) (value, ok bool) {
	if ok = f.isSet(flag); ok {
		value = f.value(flag)
	}
	return
}

// IsSet returns the boolean wether the value were already set.
func (f Container) IsSet(flag uint) (ok bool) {
	return f.isSet(flag)
}

// Value returns the flag boolean value
func (f Container) Value(flag uint) (value bool) {
	return f.value(flag)
}

// GetFirst the first flag value that exists in the provided containers.
func GetFirst(flag uint, containers ...*Container) (value, ok bool) {
	for _, f := range containers {
		if ok = f.isSet(flag); ok {
			value = f.value(flag)
			break
		}
	}
	return
}

func (f Container) isSet(flag uint) (ok bool) {
	// fmt.Printf("isSet Flag: %v\n", flag)
	// fmt.Printf("isSet f: \t%.64b\n", f)
	v := f >> (2 * flag)
	// fmt.Printf("isSet value: \t%.64b\n", v)
	if v&1 > 0 {
		ok = true
	} else {
		ok = false
	}
	// fmt.Printf("isSet ok: %v\n", ok)
	return
}

func (f *Container) set(flag uint, value bool) {
	// set the 'isSet' flag at 2*n position
	*f = *f | 1<<(flag*2)

	// set thte 'value' flag at 2*n+1 position
	if value {
		*f = *f | (1 << (2*flag + 1))
	} else {
		*f = *f &^ (1 << (2*flag + 1))
	}
}

func (f Container) value(flag uint) (value bool) {
	// fmt.Printf("value Flag: %v\n", flag)
	// fmt.Printf("value f: \t%.64b\n", f)
	v := f >> (2*flag + 1)
	// fmt.Printf("value v: \t%.64b\n", v)

	if v&1 > 0 {
		value = true
	} else {
		value = false
	}
	// fmt.Printf("value, value: \t%v\n", value)
	return
}
