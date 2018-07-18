package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

var (
	cacheDir                  = "notion_cache"
	notionLogDir              = "log"
	notionBlogsStartPage      = "300db9dc27c84958a08b8d0c37f4cfe5"
	notionWebsiteStartPage    = "568ac4c064c34ef6a6ad0b8d77230681"
	notionGoCookbookStartPage = "7495260a1daa46118858ad2e049e77e6"
)

func findSubPageIDs(blocks []*notionapi.Block) []string {
	var res []string
	for _, block := range blocks {
		if block.Type == notionapi.BlockPage {
			res = append(res, block.ID)
		}
	}
	return res
}

func loadPageFromCache(pageID string) *notionapi.PageInfo {
	var pageInfo notionapi.PageInfo
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	if useCache {
		d, err := ioutil.ReadFile(cachedPath)
		if err == nil {
			err = json.Unmarshal(d, &pageInfo)
			panicIfErr(err)
			//fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
			return &pageInfo
		}
	}
	return nil
}

func downloadAndCachePage(pageID string) (*notionapi.PageInfo, error) {
	//fmt.Printf("downloading page with id %s\n", pageID)
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
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

func loadPageAsArticle(pageID string) (*Article, error) {
	var err error
	pageInfo := loadPageFromCache(pageID)
	if pageInfo == nil {
		pageInfo, err = downloadAndCachePage(pageID)
		if err != nil {
			return nil, err
		}
	}
	return articleFromPage(pageInfo), nil
}

func loadNotionPages(indexPageID string) map[string]*Article {
	toVisit := []string{indexPageID}
	res := make(map[string]*Article)

	for len(toVisit) > 0 {
		pageID := normalizeID(toVisit[0])
		toVisit = toVisit[1:]

		if _, ok := res[pageID]; ok {
			continue
		}

		article, err := loadPageAsArticle(pageID)
		panicIfErr(err)
		fmt.Printf("Downloaded %s %s\n", pageID, article.Title)
		if article.Status == statusHidden {
			continue
		}

		res[pageID] = article

		page := article.pageInfo.Page
		subPages := findSubPageIDs(page.Content)
		toVisit = append(toVisit, subPages...)
	}
	return res
}

// this re-downloads pages from Notion by deleting cache locally
func notionRedownload() {
	//notionapi.DebugLog = true

	err := os.RemoveAll(cacheDir)
	panicIfErr(err)
	err = os.MkdirAll(notionLogDir, 0755)
	panicIfErr(err)
	err = os.MkdirAll(cacheDir, 0755)
	panicIfErr(err)

	docs := make(map[string]*Article)

	if true {
		articles := loadNotionPages(notionBlogsStartPage)
		fmt.Printf("Loaded %d blog articles\n\n", len(articles))
		for k, v := range articles {
			docs[k] = v
		}
	}

	if false {
		articles := loadNotionPages(notionGoCookbookStartPage)
		fmt.Printf("Loaded %d go cookbook articles\n\n", len(articles))
		for k, v := range articles {
			docs[k] = v
		}
	}

	if false {
		articles := loadNotionPages(notionWebsiteStartPage)
		fmt.Printf("Loaded %d articles\n", len(articles))
		for k, v := range articles {
			docs[k] = v
		}
	}

	for _, doc := range docs {
		// generate html to verify it'll work
		genHTML(doc.pageInfo)
	}
}
