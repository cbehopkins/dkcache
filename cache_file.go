package dkcache

import (
	"io"
	"log"

	"sync"
)

// A Cached file behaves like a perfect file sort of
// It writes Instantly (for large values of instant)
// Ir Reads Instantly if it hits
// It does not error (system memory permitting)
// It supports concurrent access (for appropriate methods)
// That is Many procs can access it as it it were a single file
// We take care of appropriate locking
type CachedFile struct {
	// Our rule is that all externall accessable functions must
	// obtain the appropriate locks
	// Internal functions must not.
	sync.RWMutex
	store []byte
	loc   int64
}

func (pf *CachedFile) Grow(desired_size int64) {
	pf.Lock()
	defer pf.Unlock()

	pf.grow(desired_size)
}
func (pf *CachedFile) grow(desired_size int64) {
	log.Printf("Current Cap is %d, Asked for %d\n", cap(pf.store), desired_size)
	// First of all round up to a power of 2
	desired_size = int64(roundup2(desired_size))
	if desired_size > int64(cap(pf.store)) {
		new_store := make([]byte, len(pf.store), desired_size)
		copy(new_store, pf.store)
		pf.store = new_store
	}
}

func (pf *CachedFile) Seek(offset int64, whence int) (int64, error) {
	pf.Lock()
	defer pf.Unlock()

	return pf.seek(offset, whence)
}
func (pf *CachedFile) seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		pf.loc = offset
	} else if whence == io.SeekCurrent {
		pf.loc = offset + pf.loc
	} else if whence == io.SeekEnd {
		pf.loc = int64(len(pf.store)) + offset
	} else {
		return 0, ErrUnknownSeek
	}
	return pf.loc, nil
}

// writeAt relies on a sufficiently sized pf
// call WriteAt if you're not sure the buffer is in a good state
func (pf *CachedFile) writeAtWk(p []byte, offset int64) (n int, err error) {
	start_len := int64(len(pf.store))
	end_loc := offset + int64(len(p))
	if end_loc > int64(cap(pf.store)) {
		end_loc = int64(cap(pf.store))
	}
	n = copy(pf.store[offset:end_loc], p)
	if end_loc > start_len {
		pf.store = pf.store[:end_loc]
	}
	return n, nil
}
func (pf *CachedFile) WriteAt(p []byte, offset int64) (n int, err error) {
	pf.Lock()
	defer pf.Unlock()
	return pf.writeAt(p, offset)
}
func (pf *CachedFile) writeAt(p []byte, offset int64) (n int, err error) {

	if pf.store == nil {
		pf.store = make([]byte, 0, InitialBufferSize)
		//log.Println("Made Array")
		pf.loc = 0
	}
	//blen := len(pf.store)
	bcap := int64(cap(pf.store))
	plen := len(p)
	final_len := offset + int64(plen)

	if final_len > bcap {
		log.Println("Growing")
		pf.grow(final_len)
		bcap = int64(cap(pf.store))
	}
	remaining := plen
	for remaining != 0 {
		//log.Println("Remaining", remaining)
		nc, err := pf.writeAtWk(p, offset)
		check(err)
		// reslice p down to the remaining
		p = p[nc:]
		//log.Println("Resliced", nc, p)
		offset += int64(nc)
		remaining -= nc
		n += nc
	}
	return
}

func (pf *CachedFile) Write(p []byte) (n int, err error) {
	pf.Lock()
	n, err = pf.writeAt(p, pf.loc)
	pf.loc += int64(n)
	pf.Unlock()
	return
}
func (pf *CachedFile) ReadAt(p []byte, offset int64) (n int, err error) {
	pf.Lock()
	if pf.store == nil {
		pf.Unlock()
		return 0, io.EOF
	}
	if p == nil {
		log.Fatal("WTF")
	} else {
		//log.Printf("It's %d,%d\n", len(p), cap(p))
	}
	n, err = pf.readAt(p, offset)
	pf.Unlock()
	return
}
func (pf *CachedFile) readAt(p []byte, offset int64) (n int, err error) {
	//log.Printf("Reading into: %d,%d\n", len(p), cap(p))
	blen := int64(len(pf.store))

	//end_location := offset + int64(len(p))
	//if end_location > blen {
	//	end_location = blen
	//}
	if offset >= blen {
		log.Fatal("Asked for beyond end of buffer", offset, blen, cap(pf.store))
	}
	n = copy(p, pf.store[offset:])
	if (offset + int64(n)) >= blen {
		err = io.EOF
	}
	return
}
func (pf *CachedFile) Read(p []byte) (n int, err error) {
	if pf.store == nil {
		return 0, io.EOF
	}
	pf.Lock()
	start_location := pf.loc
	n, err = pf.readAt(p, start_location)
	pf.loc += int64(n)
	pf.Unlock()

	return
}
func (pf *CachedFile) Close() (err error) {
	return nil
}
func (pf *CachedFile) Cap() int {
	return cap(pf.store)
}
func (pf *CachedFile) Len() int {
	return len(pf.store)
}
