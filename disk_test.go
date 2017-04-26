package dkcache

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"sync"
	"testing"
)

func TestDisk0(t *testing.T) {
	var ref PerfectFile
	n, err := ref.Write([]byte("fred"))
	check(err)
	log.Printf("Wrote %d bytes\n", n)
	tv(n, 4, "Write Bytes Expected")
	ref.Seek(0, io.SeekStart)
	tmp_buf := make([]byte, 4, 8)
	n, err = ref.Read(tmp_buf)
	if err != io.EOF {
		check(err)
	}
	log.Printf("Read %d bytes\n", n)
	tv(4, n, "Read Bytes Expected")
}
func TestDisk1(t *testing.T) {
	var ref PerfectFile
	n, err := ref.Write([]byte("fred"))
	check(err)
	log.Printf("Wrote %d bytes\n", n)
	tv(n, 4, "Write Bytes Expected")
	ref.Seek(0, io.SeekStart)
	tmp_buf := make([]byte, 5, 8)
	n, err = ref.Read(tmp_buf)
	if err != io.EOF {
		check(err)
	}
	log.Printf("Read %d bytes\n", n)
	tv(4, n, "Read Bytes Expected")
}
func TestDisk2(t *testing.T) {
	var ref PerfectFile
	n, err := ref.Write([]byte("fred"))
	check(err)
	log.Printf("Wrote %d bytes\n", n)
	tv(n, 4, "Write Bytes Expected")
	ref.Seek(0, io.SeekStart)
	tmp_buf := make([]byte, 3, 8)
	n, err = ref.Read(tmp_buf)
	if err != io.EOF {
		check(err)
	}
	log.Printf("Read %d bytes\n", n)
	tv(3, n, "Read Bytes Expected")
}
func stimLoc(start, end int) int {
	// I would like to call it range, but that is a reserved word
	aga := end - start
	if aga <= 0 {
		log.Fatal("Illegal Stim range")
	}

	return start + rand.Intn(aga)
}
func stimWrite(ref *PerfectFile, can *CachedFile, stim_loc, stim_end int) {
	max_length := stim_end - stim_loc
	expected_length := rand.Intn(max_length)
	if expected_length == 0 {
		expected_length++
	}

	// 3 guaranteed to be random. Selected from fair dice roll
	rs := rand.New(rand.NewSource(3))
	stim_ba := make([]byte, expected_length)
	n, err := rs.Read(stim_ba)
	check(err)
	tv(expected_length, n, "stimWrite")
	n, err = ref.WriteAt(stim_ba, int64(stim_loc))
	check(err)
	tv(expected_length, n, "stimWrite")

	n, err = can.WriteAt(stim_ba, int64(stim_loc))
	check(err)
	tv(expected_length, n, "stimWrite")
}

func stimRead(ref *PerfectFile, can *CachedFile, stim_loc, stim_end int) {
	max_length := stim_end - stim_loc
	expected_length := rand.Intn(max_length)
	if expected_length == 0 {
		expected_length++
	}

	ref_ba := make([]byte, expected_length)
	can_ba := make([]byte, expected_length)

	//log.Printf("Reading %d bytes from %d.\nref is %d,%d\n", expected_length, stim_loc, ref.Cap(), ref.Len())

	n, err := ref.ReadAt(ref_ba, int64(stim_loc))
	if err == io.EOF {
		// if it's the expected length
		if n != expected_length {
			log.Fatal("Length not as expected", n, expected_length)
		}
	} else {
		check(err)
	}
	tv(expected_length, n, "stimRead0")

	n, err = can.ReadAt(can_ba, int64(stim_loc))
	if err == io.EOF {
		// if it's the expected length
		if n != expected_length {
			log.Fatal("Length not as expected")
		}
	} else {
		check(err)
	}
	tv(expected_length, n, "stimRead1")

	if bytes.Compare(ref_ba, can_ba) != 0 {
		log.Fatal("Error array mismatch", ref_ba, can_ba)
	}

}

// A stimmer gets a start and end region of the file to use
// It is expected that we are coherent with both files
// The regions are inclusive
func stimmer(ref *PerfectFile, cand *CachedFile, start, end, txs int, wg *sync.WaitGroup) {
	ref.WriteAt([]byte{0}, int64(end))
	cand.WriteAt([]byte{0}, int64(end))
	//log.Println("Setup buffers to Length of", end, ref.Len(), cand.Len())
	for i := 0; i < txs; i++ {
		stim_loc := stimLoc(start, end)
		if rand.Intn(1) == 1 {
			stimWrite(ref, cand, stim_loc, end)
		} else {
			stimRead(ref, cand, stim_loc, end)
		}
	}
	wg.Done()
}

func TestMulti0(t *testing.T) {
	var ref PerfectFile
	dut := *NewCachedFile(&ref)
	var wg sync.WaitGroup
	wg.Add(1)
	go stimmer(&ref, &dut, 0, 16, 16, &wg)
	wg.Wait()
}
func TestMulti1(t *testing.T) {
	var ref PerfectFile
	dut := *NewCachedFile(&ref)
	var wg sync.WaitGroup
	num_stimmers := 2
	wg.Add(num_stimmers)
	stimmer(&ref, &dut, 0, 15, 512, &wg)
	log.Println("Stim0 Complete")
	stimmer(&ref, &dut, 16, 31, 512, &wg)
	log.Println("Stim1 Complete")
	wg.Wait()
}
func TestMulti2(t *testing.T) {
	var ref PerfectFile
	dut := *NewCachedFile(&ref)
	var wg sync.WaitGroup
	num_stimmers := 20
	wg.Add(num_stimmers)
	gap := 16
	for i := 0; i < num_stimmers; i++ {
		start := gap * i
		end := start + gap - 1
		stimmer(&ref, &dut, start, end, 512, &wg)
	}
	wg.Wait()
}
func TestMulti21(t *testing.T) {
	// Check for disagreements between Teammates
	var ref PerfectFile
	dut := *NewCachedFile(&ref)
	var wg sync.WaitGroup
	num_stimmers := 20
	wg.Add(num_stimmers)
	gap := 16
	for i := 0; i < num_stimmers; i++ {
		start := gap * i
		end := start + gap - 1
		go stimmer(&ref, &dut, start, end, 512, &wg)
	}
	wg.Wait()
}
