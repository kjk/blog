package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kjk/notionapi"
)

type CachingDownloader struct {
	Client              *notionapi.Client
	CacheDir            string
	idToPage            map[string]*notionapi.Page
	cachedPagesFromDisk map[string]*notionapi.Page
	// pages that were loaded from cache but are outdated
	cachedOutdatedPages map[string]bool
	nDownloaded         int
}

func NewCachingDownloader(cacheDir string) *CachingDownloader {
	return &CachingDownloader{
		Client:              &notionapi.Client{},
		CacheDir:            cacheDir,
		idToPage:            make(map[string]*notionapi.Page),
		cachedPagesFromDisk: make(map[string]*notionapi.Page),
		cachedOutdatedPages: map[string]bool{},
	}
}

// I got "connection reset by peer" error once so retry download 3 times, with a short sleep in-between
func downloadPageRetry(c *notionapi.Client, pageID string) (*notionapi.Page, *notionapi.HTTPCache, error) {
	var res *notionapi.Page
	var err error
	for i := 0; i < 3; i++ {
		if i > 0 {
			lg("Download %s failed with '%s'\n", pageID, err)
			time.Sleep(5 * time.Second) // not sure if it matters
		}
		httpCache := notionapi.NewHTTPCache()
		c.HTTPClient = notionapi.NewCachingHTTPClient(httpCache)
		res, err = c.DownloadPage(pageID)
		if err == nil {
			return res, httpCache, nil
		}
	}
	return nil, nil, err
}

func loadHTTPCacheForPage(path string) *notionapi.HTTPCache {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		// it's ok if file doesn't exit
		return nil
	}
	httpCache, err := deserializeHTTPCache(d)
	if err != nil {
		err = os.Remove(path)
		must(err)
		fmt.Printf("Deleted file %s\n", path)
	}
	return httpCache
}

func loadPageFromCache(dir, pageID string) *notionapi.Page {
	path := filepath.Join(dir, pageID+".txt")
	httpCache := loadHTTPCacheForPage(path)
	if httpCache == nil {
		return nil
	}
	httpClient := notionapi.NewCachingHTTPClient(httpCache)
	client := &notionapi.Client{
		//DebugLog:   true,
		//Logger:     os.Stdout,
		HTTPClient: httpClient,
	}
	page, err := client.DownloadPage(pageID)
	must(err)
	panicIf(httpCache.RequestsNotFromCache != 0, "unexpectedly made %d server connections for page %s", httpCache.RequestsNotFromCache, pageID)
	return page
}

func pageIDFromFileName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return ""
	}
	id := parts[0]
	if notionapi.IsValidNoDashID(id) {
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
		panicIf(page == nil)
		cachedPagesFromDisk[pageID] = page
	}
	lg("loadPagesFromDisk: loaded %d cached pages from %s\n", len(cachedPagesFromDisk), dir)
	return cachedPagesFromDisk
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

func (d *CachingDownloader) checkIfPagesAreOutdated() {
	var ids []string
	for id := range d.cachedPagesFromDisk {
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
		tmpVers, err := getVersionsForPages(d.Client, tmpIDs)
		panicIfErr(err)
		versions = append(versions, tmpVers...)
	}
	panicIf(len(ids) != len(versions))
	nOutdated := 0
	for i, ver := range versions {
		id := ids[i]
		page := d.cachedPagesFromDisk[id]
		isOutdated := ver > page.Root().Version
		d.cachedOutdatedPages[id] = !isOutdated
		if isOutdated {
			nOutdated++
		}
	}
	lg("checkIfPagesAreOutdated: %d pages, %d outdated\n", len(ids), nOutdated)
}

// returns true if did build
func (d *CachingDownloader) maybeBuildIDToPageMap() bool {
	if !flgNoDownload {
		return false
	}
	if len(d.cachedPagesFromDisk) == 0 {
		fmt.Printf("ignoring flgNoDownload=%v because no cached pages\n", flgNoDownload)
		return false
	}
	for _, page := range d.cachedPagesFromDisk {
		id := page.ID
		id = notionapi.ToNoDashID(id)
		d.idToPage[id] = page
	}
	return true
}

func (d *CachingDownloader) DownloadPages(indexPageID string) ([]*notionapi.Page, error) {
	d.cachedPagesFromDisk = loadPagesFromDisk(d.CacheDir)
	if d.maybeBuildIDToPageMap() {
		return nil, nil
	}

	d.checkIfPagesAreOutdated()
	toVisit := []string{indexPageID}

	d.nDownloaded = 1
	for len(toVisit) > 0 {
		pageID := notionapi.ToNoDashID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := d.idToPage[pageID]; ok {
			continue
		}

		page, err := d.DownloadPage(pageID)
		panicIfErr(err)
		d.nDownloaded++

		d.idToPage[pageID] = page

		subPages := notionapi.GetSubPages(page.Root().Content)
		toVisit = append(toVisit, subPages...)
	}

	return nil, nil
}

func (d *CachingDownloader) loadAllPages(startIDs []string) map[string]*notionapi.Page {
	nPrev := 0
	for _, startID := range startIDs {
		d.DownloadPages(startID)
		nDownloaded := len(d.idToPage) - nPrev
		lg("Downloaded %d pages\n", nDownloaded)
		nPrev = len(d.idToPage)
	}
	return d.idToPage
}

func (d *CachingDownloader) downloadAndCachePage(pageID string) (*notionapi.Page, error) {
	//verbose("downloading page with id %s\n", pageID)
	c := d.Client
	prevClient := c.HTTPClient
	defer func() {
		c.HTTPClient = prevClient
	}()

	page, httpCache, err := downloadPageRetry(c, pageID)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cacheDir, pageID+".txt")
	data, err := serializeHTTPCache(httpCache)
	must(err)
	err = ioutil.WriteFile(path, data, 0644)
	panicIfErr(err)
	return page, nil
}

func (d *CachingDownloader) DownloadPage(pageID string) (*notionapi.Page, error) {
	if d.cachedOutdatedPages[pageID] {
		page := d.cachedPagesFromDisk[pageID]
		//nTotalFromCache++
		title := page.Root().Title
		verbose("Page %4d %s: skipping (ver not changed), title: %s\n", d.nDownloaded, page.ID, title)
		return page, nil
	}

	page, err := d.downloadAndCachePage(pageID)
	if err != nil {
		return nil, err
	}
	lg("Page %4d %s: downloaded. Title: %s\n", d.nDownloaded, page.ID, page.Root().Title)
	return page, nil
	return nil, nil
}
