package contentstore

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"

	"github.com/kjk/u"
)

const (
	SEGMENT_MAX_SIZE = 100 * 1024       // 100 kB
	SMALL_BLOB_MAX   = 10 * 1024        // 10 kB
	BIG_BLOB_MAX     = 200 * 1024       // 200 kB
	MAX_TO_WRITE     = 20 * 1024 * 1024 // 20 MB
	//MAX_TO_WRITE = 20 * 1024 // 20 kB
)

func removeStoreFiles(basePath string) {
	path := idxFilePath(basePath)
	os.Remove(path)
	nSegment := 0
	for {
		path = segmentFilePath(basePath, nSegment)
		if !u.PathExists(path) {
			break
		}
		os.Remove(path)
		nSegment += 1
	}
}

func genRandBytes(rnd *rand.Rand, n int) []byte {
	res := make([]byte, n, n)
	for i := 0; i < n; i++ {
		res[i] = byte(rnd.Intn(256))
	}
	return res
}

func populate(t *testing.T, store *Store, totalWritten *int32, maxToWrite int32) []string {
	rnd := rand.New(rand.NewSource(0))
	blobIds := make([]string, 0)
	for {
		var d []byte
		isBig := (rnd.Intn(10) == 0) // 10% of the time generate big blob
		if isBig {
			d = genRandBytes(rnd, rnd.Intn(BIG_BLOB_MAX))
		} else {
			d = genRandBytes(rnd, rnd.Intn(SMALL_BLOB_MAX))
		}
		id, err := store.Put(d)
		if err != nil {
			t.Fatalf("store.Put() failed with %q", err)
		}
		blobIds = append(blobIds, id)
		total := atomic.AddInt32(totalWritten, int32(len(d)))
		if total > maxToWrite {
			break
		}
		// TODO: test that created segments as expected
		//fmt.Printf("nWritten: %d\n", nWritten)
	}
	return blobIds
}

func testGet(t *testing.T, store *Store, rnd *rand.Rand, blobIds []string) {
	var d []byte
	var err error
	nBlobs := len(blobIds)
	for i := 0; i < nBlobs; i++ {
		// we don't test all blobs due to randomness but testing most of them
		// is good enough
		n := rnd.Intn(nBlobs)
		id := blobIds[n]
		d, err = store.Get(id)
		if err != nil {
			t.Fatalf("store.Get(%q) failed with %q, i: %d, sha1: %s", id, err, i, hex.EncodeToString([]byte(id)))
		}
		sha1Hex := fmt.Sprintf("%x", u.Sha1OfBytes(d))
		if sha1Hex != id {
			t.Fatalf("store.Get() returned bad content, id is %s while sha1 is %s, should be same", id, sha1Hex)
		}
	}
	k := "non-existint"
	d, err = store.Get(k)
	if err == nil {
		t.Fatalf("store.Get(%q) returned nil err, expected an error", k)
	}
}

func TestStore(t *testing.T) {
	basePath := "test"
	removeStoreFiles(basePath)
	store, err := NewWithLimit(basePath, SEGMENT_MAX_SIZE)
	if err != nil {
		t.Fatalf("NewWithLimit(%q, %d) failed with %q", basePath, SEGMENT_MAX_SIZE, err)
	}
	defer func() {
		if store != nil {
			store.Close()
		}
	}()
	nWorkers := 5
	resChan := make(chan []string)
	var totalWritten int32 = 0
	for i := 0; i < nWorkers; i++ {
		go func(c chan []string) {
			res := populate(t, store, &totalWritten, MAX_TO_WRITE)
			c <- res
		}(resChan)
	}
	blobIds := make([]string, 0)
	for i := 0; i < nWorkers; i++ {
		res := <-resChan
		//fmt.Printf("worker %d created %d strings\n", i, len(res))
		blobIds = append(blobIds, res...)
	}
	//fmt.Printf("total strings: %d\n", len(blobIds))
	rnd := rand.New(rand.NewSource(0))
	testGet(t, store, rnd, blobIds)
	store.Close()
	store = nil
	store, err = NewWithLimit(basePath, SEGMENT_MAX_SIZE)
	if err != nil {
		t.Fatalf("NewWithLimit(%q, %d) failed with %q", basePath, SEGMENT_MAX_SIZE, err)
	}
	testGet(t, store, rnd, blobIds)
	// TODO: add new items
	store.Close()
	store = nil
	removeStoreFiles(basePath)
}
