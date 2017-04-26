package dkcache

import "errors"

func check(err error) {
	switch err {
	case nil:
		return
	//case io.EOF:
	//	return
	default:
		panic(err)
	}
}

// Start off as a 1k buffer
const (
	InitialBufferSize = 1 << 10
)

var ErrUnknownSeek = errors.New("Unknown Seek Request")

func log2(in int64) (lg2 int) {
	for lg2 = 0; in > 0; in = in >> 1 {
		lg2++
	}
	return
}
func roundup2(in int64) (out int) {
	// Roundup to the next power of 2 size
	bits := log2(in)
	out = 1 << (uint(bits))
	return
}
