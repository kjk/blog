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
	useCacheForNotion = true
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
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func findSubPageIDs(blocks []*notionapi.Block) []string {
	pageIDs := map[string]struct{}{}
	seen := map[string]struct{}{}
	toVisit := blocks
	for len(toVisit) > 0 {
		block := toVisit[0]
		toVisit = toVisit[1:]
		id := normalizeID(block.ID)
		if block.Type == notionapi.BlockPage {
			pageIDs[id] = struct{}{}
			seen[id] = struct{}{}
		}
		for _, b := range block.Content {
			if b == nil {
				continue
			}
			id := normalizeID(block.ID)
			if _, ok := seen[id]; ok {
				continue
			}
			toVisit = append(toVisit, b)
		}
	}
	res := []string{}
	for id := range pageIDs {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
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
			fmt.Printf("Download %s failed with '%s'\n", pageID, err)
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
	panic(fmt.Errorf("Didn't find ext for file '%s', content type '%s'\n", fileName, contentType))
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
		fmt.Printf("Image %s already downloaded as %s\n", uri, cachedPath)
		return cachedPath, nil
	}

	timeStart := time.Now()
	fmt.Printf("Downloading %s ... ", uri)
	img, err := c.DownloadFile(uri)
	if err != nil {
		fmt.Printf("failed with %s\n", err)
		return "", err
	}
	ext := guessExt(uri, img.Header.Get("Content-Type"))
	cachedPath = filepath.Join(imgDir, sha+ext)

	err = ioutil.WriteFile(cachedPath, img.Data, 0644)
	if err != nil {
		return "", err
	}
	fmt.Printf("finished in %s. Wrote as '%s'\n", time.Since(timeStart), cachedPath)

	return cachedPath, nil
}

func downloadAndCachePage(c *notionapi.Client, pageID string) (*notionapi.Page, error) {
	//fmt.Printf("downloading page with id %s\n", pageID)
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
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return page, nil
}

func notionToHTML(c *notionapi.Client, page *notionapi.Page, articles *Articles) ([]byte, []ImageMapping) {
	gen := NewHTMLGenerator(c, page)
	if articles != nil {
		gen.idToArticle = func(id string) *Article {
			return articles.idToArticle[id]
		}
	}
	return gen.Gen(), gen.images
}

func loadPageBlockInfo(c *notionapi.Client, pageID string) (*notionapi.Block, error) {
	recVals, err := c.GetRecordValues([]string{pageID})
	if err != nil {
		return nil, err
	}
	res := recVals.Results[0]
	// this might happen e.g. when a page is not publicly visible
	if res.Value == nil {
		return nil, fmt.Errorf("Couldn't retrieve page with id %s", pageID)
	}
	return res.Value, nil
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
		fmt.Printf("loadPagesFromDisk: os.ReadDir('%s') failed with '%s'\n", dir, err)
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
	fmt.Printf("loadPagesFromDisk: loaded %d cached pages from %s\n", len(cachedPagesFromDisk), dir)
	return cachedPagesFromDisk
}

func loadNotionPage(c *notionapi.Client, pageID string, getFromCache bool, n int, isCachedPageNotOutdated map[string]bool, cachedPagesFromDisk map[string]*notionapi.Page) (*notionapi.Page, error) {
	if isCachedPageNotOutdated[pageID] {
		page := cachedPagesFromDisk[pageID]
		//nTotalFromCache++
		fmt.Printf("Page %4d %s: skipping (ver not changed), title: %s\n", n, page.ID, page.Root.Title)
		return page, nil
	}

	if getFromCache {
		page := loadPageFromCache(cacheDir, pageID)
		if page != nil {
			//nNotionPagesFromCache++
			//fmt.Printf("Got %d from cache %s %s\n", n, pageID, page.Root.Title)
			return page, nil
		}
	}

	/*
		page := loadPageFromCache(pageID)
		if page != nil {
			if getFromCache {
				fmt.Printf("Page %4d %s: got from cache. Title: %s\n", n, pageID, page.Root.Title)
				return page, nil
			}
			pageBlock, err := loadPageBlockInfo(c, pageID)
			panicIfErr(err)
			if pageBlock.Version == page.Root.Version {
				fmt.Printf("Page %4d %s: skipping re-download, same ver. Title: %s\n", n, pageID, page.Root.Title)
				return page, nil
			}
		}
	*/

	page, err := downloadAndCachePage(c, pageID)
	if err == nil {
		fmt.Printf("Page %4d %s: downloaded. Title: %s\n", n, page.ID, page.Root.Title)
	}
	return page, err
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
	maxPerCall := 128
	for len(rest) > 0 {
		n := len(rest)
		if n > maxPerCall {
			n = maxPerCall
		}
		tmpIDs := rest[:n]
		rest = rest[n:]
		fmt.Printf("getting versions for %d pages\n", len(tmpIDs))
		tmpVers, err := getVersionsForPages(c, tmpIDs)
		panicIfErr(err)
		versions = append(versions, tmpVers...)
	}
	panicIf(len(ids) != len(versions))
	nOutdated := 0
	for i, ver := range versions {
		id := ids[i]
		page := cachedPagesFromDisk[id]
		isOutdated := ver > page.Root.Version
		isCachedPageNotOutdated[id] = !isOutdated
		if isOutdated {
			nOutdated++
		}
	}
	fmt.Printf("checkIfPagesAreOutdated: %d pages, %d outdated\n", len(ids), nOutdated)
	return isCachedPageNotOutdated
}

