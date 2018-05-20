package jsonapi

import (
	"reflect"
	"testing"
)

func TestHashSetAdd(t *testing.T) {
	set := NewHashSet()
	assertEqual(t, 0, set.Length())

	assertTrue(t, set.Add(5))
	assertTrue(t, set.Add("this"))
	assertFalse(t, set.Add(5))

}

func TestHashSetContains(t *testing.T) {
	set := NewHashSet()

	input := []interface{}{}
	for i := 0; i < 1000; i++ {
		input = append(input, i)
	}

	for _, v := range input {
		assertTrue(t, set.Add(v))
	}

	for _, v := range input {
		assertFalse(t, set.Add(v))
		assertTrue(t, set.Contains(v))
	}
}

func TestHashSetAddMany(t *testing.T) {
	input := []interface{}{}

	for i := 0; i < 1000; i++ {
		input = append(input, i)
	}

	set := NewHashSet()

	notIn := set.AddMany(input...)

	assertEqual(t, len(notIn), len(input))
}

func buildSlice(input int) (slc []interface{}) {
	for i := 0; i < input; i++ {
		slc = append(slc, i)
	}
	return
}

var small, medium, big int = 50, 1000, 10000

var (
	smalls  = buildSlice(small)
	mediums = buildSlice(medium)
	bigs    = buildSlice(big)
)

func TestHashSetLength(t *testing.T) {
	set := NewHashSet()

	assertEqual(t, 0, set.Length())

	set.Add(1)

	assertEqual(t, 1, set.Length())

	set.Add(1)

	assertEqual(t, 1, set.Length())
}

func benchmarkSet(input []interface{}, set Set, bN int) {
	for i := 0; i < bN; i++ {
		set.AddMany(input...)
	}
}

// func BenchmarkArraySetSmall(b *testing.B) {
// 	set := NewArraySet()
// 	benchmarkSet(smalls, set, b.N)
// }

// func BenchmarkArraySetMedium(b *testing.B) {
// 	set := NewArraySet()
// 	benchmarkSet(mediums, set, b.N)
// }

// func BenchmarkArraySetBigs(b *testing.B) {
// 	set := NewArraySet()
// 	benchmarkSet(bigs, set, b.N)
// }

// func BenchmarkSafeArraySetSmall(b *testing.B) {
// 	set := NewSafeArraySet()
// 	benchmarkSet(smalls, set, b.N)
// }

// func BenchmarkSafeArraySetMedium(b *testing.B) {
// 	set := NewSafeArraySet()
// 	benchmarkSet(mediums, set, b.N)
// }

// func BenchmarkSafeArraySetBig(b *testing.B) {
// 	set := NewSafeArraySet()
// 	benchmarkSet(bigs, set, b.N)
// }

func BenchmarkHashSetSmall(b *testing.B) {
	set := NewHashSet()
	benchmarkSet(smalls, set, b.N)
}

func BenchmarkHashSetMedium(b *testing.B) {
	set := NewHashSet()
	benchmarkSet(mediums, set, b.N)
}

func BenchmarkHashSetBig(b *testing.B) {
	set := NewHashSet()
	benchmarkSet(bigs, set, b.N)
}

// func BenchmarkCompareInterfaces(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 	Outer:
// 		for _, v := range mediums {
// 			for _, checker := range mediums {
// 				if checker == v {
// 					continue Outer
// 				}
// 			}
// 		}
// 	}
// }

// func BenchmarkCompareInts(b *testing.B) {
// 	mediumInts := []int{}
// 	for i := 0; i < medium; i++ {
// 		mediumInts = append(mediumInts, i)
// 	}

// 	for i := 0; i < b.N; i++ {
// 	Outer:
// 		for _, v := range mediumInts {
// 			for _, checker := range mediumInts {
// 				if checker == v {
// 					continue Outer
// 				}
// 			}
// 		}
// 	}
// }

type basicKind int

const (
	intKind basicKind = iota
	uintKind
	stringKind
	otherKind
)

func getBasicKind(value reflect.Value) basicKind {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintKind
	case reflect.String:
		return stringKind
	}
	return otherKind
}

func BenchmarkCompareInterfaceFast(b *testing.B) {
	// var kind basicKind = intKind
	for i := 0; i < b.N; i++ {
	Outer:
		for _, v := range mediums {
			for _, checker := range mediums {
				// if kind == intKind {
				intChecker, ok := checker.(int64)
				if ok {
					intVal, _ := v.(int64)
					if intChecker == intVal {
						continue Outer
					}

				}
				// }
			}
		}
	}
}
