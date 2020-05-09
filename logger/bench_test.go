package unilogger

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkConcantate benchmarks concantating multiple strings using '+'.
func BenchmarkConcantate(b *testing.B) {
	previous := "some pretty long name"
	var v string
	for i := 0; i < b.N; i++ {
		v = "some" + previous + "some"
	}
	b.Log(v)
}

// BenchmarkFmt benchmarks concantating multiple strings using 'fmt.Sprintf'.
func BenchmarkFmt(b *testing.B) {
	previous := "some pretty long name"
	var v string
	for i := 0; i < b.N; i++ {
		v = fmt.Sprintf("some%ssome", previous)
	}
	b.Log(v)
}

// BenchmarkBuilder benchmarks concantating multiple strings using 'strings.Builder'.
func BenchmarkBuilder(b *testing.B) {
	previous := "some pretty long name"
	var v string
	for i := 0; i < b.N; i++ {
		bldr := strings.Builder{}
		bldr.WriteString("some")
		bldr.WriteString(previous)
		bldr.WriteString("some")
		v = bldr.String()
	}
	b.Log(v)
}
