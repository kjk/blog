package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/yosssi/gohtml"
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
	path := filepath.Join("log", name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	notionapi.Logger = f
	return f, nil
}

// HTMLGenerator is for notion -> HTML generation
type HTMLGenerator struct {
	f        *bytes.Buffer
	pageInfo *notionapi.PageInfo
	level    int
	err      error
}

func NewHTMLGenerator(pageInfo *notionapi.PageInfo) *HTMLGenerator {
	return &HTMLGenerator{
		f:        &bytes.Buffer{},
		pageInfo: pageInfo,
	}
}

func (g *HTMLGenerator) Gen() []byte {
	g.genContent(g.pageInfo.Page)
	return g.f.Bytes()
}

func (g *HTMLGenerator) genInlineBlock(b *notionapi.InlineBlock) error {
	var start, close string
	if b.AttrFlags&notionapi.AttrBold != 0 {
		start += "<b>"
		close += "</b>"
	}
	if b.AttrFlags&notionapi.AttrItalic != 0 {
		start += "<i>"
		close += "</i>"
	}
	if b.AttrFlags&notionapi.AttrStrikeThrought != 0 {
		start += "<strike>"
		close += "</strike>"
	}
	if b.AttrFlags&notionapi.AttrCode != 0 {
		start += "<code>"
		close += "</code>"
	}
	skipText := false
	if b.Link != "" {
		start += fmt.Sprintf(`<a href="%s">%s</a>`, b.Link, b.Text)
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="user">@%s</span>`, b.UserID)
		skipText = true
	}
	if b.Date != nil {
		// TODO: serialize date properly
		start += fmt.Sprintf(`<span class="date">@TODO: date</span>`)
		skipText = true
	}
	if !skipText {
		start += b.Text
	}
	_, err := io.WriteString(g.f, start+close)
	return err
}

func (g *HTMLGenerator) genInlineBlocks(blocks []*notionapi.InlineBlock) error {
	for _, block := range blocks {
		err := g.genInlineBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *HTMLGenerator) genBlockSurrouded(block *notionapi.Block, start, close string) {
	io.WriteString(g.f, start+"\n")
	g.genInlineBlocks(block.InlineContent)
	g.level++
	g.genContent(block)
	g.level--
	io.WriteString(g.f, close+"\n")
}

func (g *HTMLGenerator) genBlock(block *notionapi.Block) {
	levelCls := ""
	if g.level > 0 {
		levelCls = fmt.Sprintf(" lvl%d", g.level)
	}

	switch block.Type {
	case notionapi.BlockText:
		start := fmt.Sprintf(`<p>`)
		close := `</p>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockHeader:
		start := fmt.Sprintf(`<h1 class="hdr%s">`, levelCls)
		close := `</h1>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockSubHeader:
		start := fmt.Sprintf(`<h2 class="hdr%s">`, levelCls)
		close := `</h2>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockTodo:
		clsChecked := ""
		if block.IsChecked {
			clsChecked = " todo-checked"
		}
		start := fmt.Sprintf(`<div class="todo%s%s">`, levelCls, clsChecked)
		close := `</div>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockToggle:
		start := fmt.Sprintf(`<div class="toggle%s">`, levelCls)
		close := `</div>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockQuote:
		start := fmt.Sprintf(`<quote class="%s">`, levelCls)
		close := `</quote>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockDivider:
		fmt.Fprintf(g.f, `<hr class="%s"/>`+"\n", levelCls)
	case notionapi.BlockPage:
		id := strings.TrimSpace(block.ID)
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page-link"
		}
		title := template.HTMLEscapeString(block.Title)
		url := normalizeID(id) + ".html"
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		fmt.Fprintf(g.f, "%s\n", html)
	case notionapi.BlockCode:
		code := template.HTMLEscapeString(block.Code)
		fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
<pre class="%s">
%s
</pre>`, levelCls, block.CodeLanguage, levelCls, code)
	case notionapi.BlockBookmark:
		fmt.Fprintf(g.f, `<div class="bookmark %s">Bookmark to %s</div>`+"\n", levelCls, block.Link)
	case notionapi.BlockGist:
		fmt.Fprintf(g.f, `<div class="gist %s">Gist for %s</div>`+"\n", levelCls, block.Source)
	case notionapi.BlockImage:
		link := block.ImageURL
		fmt.Fprintf(g.f, `<img class="%s" src="%s" />`+"\n", levelCls, link)
	case notionapi.BlockColumnList:
		// TODO: implement me
	case notionapi.BlockCollectionView:
		// TODO: implement me
	default:
		fmt.Printf("Unsupported block type '%s', id: %s\n", block.Type, block.ID)
		panic(fmt.Sprintf("Unsupported block type '%s'", block.Type))
	}
}

