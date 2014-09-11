package contentstore

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/kjk/u"
)

// Files we use:
// - index where for each blob we store: sha1 of the content, number of segment
//   file in which the blob is stored, offset within the segment and size of
//   the blob
// - one or more segment files. User can control the max size of segment
//   file (10 MB by default) to pick the right file size/number of files
//   balance for his needs

// Ideas for the future:
// - add a mode where we store big blobs (e.g. over 1MB) in their own files
// - add a way to delete files and re-use deleted space via some sort of
//   best-fit allocator (if we're adding a new blob and have free space
//   due to deletion, pick the free space that is the closest in size to
//   new blob but only if it's e.g. withing 10% in size, to avoid excessive
//   fragmentation)
// - add a way to delete files by rewriting the files (expensive! we have to
//   rewrite the whole index and each segment that contains deleted files)

var (
	errNotFound           = errors.New("not found")
	errInvalidIndexHdr    = errors.New("invalid index file header")
	errInvalidIndexLine   = errors.New("invalid index line")
	errSegmentFileMissing = errors.New("segment file missing")
	errNotValidSha1       = errors.New("not a valid sha1")
	// first line in index file, for additional safety
	idxHdr = "github.com/kjk/contentstore header 1.0"
)

type blob struct {
	sha1     [20]byte
	nSegment int
	offset   int
	size     int
}

type Store struct {
	sync.Mutex
	basePath       string
	maxSegmentSize int
	blobs          []blob
	// sha1ToBlob is to quickly find a message based on sha1
	// string is really [20]byte cast to string and int is a position within blobs array
	sha1HexToBlobNo map[string]int
	idxFile         *os.File
	idxCsvWriter    *csv.Writer
	currSegmentFile *os.File
	currSegmentNo   int
	currSegmentSize int
	// we cache file descriptor for one segment file (in addition to current
	// segment file) to reduce file open/close for Get()
	cachedSegmentFile *os.File
	cachedSegmentNo   int
}

func idxFilePath(basePath string) string {
	return basePath + "_idx.txt"
}

func segmentFilePath(basePath string, nSegment int) string {
	return fmt.Sprintf("%s_%d.txt", basePath, nSegment)
}

func decodeIndexLine(rec []string) (blob blob, err error) {
	if len(rec) != 4 {
		return blob, errInvalidIndexLine
	}
	sha1, err := hex.DecodeString(rec[0])
	if err != nil {
		return blob, err
	}
	if len(sha1) != 20 {
		return blob, errNotValidSha1
	}
	copy(blob.sha1[:], sha1)
	if blob.nSegment, err = strconv.Atoi(rec[1]); err != nil {
		return blob, err
	}
	if blob.offset, err = strconv.Atoi(rec[2]); err != nil {
		return blob, err
	}
	if blob.size, err = strconv.Atoi(rec[3]); err != nil {
		return blob, err
	}
	return blob, nil
}

// appends x to array of ints
func appendIntIfNotExists(aPtr *[]int, x int) {
	a := *aPtr
	if sort.SearchInts(a, x) >= 0 {
		return
	}
	a = append(a, x)
	sort.Ints(a)
	*aPtr = a
}

// TODO: error out if already in sha1HexToBlobNo
func (store *Store) appendBlob(blob blob) {
	sha1 := fmt.Sprintf("%x", blob.sha1[:])
	blobNo := len(store.blobs)
	store.blobs = append(store.blobs, blob)
	store.sha1HexToBlobNo[sha1] = blobNo
}

func (store *Store) readIndex() error {
	// at this point idx file must exist
	file, err := os.Open(idxFilePath(store.basePath))
	if err != nil {
		return err
	}
	defer file.Close()
	// TODO: would be faster (and easier?) to use a bitset since we know
	// segment numbers are consequitive integers
	segments := make([]int, 0)
	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	csvReader.FieldsPerRecord = -1
	rec, err := csvReader.Read()
	if err != nil || len(rec) != 1 || rec[0] != idxHdr {
		return errInvalidIndexHdr
	}
	var blob blob
	for {
		if rec, err = csvReader.Read(); err != nil {
			break
		}
		if blob, err = decodeIndexLine(rec); err != nil {
			break
		}
		appendIntIfNotExists(&segments, blob.nSegment)
		store.appendBlob(blob)
	}
	if err == io.EOF {
		err = nil
	}
	// verify segment files exist
	// TODO: also verify offset + size is <= size of segment file
	for _, nSegment := range segments {
		path := segmentFilePath(store.basePath, nSegment)
		if !u.PathExists(path) {
			return errSegmentFileMissing
		}
		store.currSegmentNo = nSegment
	}
	return nil
}

