package main

import (
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
	flgRecursive bool
	useCache     = true
	destDir      = "notion_www"
	toVisit      = []string{
		// 57-MicroConf-videos-for-self-funded-software-businesses
		"0c896ea2efd24ec7be1d1f6e3b22d254",
	}
)

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
func normalizeID(s string) string {
	return strings.Replace(s, "-", "", -1)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
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

func articleFromPage(pageInfo *notionapi.PageInfo) *Article {
	blocks := pageInfo.Page.Content
	//fmt.Printf("extractMetadata: %s-%s, %d blocks\n", title, id, len(blocks))
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	page := pageInfo.Page
	title := page.Title
	id := normalizeID(page.ID)
	res := &Article{
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
			res.Tags = parseTags(val)
			//fmt.Printf("Tags: %v\n", res.Tags)
		case "id":
			articleSetID(res, val)
			//fmt.Printf("ID: %s\n", res.ID)

		case "publishedon":
			publishedOn, err = parseDate(val)
			panicIfErr(err)
		case "date", "createdat":
			res.PublishedOn, err = parseDate(val)
			panicIfErr(err)
		case "updatedat":
			res.UpdatedOn, err = parseDate(val)
			panicIfErr(err)
		case "status":
			res.Status, err = parseStatus(val)
			panicIfErr(err)
		case "description":
			res.Description = val
			//fmt.Printf("Description: %s\n", res.Description)
		case "headerimage":
			setHeaderImageMust(res, val)
		case "collection":
			setCollectionMust(res, val)
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
	pageInfo.Page.Content = blocks

	// PublishedOn over-writes Date and CreatedAt
	if !publishedOn.IsZero() {
		res.PublishedOn = publishedOn
	}

	if res.UpdatedOn.IsZero() {
		res.UpdatedOn = res.PublishedOn
	}

	if res.ID == "" {
		res.ID = id
	}

	gen := NewHTMLGenerator(pageInfo)
	res.Body = gen.Gen()

	res.BodyHTML = string(res.Body)
	res.HTMLBody = template.HTML(res.BodyHTML)

	return res
}

func loadArticlesFromNotion() []*Article {
	indexID := "300db9dc27c84958a08b8d0c37f4cfe5"
	fileInfos, err := ioutil.ReadDir(cacheDir)
	panicIfErr(err)

	var res []*Article
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		ext := filepath.Ext(name)
		if ext != ".json" {
			continue
		}
		if strings.Contains(name, indexID) {
			continue
		}
		parts := strings.Split(name, ".")
		pageID := parts[0]
		pageInfo := loadPageFromCache(pageID)
		article := articleFromPage(pageInfo)
		res = append(res, article)
	}

	return res
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

func genHTML(pageInfo *notionapi.PageInfo) []byte {
	title := pageInfo.Page.Title
	title = template.HTMLEscapeString(title)

	gen := NewHTMLGenerator(pageInfo)
	html := string(gen.Gen())

	s := fmt.Sprintf(`<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>%s</title>
		<link href="/main.css" rel="stylesheet">
	</head>
<body>
<div id="tophdr">
<ul id="nav">
  <li><a href="/software/">Software</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/archives.html">Articles</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/documents.html">Documents</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/resume.html">Résumé</a></li>
</ul>
</div>

<div id="content">
  <div id="post" style="margin-left:auto;margin-right:auto;margin-top:2em;">
    <div class="title">
      <a href="/">Home</a>  / %s
    </div>
    <div>
      %s
    </div>
  </div>
</div>
</body>
</html>
`, title, title, html)

	d := prettyHTML([]byte(s))
	return d
}

func toHTML(pageID, path string) (*Article, error) {
	fmt.Printf("toHTML: pageID=%s, path=%s\n", pageID, path)
	article, err := loadPageAsArticle(pageID)
	if err != nil {
		return nil, err
	}
	d := genHTML(article.pageInfo)
	err = ioutil.WriteFile(path, d, 0644)
	return article, err
}

func copyCSS() {
	src := filepath.Join("www", "css", "main.css")
	dst := filepath.Join(destDir, "main.css")
	err := copyFile(dst, src)
	panicIfErr(err)
}

func loadOne(id string) {
	id = normalizeID(id)
	_, err := loadPageAsArticle(id)
	panicIfErr(err)
}

func genIndexHTML(docs []*Article) []byte {
	lines := []string{}
	for _, doc := range docs {
		if doc.Status == statusNotImportant {
			continue
		}
		page := doc.pageInfo.Page
		id := normalizeID(page.ID)
		title := page.Title
		s := fmt.Sprintf(`<div>
		<a href="/article/%s/index.html">%s</a>
			<span style="font-size:80%%">
				<span class="taglink">in:</span> <a href="/tag/go" class="taglink">go</a>, <a href="/tag/programming" class="taglink">programming</a>
			</span>
</div>`, id, title)
		lines = append(lines, s)
	}
	html := strings.Join(lines, "\n")

	s := fmt.Sprintf(`<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Krzysztof Kowalczyk's external brain</title>
		<link href="/main.css" rel="stylesheet">
	</head>
<body>
<div id="tophdr">
<ul id="nav">
  <li><a href="/software/">Software</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/archives.html">Articles</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/documents.html">Documents</a></li>
  <li><span style="color:#aaa">&bull;</span></li>
  <li><a href="/resume.html">Résumé</a></li>
</ul>
</div>

<div id="content">
  <div id="post" style="margin-left:auto;margin-right:auto;margin-top:2em;">
    <div class="title">
      <a href="/">Home</a>
    </div>
    <div class="articles-list-wrap">
      %s
    </div>
  </div>
</div>
</body>
</html>
`, html)

	d := prettyHTML([]byte(s))
	return d
}

func genNotionBasic(pages map[string]*Article) {
	docs := make([]*Article, 0)
	for _, doc := range pages {
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(i, j int) bool {
		d1 := docs[i].PublishedOn
		d2 := docs[j].PublishedOn
		return d1.Sub(d2) > 0
	})
	d := genIndexHTML(docs)
	path := filepath.Join(destDir, "index.html")
	err := ioutil.WriteFile(path, d, 0644)
	panicIfErr(err)
	for _, doc := range docs {
		d := genHTML(doc.pageInfo)
		id := normalizeID(doc.pageInfo.Page.ID)
		path := filepath.Join(destDir, id+".html")
		err = ioutil.WriteFile(path, d, 0644)
	}
}

func importNotion() {
	os.MkdirAll(notionLogDir, 0755)
	os.MkdirAll(cacheDir, 0755)
	os.MkdirAll(destDir, 0755)

	if false {
		//loadOne("431295a5-4f7e-4208-869f-4763862c1f05")
		docs := loadNotionPages(notionBlogsStartPage)
		genNotionBasic(docs)
		return
	}

	notionapi.DebugLog = true
	seen := map[string]struct{}{}
	firstPage := true
	for len(toVisit) > 0 {
		pageID := toVisit[0]
		toVisit = toVisit[1:]
		id := normalizeID(pageID)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		name := id + ".html"
		if firstPage {
			name = "index.html"
		}
		path := filepath.Join(destDir, name)
		doc, err := toHTML(id, path)
		if err != nil {
			fmt.Printf("toHTML('%s') failed with %s\n", id, err)
		}
		if flgRecursive {
			subPages := findSubPageIDs(doc.pageInfo.Page.Content)
			toVisit = append(toVisit, subPages...)
		}
		firstPage = false
	}
	copyCSS()
}
