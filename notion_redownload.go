package main

import (
	"fmt"
	"os"

	"github.com/kjk/notionapi"
)

var (
	cacheDir = "cache"
)

func loadNotionBlogPosts() map[string]*NotionDoc {
	indexPageID := normalizeID("300db9dc27c84958a08b8d0c37f4cfe5")
	doc, err := loadPage(indexPageID)
	panicIfErr(err)
	pageInfo := doc.pageInfo

	res := make(map[string]*NotionDoc)
	for _, block := range pageInfo.Page.Content {
		if block.Type != notionapi.BlockPage {
			continue
		}

		title := block.Title
		id := normalizeID(block.ID)
		if _, ok := res[id]; ok {
			continue
		}
		fmt.Printf("%s-%s\n", title, id)
		doc, err := loadPage(id)
		panicIfErr(err)
		if doc.meta.IsHidden() {
			continue
		}
		res[id] = doc
	}
	return res
}

func notionRedownload() {
	err := os.RemoveAll(cacheDir)
	panicIfErr(err)
	err = os.MkdirAll(cacheDir, 0755)
	panicIfErr(err)
	docs := loadNotionBlogPosts()
	for _, doc := range docs {
		// generate html to verify it'll work
		genHTML(doc.pageInfo)
	}
}
