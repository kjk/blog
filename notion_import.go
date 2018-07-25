package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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

	destDir      = "netlify_static"
	cacheDir     = "notion_cache"
	notionLogDir = "log"

	notionBlogsStartPage      = "300db9dc27c84958a08b8d0c37f4cfe5"
	notionWebsiteStartPage    = "568ac4c064c34ef6a6ad0b8d77230681"
	notionGoCookbookStartPage = "7495260a1daa46118858ad2e049e77e6"
)

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
func normalizeID(s string) string {
	return strings.Replace(s, "-", "", -1)
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
	notionapi.Logger = f
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

func loadPageFromCache(pageID string) *notionapi.PageInfo {
	if !useCacheForNotion {
		return nil
	}

	cachedPath := filepath.Join(cacheDir, pageID+".json")
	d, err := ioutil.ReadFile(cachedPath)
	if err != nil {
		return nil
	}

	var pageInfo notionapi.PageInfo
	err = json.Unmarshal(d, &pageInfo)
	panicIfErr(err)
	fmt.Printf("Got %s from cache (%s)\n", pageID, pageInfo.Page.Title)
	return &pageInfo
}

func downloadAndCachePage(pageID string) (*notionapi.PageInfo, error) {
	//fmt.Printf("downloading page with id %s\n", pageID)
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	res, err := notionapi.GetPageInfo(pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(res, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		panicIfErr(err)
	} else {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return res, nil
}

func notionToHTML(pageInfo *notionapi.PageInfo) []byte {
	gen := NewHTMLGenerator(pageInfo)
	return gen.Gen()
}

func notionPageToArticle(pageInfo *notionapi.PageInfo) *Article {
	blocks := pageInfo.Page.Content
	//fmt.Printf("extractMetadata: %s-%s, %d blocks\n", title, id, len(blocks))
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	page := pageInfo.Page
	title := page.Title
	id := normalizeID(page.ID)
	article := &Article{
		pageInfo: pageInfo,
		Title:    title,
	}
	nBlock := 0
	var publishedOn time.Time
	var err error
	endLoop := false
	for len(blocks) > 0 {
		block := blocks[0]
		//fmt.Printf("  %d %s '%s'\n", nBlock, block.Type, block.Title)

		if block.Type != notionapi.BlockText {
			//fmt.Printf("extractMetadata: ending look because block %d is of type %s\n", nBlock, block.Type)
			break
		}

		if len(block.InlineContent) == 0 {
			//fmt.Printf("block %d of type %s and has no InlineContent\n", nBlock, block.Type)
			blocks = blocks[1:]
			break
		} else {
			//fmt.Printf("block %d has %d InlineContent\n", nBlock, len(block.InlineContent))
		}

		inline := block.InlineContent[0]
		// must be plain text
		if !inline.IsPlain() {
			//fmt.Printf("block: %d of type %s: inline has attributes\n", nBlock, block.Type)
			break
		}

		// remove empty lines at the top
		s := strings.TrimSpace(inline.Text)
		if s == "" {
			//fmt.Printf("block: %d of type %s: inline.Text is empty\n", nBlock, block.Type)
			blocks = blocks[2:]
			break
		}
		//fmt.Printf("  %d %s '%s'\n", nBlock, block.Type, s)

		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			//fmt.Printf("block: %d of type %s: inline.Text is not key/value. s='%s'\n", nBlock, block.Type, s)
			break
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "tags":
			article.Tags = parseTags(val)
			//fmt.Printf("Tags: %v\n", res.Tags)
		case "id":
			articleSetID(article, val)
			//fmt.Printf("ID: %s\n", res.ID)
		case "publishedon":
			publishedOn, err = parseDate(val)
			panicIfErr(err)
			article.inBlog = true
		case "date", "createdat":
			article.PublishedOn, err = parseDate(val)
			panicIfErr(err)
			article.inBlog = true
		case "updatedat":
			article.UpdatedOn, err = parseDate(val)
			panicIfErr(err)
		case "status":
			setStatusMust(article, val)
		case "description":
			article.Description = val
			//fmt.Printf("Description: %s\n", res.Description)
		case "headerimage":
			setHeaderImageMust(article, val)
		case "collection":
			setCollectionMust(article, val)
		default:
			// assume that unrecognized meta means this article doesn't have
			// proper meta tags. It might miss meta-tags that are badly named
			endLoop = true
			/*
				rmCached(pageInfo.ID)
				title := pageInfo.Page.Title
				panicMsg("Unsupported meta '%s' in notion page with id '%s', '%s'", key, normalizeID(pageInfo.ID), title)
			*/
		}
		if endLoop {
			break
		}
		blocks = blocks[1:]
		nBlock++
	}
	page.Content = blocks

	// PublishedOn over-writes Date and CreatedAt
	if !publishedOn.IsZero() {
		// TODO: use pageInfo.Page.CreatedTime if publishedOn.IsZero()
		article.PublishedOn = publishedOn
	}

	if article.UpdatedOn.IsZero() {
		article.UpdatedOn = article.PublishedOn
	}

	if article.PublishedOn.IsZero() {
		article.PublishedOn = page.CreatedOn()
	}

	if article.UpdatedOn.IsZero() {
		article.UpdatedOn = page.UpdatedOn()
	}

	if article.ID == "" {
		article.ID = id
	}

	article.Body = notionToHTML(pageInfo)
	article.BodyHTML = string(article.Body)
	article.HTMLBody = template.HTML(article.BodyHTML)

	if article.Collection != "" {
		path := URLPath{
			Name: article.Collection,
			URL:  article.CollectionURL,
		}
		article.Paths = append(article.Paths, path)
	}

	format := page.FormatPage
	// set image header from cover page
	if article.HeaderImageURL == "" && format != nil && format.PageCoverURL != "" {
		article.HeaderImageURL = format.PageCoverURL
	}
	return article
}

func loadPageAsArticle(pageID string) *Article {
	var err error
	pageInfo := loadPageFromCache(pageID)
	if pageInfo == nil {
		pageInfo, err = downloadAndCachePage(pageID)
		panicIfErr(err)
		fmt.Printf("Downloaded %s %s\n", pageID, pageInfo.Page.Title)
	}
	return notionPageToArticle(pageInfo)
}

func addIdToBlock(block *notionapi.Block, idToBlock map[string]*notionapi.Block) {
	id := normalizeID(block.ID)
	idToBlock[id] = block
	for _, block := range block.Content {
		if block == nil {
			continue
		}
		addIdToBlock(block, idToBlock)
	}
}

func updateArticlesPaths(articles []*Article, rootPageID string) {
	idToBlock := map[string]*notionapi.Block{}
	for _, a := range articles {
		page := a.pageInfo
		if page == nil {
			continue
		}
		addIdToBlock(page.Page, idToBlock)
	}

	for _, article := range articles {
		// some already have path (e.g. those that belong to a collection)
		if len(article.Paths) > 0 {
			continue
		}
		currID := normalizeID(article.pageInfo.Page.ParentID)
		var paths []URLPath
		for currID != rootPageID {
			block := idToBlock[currID]
			if block == nil {
				break
			}
			// parent could be a column
			if block.Type != notionapi.BlockPage {
				currID = normalizeID(block.ParentID)
				continue
			}
			title := block.Title
			uri := "/article/" + normalizeID(block.ID) + "/" + urlify(title)
			path := URLPath{
				Name: title,
				URL:  uri,
			}
			paths = append(paths, path)
			currID = normalizeID(block.ParentID)
		}
		n := len(paths)
		for i := 1; i <= n; i++ {
			path := paths[n-i]
			article.Paths = append(article.Paths, path)
		}
	}
}

func loadNotionPages(indexPageID string) []*Article {
	var articles []*Article

	toVisit := []string{indexPageID}

	for len(toVisit) > 0 {
		pageID := normalizeID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := notionIDToArticle[pageID]; ok {
			continue
		}

		article := loadPageAsArticle(pageID)

		if false && article.Status == statusHidden {
			continue
		}

		notionIDToArticle[pageID] = article
		articles = append(articles, article)

		page := article.pageInfo.Page
		subPages := findSubPageIDs(page.Content)
		toVisit = append(toVisit, subPages...)
	}

	updateArticlesPaths(articles, indexPageID)
	return articles
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

func copyCSS() {
	src := filepath.Join("www", "css", "main.css")
	dst := filepath.Join(destDir, "main.css")
	err := copyFile(dst, src)
	panicIfErr(err)
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

func createDestDir() {
	err := os.MkdirAll(destDir, 0755)
	panicIfErr(err)
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
func notionRedownloadAll() {
	//notionapi.DebugLog = true
	removeCachedNotion()

	loadAllArticles()
	articles := storeArticles
	fmt.Printf("Loaded %d articles\n", len(articles))
}

func notionRedownloadOne(id string) {
	id = normalizeID(id)
	pageInfo, err := downloadAndCachePage(id)
	panicIfErr(err)
	fmt.Printf("Downloaded %s %s\n", id, pageInfo.Page.Title)
}

// downloads and html
func testNotionToHTMLOnePage(id string) {
	//id := "c9bef0f1c8fe40a2bc8b06ace2bd7d8f" // tools page, columns
	//id := "0a66e6c0c36f4de49417a47e2c40a87e" // mono-spaced page with toggle, devlog 2018
	//id := "484919a1647144c29234447ce408ff6b" // test toggle
	//id := "88aee8f43620471aa9dbcad28368174c" // test image and gist
	createNotionDirs()
	createDestDir()
	useCacheForNotion = false

	id = normalizeID(id)
	article := loadPageAsArticle(id)
	path := filepath.Join(destDir, "index.html")
	err := ioutil.WriteFile(path, article.Body, 0644)
	panicIfErr(err)
	copyCSS()

	err = os.Chdir(destDir)
	panicIfErr(err)

	go func() {
		time.Sleep(time.Second * 1)
		openBrowser("http://localhost:2015")
	}()
	runCaddy()
}
