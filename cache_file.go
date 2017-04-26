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
// That is many procs can access it as if it were a single file
// We take care of appropriate locking
type CachedFile struct {
	// Our rule is that all externall accessable functions must
	// obtain the appropriate locks
	// Internal functions must not.
	sync.RWMutex
	store []byte
	loc   int64
}

func (cf *CachedFile) Grow(desired_size int64) {
	cf.Lock()
	defer cf.Unlock()

	cf.grow(desired_size)
}
func (cf *CachedFile) grow(desired_size int64) {
	log.Printf("Current Cap is %d, Asked for %d\n", cap(cf.store), desired_size)
	// First of all round up to a power of 2
	desired_size = int64(roundup2(desired_size))
	if desired_size > int64(cap(cf.store)) {
		new_store := make([]byte, len(cf.store), desired_size)
		copy(new_store, cf.store)
		cf.store = new_store
	}
}

func (cf *CachedFile) Seek(offset int64, whence int) (int64, error) {
	cf.Lock()
	defer cf.Unlock()

	return cf.seek(offset, whence)
}
func (cf *CachedFile) seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		cf.loc = offset
	} else if whence == io.SeekCurrent {
		cf.loc = offset + cf.loc
	} else if whence == io.SeekEnd {
		cf.loc = int64(len(cf.store)) + offset
	} else {
		return 0, ErrUnknownSeek
	}
	return cf.loc, nil
}

// writeAt relies on a sufficiently sized cf
// call WriteAt if you're not sure the buffer is in a good state
func (cf *CachedFile) writeAtWk(p []byte, offset int64) (n int, err error) {
	start_len := int64(len(cf.store))
	end_loc := offset + int64(len(p))
	if end_loc > int64(cap(cf.store)) {
		end_loc = int64(cap(cf.store))
	}
	n = copy(cf.store[offset:end_loc], p)
	if end_loc > start_len {
		cf.store = cf.store[:end_loc]
	}
	return n, nil
}
func (cf *CachedFile) WriteAt(p []byte, offset int64) (n int, err error) {
	cf.Lock()
	defer cf.Unlock()
	return cf.writeAt(p, offset)
}
func (cf *CachedFile) writeAt(p []byte, offset int64) (n int, err error) {

	if cf.store == nil {
		cf.store = make([]byte, 0, InitialBufferSize)
		//log.Println("Made Array")
		cf.loc = 0
	}
	//blen := len(cf.store)
	bcap := int64(cap(cf.store))
	plen := len(p)
	final_len := offset + int64(plen)

	if final_len > bcap {
		log.Println("Growing")
		cf.grow(final_len)
		bcap = int64(cap(cf.store))
	}
	remaining := plen
	for remaining != 0 {
		//log.Println("Remaining", remaining)
		nc, err := cf.writeAtWk(p, offset)
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

func (cf *CachedFile) Write(p []byte) (n int, err error) {
	cf.Lock()
	n, err = cf.writeAt(p, cf.loc)
	cf.loc += int64(n)
	cf.Unlock()
	return
}
func (cf *CachedFile) ReadAt(p []byte, offset int64) (n int, err error) {
	cf.Lock()
	if cf.store == nil {
		cf.Unlock()
		return 0, io.EOF
	}
	if p == nil {
		log.Fatal("WTF")
	} else {
		//log.Printf("It's %d,%d\n", len(p), cap(p))
	}
	n, err = cf.readAt(p, offset)
	cf.Unlock()
	return
}
func (cf *CachedFile) readAt(p []byte, offset int64) (n int, err error) {
	//log.Printf("Reading into: %d,%d\n", len(p), cap(p))
	blen := int64(len(cf.store))

	//end_location := offset + int64(len(p))
	//if end_location > blen {
	//	end_location = blen
	//}
	if offset >= blen {
		log.Fatal("Asked for beyond end of buffer", offset, blen, cap(cf.store))
	}
	n = copy(p, cf.store[offset:])
	if (offset + int64(n)) >= blen {
		err = io.EOF
	}
	return
}
func (cf *CachedFile) Read(p []byte) (n int, err error) {
	if cf.store == nil {
		return 0, io.EOF
	}
	cf.Lock()
	start_location := cf.loc
	n, err = cf.readAt(p, start_location)
	cf.loc += int64(n)
	cf.Unlock()

	return
}
func (cf *CachedFile) Close() (err error) {
	return nil
}
func (cf *CachedFile) Cap() int {
	return cap(cf.store)
}
func (cf *CachedFile) Len() int {
	return len(cf.store)
}
