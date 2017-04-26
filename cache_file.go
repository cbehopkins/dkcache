package dkcache

import "sync"

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
	f   RWAtSeekCloser
	loc int64
}

func NewCachedFile(f RWAtSeekCloser) *CachedFile {
	itm := new(CachedFile)
	itm.f = f
	return itm
}
func (cf *CachedFile) Seek(offset int64, whence int) (int64, error) {
	cf.Lock()
	defer cf.Unlock()

	return cf.f.Seek(offset, whence)
}
func (cf *CachedFile) WriteAt(p []byte, offset int64) (n int, err error) {
	cf.Lock()
	defer cf.Unlock()
	return cf.writeAt(p, offset)
}
func (cf *CachedFile) writeAt(p []byte, offset int64) (n int, err error) {

	remaining := len(p)
	for remaining != 0 {
		//log.Println("Remaining", remaining)
		nc, err := cf.f.WriteAt(p, offset)
		// reslice p down to the remaining
		p = p[nc:]
		//log.Println("Resliced", nc, p)
		offset += int64(nc)
		remaining -= nc
		n += nc
		if err != nil {
			return n, err
		}
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
	n, err = cf.readAt(p, offset)
	cf.Unlock()
	return
}
func (cf *CachedFile) readAt(p []byte, offset int64) (n int, err error) {
	n, err = cf.f.ReadAt(p, offset)
	return
}
func (cf *CachedFile) Read(p []byte) (n int, err error) {
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
