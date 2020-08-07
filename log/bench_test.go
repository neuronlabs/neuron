package log

import (
	"fmt"
	"strings"
	"testing"
)

const benchConst = "some pretty long name"

// BenchmarkConcatenate benchmarks concatenating multiple strings using '+'.
func BenchmarkConcatenate(b *testing.B) {
	previous := benchConst
	var v string
	for i := 0; i < b.N; i++ {
		v = "some" + previous + "some"
	}
	b.Log(v)
}

// BenchmarkFmt benchmarks concatenating multiple strings using 'fmt.Sprintf'.
func BenchmarkFmt(b *testing.B) {
	previous := benchConst
	var v string
	for i := 0; i < b.N; i++ {
		v = fmt.Sprintf("some%ssome", previous)
	}
	b.Log(v)
}

// BenchmarkBuilder benchmarks concatenating multiple strings using 'strings.Builder'.
func BenchmarkBuilder(b *testing.B) {
	previous := benchConst
	var v string
	for i := 0; i < b.N; i++ {
		sb := strings.Builder{}
		sb.WriteString("some")
		sb.WriteString(previous)
		sb.WriteString("some")
		v = sb.String()
	}
	b.Log(v)
}
