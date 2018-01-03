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
func NewCachedFile (f RWAtSeekCloser) *CachedFile {
  itm := new(CachedFile)
  itm.f = f
  return itm
}
func (cf *CachedFile) Seek(offset int64, whence int) (int64, error) {
	cf.Lock()
	defer cf.Unlock()

	return cf.f.Seek(offset, whence)
}
func (cf *CachedFile) Sync () () {
}
func (cf *CachedFile) WriteAt(p []byte, offset int64) (n int, err error) {
	cf.Lock()
	defer cf.Unlock()
	return cf.writeAt(p, offset)
}

func (cf *CachedFile) writeAt(p []byte, offset int64) (n int, err error) {
  return cf.writeFile(p,offset)
}
func (cf *CachedFile) writeFile(p []byte, offset int64) (n int, err error) {
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
			return n,err
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

// A Transaction to a Cache line may not be to the whole line
// So we need to 
type Trans struct {
  ln Line
  off int
  len int
  // A Read request will set this to the slice to modify
  // A write request will have the data here
  data []byte
  // Execute Done once the data are either swallowed or populated
  wg *sync.WaitGroup
}
// Fragment a transaction into lots of little ones
// 
func fragmentTx (offset int64, length int) (transaction []Trans) {
}
type Line int64
const (
  // 16 Bytes in each line
  LineLength = 16
)
func (cf *CachedFile) checkCache (off Line ) bool {
  // Returns true if this Line is in the cache
  return false
}
// Translate from the Line number into the byte offset
func (cf *CachedFile) lineNumber (ln Line) (offset int64) {

}

func (cf *CachedFile) readAt(p []byte, offset int64) (n int, err error) {
  // Read at needs to break the transaction up into individual fragments
  // Check if any of these match anything in the cache
  // Then issue the reads as needed
  // For performance we try and issue a few big reads 
  // in preference to lots of little ones
  transactions := fragmentTx(offset, len(p))
  num_transactions := len(transactions)
  var wg sync.WaitGroup
  wg.Add(num_transactions)
  for tx := range transactions {
    tmp_trans := Trans{
                      ln:tx.ln
                      off:tx.off
                      len:tx.len
                      data:p[]
                      wg:&wg
                      }
    if cf.checkCache(tx.ln) {
      // Retrieve the data from the cache
      // Not that the cache line may itself not be full
      // so that may have to do a fetch under the hood
    } else {
      // Fetch the data from the disk
      ln := cf.lineNumber(off)
      nc, err := cf.readFile()
      if nc != LineLength
    }
  }

  return cf.readFile(p,offset)
}
func (cf *CachedFile) readFile(p []byte, offset int64) (n int, err error) {
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
