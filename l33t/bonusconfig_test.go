package l33t

import (
	"strings"
	"testing"
)

const (
	endMatch = `00000001337`
)

// func Benchmark_BonusConfig_hasHomogenicPrefix(b *testing.B) {
// 	bc := BonusConfig{
// 		SubString:  "1337",
// 		PrefixChar: 48,
// 		Greeting:   "The ultimate goal has been reached!",
// 	}
// 	bc.matchPos = strings.Index(endMatch, bc.SubString)
//
// 	for i := 0; i < b.N; i++ {
// 		if !bc.hasHomogenicPrefix(endMatch) {
// 			b.Fatal("Got unexpected result: false")
// 		}
// 	}
// }

// Turns out there's no real performance diff between these two variants, either as a method
// with a value receiver, or as a free function, only passing the params needed.
// Both take about 8.8 ns/op and have 0 allocs, where the free function is ~0.003 ns faster.

func Benchmark_hasHomogenicPrefix(b *testing.B) {
	matchPos := strings.Index(endMatch, "1337")
	for i := 0; i < b.N; i++ {
		if !hasHomogenicPrefix(endMatch, 48, matchPos) {
			b.Fatal("Got unexpected result: false")
		}
	}
}
