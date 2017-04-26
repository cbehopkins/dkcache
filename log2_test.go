package dkcache

import (
	"log"
	"testing"
)

func tv(expected, result int, test string) {
	if expected != result {
		log.Fatalf("Error %s:expected was %d, result was %d\n", test, expected, result)
	}
}
func TestLg2(t *testing.T) {
	tv(log2(0), 0, "log2")
	tv(log2(1), 1, "log2")
	tv(log2(2), 2, "log2")
	tv(log2(3), 2, "log2")
	tv(log2(4), 3, "log2")
	tv(log2(6), 3, "log2")
	tv(log2(8), 4, "log2")
	tv(log2(16), 5, "log2")
	tv(log2(31), 5, "log2")
}