func (g *HTMLGenerator) genBlocks(blocks []*notionapi.Block) {
	for len(blocks) > 0 {
		block := blocks[0]
		if block == nil {
			fmt.Printf("Missing block\n")
			blocks = blocks[1:]
			continue
		}

		if block.Type == notionapi.BlockNumberedList {
			fmt.Fprintf(g.f, `<ol>`)
			for len(blocks) > 0 {
				block := blocks[0]
				if block.Type != notionapi.BlockNumberedList {
					break
				}
				g.genBlockSurrouded(block, `<li>`, `</li>`)
				blocks = blocks[1:]
			}
			fmt.Fprintf(g.f, `</ol>`)
		} else if block.Type == notionapi.BlockBulletedList {
			fmt.Fprintf(g.f, `<ul>`)
			for len(blocks) > 0 {
				block := blocks[0]
				if block.Type != notionapi.BlockBulletedList {
					break
				}
				g.genBlockSurrouded(block, `<li>`, `</li>`)
				blocks = blocks[1:]
			}
			fmt.Fprintf(g.f, `</ul>`)
		} else {
			g.genBlock(block)
			blocks = blocks[1:]
		}
	}
}

func (g *HTMLGenerator) genContent(parent *notionapi.Block) {
	g.genBlocks(parent.Content)
}

// Metadata describes meta information extracted from the page
type Metadata struct {
	ID          string
	Tags        []string
	Date        string
	DateParsed  time.Time
	Description string
	HeaderImage string
	Collection  string
	Status      string // hidden, notimportant
}

// IsHidden returns true if page is hidden/deleted
func (m *Metadata) IsHidden() bool {
	return strings.EqualFold(m.Status, "hidden")
}

func prettyHTML(d []byte) []byte {
	gohtml.Condense = true
	s := string(d)
	s = gohtml.Format(s)
	return []byte(s)
}

// exttract metadata from blocks
func extractMetadata(pageInfo *notionapi.PageInfo) *Metadata {
	blocks := pageInfo.Page.Content
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	res := Metadata{}
	nBlock := 0
	seenEmpty := false
	for len(blocks) > 0 {
		block := blocks[0]

		if block.Type != notionapi.BlockText {
			//fmt.Printf("extractMetadata: ending look because block %d is of type %s\n", nBlock, block.Type)
			break
		}

		if len(block.InlineContent) == 0 {
			//fmt.Printf("block %d of type %s and has no InlineContent\n", nBlock, block.Type)
			seenEmpty = true
			blocks = blocks[1:]
			break
		}

		inline := block.InlineContent[0]
		// must be plain text
		if !inline.IsPlain() {
			fmt.Printf("block: %d of type %s: inline has attributes\n", nBlock, block.Type)
			break
		}

		blocks = blocks[1:]

		// remove empty lines at the top
		s := strings.TrimSpace(inline.Text)
		if s == "" {
			//fmt.Printf("block: %d of type %s: inline.Text is empty\n", nBlock, block.Type)
			seenEmpty = true
			continue
		}

		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			//fmt.Printf("block: %d of type %s: inline.Text is not key/value. s='%s'\n", nBlock, block.Type, s)
			if seenEmpty {
				break
			}
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "tags":
			res.Tags = strings.Split(val, ",")
			for i, tag := range res.Tags {
				res.Tags[i] = strings.TrimSpace(tag)
			}
			//fmt.Printf("Tags: %v\n", res.Tags)
		case "id":
			res.ID = val
			//fmt.Printf("ID: %s\n", res.ID)
		case "date", "createdat", "updatedat":
			res.Date = val
			// 2002-06-21T04:15:29-07:00
			parsed, err := time.Parse(time.RFC3339, res.Date)
			if err != nil {
				panicMsg("Failed to parse date '%s' in notion page with id '%s'. Error: %s", res.Date, pageInfo.ID, err)
			}
			res.DateParsed = parsed
			//fmt.Printf("Date: %s\n", res.Date)
		case "status":
			res.Status = val
		case "description":
			res.Description = val
			//fmt.Printf("Description: %s\n", res.Description)
		case "headerimage":
			res.HeaderImage = val
		case "collection":
			res.Collection = val
		default:
			rmCached(pageInfo.ID)
			panicMsg("Unsupported meta '%s' in notion page with id '%s'", key, pageInfo.ID)
		}
		nBlock++
	}
	pageInfo.Page.Content = blocks
	return &res
}