func NewWithLimit(basePath string, maxSegmentSize int) (store *Store, err error) {
	store = &Store{
		basePath:        basePath,
		blobs:           make([]blob, 0),
		sha1HexToBlobNo: make(map[string]int),
		maxSegmentSize:  maxSegmentSize,
		cachedSegmentNo: -1,
	}
	idxPath := idxFilePath(basePath)
	idxDidExist := u.PathExists(idxPath)
	if idxDidExist {
		if err = store.readIndex(); err != nil {
			return nil, err
		}
	}
	if store.idxFile, err = os.OpenFile(idxPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err != nil {
		return nil, err
	}
	store.idxCsvWriter = csv.NewWriter(store.idxFile)
	if !idxDidExist {
		rec := [][]string{{idxHdr}}
		err = store.idxCsvWriter.WriteAll(rec)
		if err != nil {
			return nil, err
		}
	}
	segmentPath := segmentFilePath(store.basePath, store.currSegmentNo)
	stat, err := os.Stat(segmentPath)
	if err != nil {
		// TODO: fail if error is different than "file doesn't exist"
		if store.currSegmentNo != 0 {
			store.Close()
			return nil, errSegmentFileMissing
		}
		store.currSegmentFile, err = os.Create(segmentPath)
		if err != nil {
			store.Close()
			return nil, err
		}
	} else {
		store.currSegmentSize = int(stat.Size())
		store.currSegmentFile, err = os.OpenFile(segmentPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			store.Close()
			return nil, err
		}
	}
	return store, nil
}

func New(basePath string) (*Store, error) {
	return NewWithLimit(basePath, 10*1024*1024)
}

func closeFilePtr(filePtr **os.File) (err error) {
	f := *filePtr
	if f != nil {
		err = f.Close()
		*filePtr = nil
	}
	return err
}

func (store *Store) Close() {
	closeFilePtr(&store.idxFile)
	closeFilePtr(&store.currSegmentFile)
	closeFilePtr(&store.cachedSegmentFile)
}

func writeBlobRec(csvWriter *csv.Writer, blob *blob) error {
	sha1Str := hex.EncodeToString(blob.sha1[:])
	rec := [][]string{
		{
			sha1Str,
			strconv.Itoa(blob.nSegment),
			strconv.Itoa(blob.offset),
			strconv.Itoa(blob.size),
		}}
	return csvWriter.WriteAll(rec)
}

func (store *Store) Put(d []byte) (id string, err error) {
	store.Lock()
	defer store.Unlock()

	idBytes := u.Sha1OfBytes(d)
	id = fmt.Sprintf("%x", idBytes)
	if _, ok := store.sha1HexToBlobNo[id]; ok {
		return id, nil
	}
	blob := blob{
		size:     len(d),
		offset:   store.currSegmentSize,
		nSegment: store.currSegmentNo,
	}
	copy(blob.sha1[:], idBytes)

	if _, err = store.currSegmentFile.Write(d); err != nil {
		return "", err
	}
	if err = store.currSegmentFile.Sync(); err != nil {
		return "", err
	}
	store.currSegmentSize += blob.size
	if store.currSegmentSize >= store.maxSegmentSize {
		// filled current segment => create a new one
		err = store.currSegmentFile.Close()
		if err != nil {
			return "", err
		}
		store.currSegmentNo += 1
		store.currSegmentSize = 0
		path := segmentFilePath(store.basePath, store.currSegmentNo)
		store.currSegmentFile, err = os.Create(path)
		if err != nil {
			return "", err
		}
	}
	if err = writeBlobRec(store.idxCsvWriter, &blob); err != nil {
		return "", err
	}
	store.appendBlob(blob)
	return id, nil
}

func readFromFile(file *os.File, offset, size int) ([]byte, error) {
	res := make([]byte, size, size)
	if _, err := file.ReadAt(res, int64(offset)); err != nil {
		return nil, err
	}
	return res, nil
}

func (store *Store) getSegmentFile(nSegment int) (*os.File, error) {
	if nSegment == store.currSegmentNo {
		return store.currSegmentFile, nil
	}
	if nSegment == store.cachedSegmentNo {
		return store.cachedSegmentFile, nil
	}
	closeFilePtr(&store.cachedSegmentFile)
	path := segmentFilePath(store.basePath, nSegment)
	var err error
	if store.cachedSegmentFile, err = os.Open(path); err != nil {
		return nil, err
	}
	store.cachedSegmentNo = nSegment
	return store.cachedSegmentFile, nil
}

func (store *Store) Get(id string) ([]byte, error) {
	store.Lock()
	defer store.Unlock()

	blobNo, ok := store.sha1HexToBlobNo[id]
	if !ok {
		return nil, errNotFound
	}
	blob := store.blobs[blobNo]
	segmentFile, err := store.getSegmentFile(blob.nSegment)
	if err != nil {
		return nil, err
	}
	return readFromFile(segmentFile, blob.offset, blob.size)
}
