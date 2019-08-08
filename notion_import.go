package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kjk/notionapi"
)

var (
	// if true, we'll log
	logNotionRequests = true

	cacheDir     = "notion_cache"
	notionLogDir = "log"
)

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
func normalizeID(s string) string {
	return notionapi.ToNoDashID(s)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	if !logNotionRequests {
		return nil, nil
	}

	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join(notionLogDir, name)
	f, err := os.Create(path)
	if err != nil {
		lg("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func loadPageFromCache(dir, pageID string) *notionapi.Page {
	cachedPath := filepath.Join(dir, pageID+".json")
	d, err := ioutil.ReadFile(cachedPath)
	if err != nil {
		return nil
	}

	var page notionapi.Page
	err = json.Unmarshal(d, &page)
	panicIfErr(err)
	return &page
}

// I got "connection reset by peer" error once so retry download 3 times, with a short sleep in-between
func downloadPageRetry(c *notionapi.Client, pageID string) (*notionapi.Page, error) {
	var res *notionapi.Page
	var err error
	for i := 0; i < 3; i++ {
		if i > 0 {
			lg("Download %s failed with '%s'\n", pageID, err)
			time.Sleep(3 * time.Second) // not sure if it matters
		}
		res, err = c.DownloadPage(pageID)
		if err == nil {
			return res, nil
		}
	}
	return nil, err
}

func sha1OfLink(link string) string {
	link = strings.ToLower(link)
	h := sha1.New()
	h.Write([]byte(link))
	return fmt.Sprintf("%x", h.Sum(nil))
}

var imgFiles []os.FileInfo

func findImageInDir(imgDir string, sha1 string) string {
	if len(imgFiles) == 0 {
		imgFiles, _ = ioutil.ReadDir(imgDir)
	}
	for _, fi := range imgFiles {
		if strings.HasPrefix(fi.Name(), sha1) {
			return filepath.Join(imgDir, fi.Name())
		}
	}
	return ""
}

func guessExt(fileName string, contentType string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".png", ".jpg", ".jpeg":
		return ext
	}
	switch contentType {
	case "image/png":
		return ".png"
	}
	panic(fmt.Errorf("Didn't find ext for file '%s', content type '%s'", fileName, contentType))
}

func downloadImage(c *notionapi.Client, uri string) ([]byte, string, error) {
	img, err := c.DownloadFile(uri)
	if err != nil {
		lg("\n  failed with %s\n", err)
		return nil, "", err
	}
	ext := guessExt(uri, img.Header.Get("Content-Type"))
	return img.Data, ext, nil
}

// return path of cached image on disk
func downloadAndCacheImage(c *notionapi.Client, uri string) (string, error) {
	sha := sha1OfLink(uri)

	//ext := strings.ToLower(filepath.Ext(uri))

	imgDir := filepath.Join(cacheDir, "img")
	err := os.MkdirAll(imgDir, 0755)
	panicIfErr(err)

	cachedPath := findImageInDir(imgDir, sha)
	if cachedPath != "" {
		verbose("Image %s already downloaded as %s\n", uri, cachedPath)
		return cachedPath, nil
	}

	timeStart := time.Now()
	lg("Downloading %s ... ", uri)

	imgData, ext, err := downloadImage(c, uri)

	cachedPath = filepath.Join(imgDir, sha+ext)

	err = ioutil.WriteFile(cachedPath, imgData, 0644)
	if err != nil {
		return "", err
	}
	lg("finished in %s. Wrote as '%s'\n", time.Since(timeStart), cachedPath)

	return cachedPath, nil
}

func downloadAndCachePage(c *notionapi.Client, pageID string) (*notionapi.Page, error) {
	//verbose("downloading page with id %s\n", pageID)
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		c.Logger = lf
		defer lf.Close()
	}
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	page, err := downloadPageRetry(c, pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(page, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		panicIfErr(err)
	} else {
		// not a fatal error, just a warning
		lg("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return page, nil
}

func pageIDFromFileName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return ""
	}
	id := parts[0]
	if len(id) == len("2b831bac5afc414493cff5e06e8e4460") {
		return id
	}
	return ""
}

func loadPagesFromDisk(dir string) map[string]*notionapi.Page {
	cachedPagesFromDisk := map[string]*notionapi.Page{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		lg("loadPagesFromDisk: os.ReadDir('%s') failed with '%s'\n", dir, err)
		return cachedPagesFromDisk
	}
	for _, f := range files {
		pageID := pageIDFromFileName(f.Name())
		if pageID == "" {
			continue
		}
		page := loadPageFromCache(dir, pageID)
		cachedPagesFromDisk[pageID] = page
	}
	lg("loadPagesFromDisk: loaded %d cached pages from %s\n", len(cachedPagesFromDisk), dir)
	return cachedPagesFromDisk
}

func loadNotionPage(c *notionapi.Client, pageID string, n int, isCachedPageNotOutdated map[string]bool, cachedPagesFromDisk map[string]*notionapi.Page) (*notionapi.Page, error) {
	if isCachedPageNotOutdated[pageID] {
		page := cachedPagesFromDisk[pageID]
		//nTotalFromCache++
		title := page.Root().Title
		verbose("Page %4d %s: skipping (ver not changed), title: %s\n", n, page.ID, title)
		return page, nil
	}

	page, err := downloadAndCachePage(c, pageID)
	if err != nil {
		return nil, err
	}
	lg("Page %4d %s: downloaded. Title: %s\n", n, page.ID, page.Root().Title)
	return page, nil
}

func isIDEqual(id1, id2 string) bool {
	return notionapi.ToNoDashID(id1) == notionapi.ToNoDashID(id2)
}

func getVersionsForPages(c *notionapi.Client, ids []string) ([]int64, error) {
	// c.Logger = os.Stdout
	recVals, err := c.GetRecordValues(ids)
	if err != nil {
		return nil, err
	}
	results := recVals.Results
	if len(results) != len(ids) {
		return nil, fmt.Errorf("getVersionssForPages(): got %d results, expected %d", len(results), len(ids))
	}
	var versions []int64
	for i, res := range results {
		// res.Value might be nil when a page is not publicly visible or was deleted
		if res.Value == nil {
			versions = append(versions, 0)
			continue
		}
		id := res.Value.ID
		panicIf(!isIDEqual(ids[i], id), "got result in the wrong order, ids[i]: %s, id: %s", ids[0], id)
		versions = append(versions, res.Value.Version)
	}
	return versions, nil
}

func checkIfPagesAreOutdated(c *notionapi.Client, cachedPagesFromDisk map[string]*notionapi.Page) map[string]bool {
	isCachedPageNotOutdated := map[string]bool{}
	var ids []string
	for id := range cachedPagesFromDisk {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var versions []int64
	rest := ids
	maxPerCall := 256
	for len(rest) > 0 {
		n := len(rest)
		if n > maxPerCall {
			n = maxPerCall
		}
		tmpIDs := rest[:n]
		rest = rest[n:]
		lg("getting versions for %d pages\n", len(tmpIDs))
		tmpVers, err := getVersionsForPages(c, tmpIDs)
		panicIfErr(err)
		versions = append(versions, tmpVers...)
	}
	panicIf(len(ids) != len(versions))
	nOutdated := 0
	for i, ver := range versions {
		id := ids[i]
		page := cachedPagesFromDisk[id]
		isOutdated := ver > page.Root().Version
		isCachedPageNotOutdated[id] = !isOutdated
		if isOutdated {
			nOutdated++
		}
	}
	lg("checkIfPagesAreOutdated: %d pages, %d outdated\n", len(ids), nOutdated)
	return isCachedPageNotOutdated
}

// returns true if did build
func maybeBuildIDToPageMap(cachedPagesFromDisk map[string]*notionapi.Page, idToPage map[string]*notionapi.Page) bool {
	if !flgNoDownload {
		return false
	}
	if len(cachedPagesFromDisk) == 0 {
		fmt.Printf("ignoring flgNoDownload=%v because no cached pages\n", flgNoDownload)
		return false
	}
	for _, page := range cachedPagesFromDisk {
		id := page.ID
		id = normalizeID(id)
		idToPage[id] = page
	}
	return true
}

func loadNotionPages(c *notionapi.Client, indexPageID string, idToPage map[string]*notionapi.Page) {
	cachedPagesFromDisk := loadPagesFromDisk(cacheDir)

	if maybeBuildIDToPageMap(cachedPagesFromDisk, idToPage) {
		return
	}

	isCachedPageNotOutdated := checkIfPagesAreOutdated(c, cachedPagesFromDisk)

	toVisit := []string{indexPageID}

	n := 1
	for len(toVisit) > 0 {
		pageID := normalizeID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := idToPage[pageID]; ok {
			continue
		}

		page, err := loadNotionPage(c, pageID, n, isCachedPageNotOutdated, cachedPagesFromDisk)
		panicIfErr(err)
		n++

		idToPage[pageID] = page

		subPages := notionapi.GetSubPages(page.Root().Content)
		toVisit = append(toVisit, subPages...)
	}
}

func loadAllPages(c *notionapi.Client, startIDs []string) map[string]*notionapi.Page {
	idToPage := map[string]*notionapi.Page{}
	nPrev := 0
	for _, startID := range startIDs {
		loadNotionPages(c, startID, idToPage)
		nDownloaded := len(idToPage) - nPrev
		lg("Downloaded %d pages\n", nDownloaded)
		nPrev = len(idToPage)
	}
	return idToPage
}

func rmFile(path string) {
	err := os.Remove(path)
	if err != nil {
		lg("os.Remove(%s) failed with %s\n", path, err)
	}
}

func rmCached(pageID string) {
	id := normalizeID(pageID)
	rmFile(filepath.Join(notionLogDir, id+".go.log.txt"))
	rmFile(filepath.Join(cacheDir, id+".json"))
}

func loadPageAsArticle(c *notionapi.Client, pageID string) *Article {
	page, err := downloadAndCachePage(c, pageID)
	panicIfErr(err)
	lg("Downloaded %s %s\n", pageID, page.Root().Title)
	return notionPageToArticle(c, page)
}