func rmCached(pageID string) {
	id := normalizeID(pageID)
	{
		path := filepath.Join("log", id+".log.txt")
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("os.Remove(%s) failed with %s\n", path, err)
		}
	}

	{
		path := filepath.Join("cache", id+".json")
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("os.Remove(%s) failed with %s\n", path, err)
		}
	}
}

func panicMsg(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Printf("%s\n", s)
	panic(s)
}

func genHTML(pageID string, pageInfo *notionapi.PageInfo) []byte {
	extractMetadata(pageInfo)
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
  <li><a href="/dailynotes">Daily Notes</a></li>
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

func getPageInfoCached(pageID string) (*notionapi.PageInfo, error) {
	var pageInfo notionapi.PageInfo
	cachedPath := filepath.Join("cache", pageID+".json")
	if useCache {
		d, err := ioutil.ReadFile(cachedPath)
		if err == nil {
			err = json.Unmarshal(d, &pageInfo)
			if err == nil {
				fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
				return &pageInfo, nil
			}
			// not a fatal error, just a warning
			fmt.Printf("json.Unmarshal() on '%s' failed with %s\n", cachedPath, err)
		}
	}
	res, err := notionapi.GetPageInfo(pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(res, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		if err != nil {
			// not a fatal error, just a warning
			fmt.Printf("ioutil.WriteFile(%s) failed with %s\n", cachedPath, err)
		}
	} else {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return res, nil
}

func loadPage(pageID string) (*notionapi.PageInfo, error) {
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	pageInfo, err := getPageInfoCached(pageID)
	if err != nil {
		fmt.Printf("getPageInfoCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return pageInfo, nil
}

func toHTML(pageID, path string) (*notionapi.PageInfo, error) {
	fmt.Printf("toHTML: pageID=%s, path=%s\n", pageID, path)
	pageInfo, err := loadPage(pageID)
	if err != nil {
		return nil, err
	}
	d := genHTML(pageID, pageInfo)
	err = ioutil.WriteFile(path, d, 0644)
	return pageInfo, err
}

func findSubPageIDs(blocks []*notionapi.Block) []string {
	var res []string
	for _, block := range blocks {
		if block.Type == notionapi.BlockPage {
			res = append(res, block.ID)
		}
	}
	return res
}

func copyCSS() {
	src := filepath.Join("www", "css", "main.css")
	dst := filepath.Join(destDir, "main.css")
	err := copyFile(dst, src)
	panicIfErr(err)
}

// NotionDoc represents a notion page and additional info we need about it
type NotionDoc struct {
	pageInfo *notionapi.PageInfo
	meta     *Metadata
}

func loadNotionBlogPosts() map[string]*NotionDoc {
	indexPageID := normalizeID("300db9dc27c84958a08b8d0c37f4cfe5")
	pageInfo, err := loadPage(indexPageID)
	panicIfErr(err)
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
		page, err := loadPage(id)
		panicIfErr(err)
		meta := extractMetadata(page)
		if meta.IsHidden() {
			continue
		}
		doc := &NotionDoc{
			pageInfo: page,
			meta:     meta,
		}
		res[id] = doc
	}
	return res
}

func importNotion() {
	os.MkdirAll("log", 0755)
	os.MkdirAll("cache", 0755)
	os.MkdirAll(destDir, 0755)

	if true {
		loadNotionBlogPosts()
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
		pageInfo, err := toHTML(id, path)
		if err != nil {
			fmt.Printf("toHTML('%s') failed with %s\n", id, err)
		}
		if flgRecursive {
			subPages := findSubPageIDs(pageInfo.Page.Content)
			toVisit = append(toVisit, subPages...)
		}
		firstPage = false
	}
	copyCSS()
}