func loadNotionPages(c *notionapi.Client, indexPageID string, idToPage map[string]*notionapi.Page, useCache bool) {
	cachedPagesFromDisk := loadPagesFromDisk(cacheDir)
	isCachedPageNotOutdated := checkIfPagesAreOutdated(c, cachedPagesFromDisk)

	toVisit := []string{indexPageID}

	n := 1
	for len(toVisit) > 0 {
		pageID := normalizeID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := idToPage[pageID]; ok {
			continue
		}

		page, err := loadNotionPage(c, pageID, useCache, n, isCachedPageNotOutdated, cachedPagesFromDisk)
		panicIfErr(err)
		n++

		idToPage[pageID] = page

		subPages := findSubPageIDs(page.Root.Content)
		toVisit = append(toVisit, subPages...)
	}
}

func loadAllPages(c *notionapi.Client, startIDs []string, useCache bool) map[string]*notionapi.Page {
	idToPage := map[string]*notionapi.Page{}
	nPrev := 0
	for _, startID := range startIDs {
		loadNotionPages(c, startID, idToPage, useCache)
		nDownloaded := len(idToPage) - nPrev
		fmt.Printf("Downloaded %d pages\n", nDownloaded)
		nPrev = len(idToPage)
	}
	return idToPage
}

func rmFile(path string) {
	err := os.Remove(path)
	if err != nil {
		fmt.Printf("os.Remove(%s) failed with %s\n", path, err)
	}
}

func rmCached(pageID string) {
	id := normalizeID(pageID)
	rmFile(filepath.Join(notionLogDir, id+".go.log.txt"))
	rmFile(filepath.Join(cacheDir, id+".json"))
}

func createNotionCacheDir() {
	err := os.MkdirAll(cacheDir, 0755)
	panicIfErr(err)
}

func createNotionLogDir() {
	if logNotionRequests {
		err := os.MkdirAll(notionLogDir, 0755)
		panicIfErr(err)
	}
}

func createNotionDirs() {
	createNotionLogDir()
	createNotionCacheDir()
}

func removeCachedNotion() {
	err := os.RemoveAll(cacheDir)
	panicIfErr(err)
	err = os.RemoveAll(notionLogDir)
	panicIfErr(err)
	createNotionDirs()
}

// this re-downloads pages from Notion by deleting cache locally
func notionRedownloadAll(c *notionapi.Client) {
	//notionapi.DebugLog = true
	//removeCachedNotion()
	useCacheForNotion = false
	err := os.RemoveAll(notionLogDir)
	panicIfErr(err)
	createNotionDirs()

	timeStart := time.Now()
	articles := loadArticles(c)
	fmt.Printf("Loaded %d articles in %s\n", len(articles.idToPage), time.Since(timeStart))
}

func notionRedownloadOne(c *notionapi.Client, id string) {
	id = normalizeID(id)
	page, err := downloadAndCachePage(c, id)
	panicIfErr(err)
	fmt.Printf("Downloaded %s %s\n", id, page.Root.Title)
}

func loadPageAsArticle(c *notionapi.Client, pageID string) *Article {
	var err error
	var page *notionapi.Page
	if useCacheForNotion {
		page = loadPageFromCache(cacheDir, pageID)
	}
	if page == nil {
		page, err = downloadAndCachePage(c, pageID)
		panicIfErr(err)
		fmt.Printf("Downloaded %s %s\n", pageID, page.Root.Title)
	}
	return notionPageToArticle(c, page)
}
