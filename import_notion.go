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
)

var (
	flgRecursive bool
	useCache     = true
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

func genHTMLTitle(f io.Writer, pageBlock *notionapi.Block) {
	title := ""
	if len(pageBlock.InlineContent) > 0 {
		title = pageBlock.InlineContent[0].Text
		title = template.HTMLEscapeString(title)
	}

	s := fmt.Sprintf(`  <div class="title">%s</div>%s`, title, "\n")
	io.WriteString(f, s)
}

func genInlineBlockHTML(f io.Writer, b *notionapi.InlineBlock) error {
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
	for _, attrRaw := range b.Attrs {
		switch attr := attrRaw.(type) {
		case *notionapi.AttrLink:
			start += fmt.Sprintf(`<a href="%s">%s</a>`, attr.Link, b.Text)
			skipText = true
		case *notionapi.AttrUser:
			start += fmt.Sprintf(`<span class="user">@%s</span>`, attr.UserID)
			skipText = true
		case *notionapi.AttrDate:
			// TODO: serialize date properly
			start += fmt.Sprintf(`<span class="date">@TODO: date</span>`)
			skipText = true
		}
	}
	if !skipText {
		start += b.Text
	}
	_, err := io.WriteString(f, start+close)
	if err != nil {
		return err
	}
	return nil
}

func genInlineBlocksHTML(f io.Writer, blocks []*notionapi.InlineBlock) error {
	for _, block := range blocks {
		err := genInlineBlockHTML(f, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func genBlockSurroudedHTML(f io.Writer, block *notionapi.Block, start, close string, level int) {
	io.WriteString(f, start+"\n")
	genInlineBlocksHTML(f, block.InlineContent)
	genBlocksHTML(f, block, level+1)
	io.WriteString(f, close+"\n")
}

func genBlockHTML(f io.Writer, block *notionapi.Block, level int) {
	levelCls := ""
	if level > 0 {
		levelCls = fmt.Sprintf(" lvl%d", level)
	}

	switch block.Type {
	case notionapi.TypeText:
		start := fmt.Sprintf(`<div class="text%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeHeader:
		start := fmt.Sprintf(`<h1 class="hdr%s">`, levelCls)
		close := `</h1>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeSubHeader:
		start := fmt.Sprintf(`<h2 class="hdr%s">`, levelCls)
		close := `</h2>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeTodo:
		clsChecked := ""
		if block.IsChecked {
			clsChecked = " todo-checked"
		}
		start := fmt.Sprintf(`<div class="todo%s%s">`, levelCls, clsChecked)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeToggle:
		start := fmt.Sprintf(`<div class="toggle%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeBulletedList:
		start := fmt.Sprintf(`<div class="bullet-list%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeNumberedList:
		start := fmt.Sprintf(`<div class="numbered-list%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeQuote:
		start := fmt.Sprintf(`<quote class="%s">`, levelCls)
		close := `</quote>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.TypeDivider:
		fmt.Fprintf(f, `<hr class="%s"/>`+"\n", levelCls)
	case notionapi.TypePage:
		id := strings.TrimSpace(block.ID)
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page-link"
		}
		title := template.HTMLEscapeString(block.Title)
		url := normalizeID(id) + ".html"
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		fmt.Fprintf(f, "%s\n", html)
	case notionapi.TypeCode:
		code := template.HTMLEscapeString(block.Code)
		fmt.Fprintf(f, `<div class="%s">Lang for code: %s</div>
<pre class="%s">
%s
</pre>`, levelCls, block.CodeLanguage, levelCls, code)
	case notionapi.TypeBookmark:
		fmt.Fprintf(f, `<div class="bookmark %s">Bookmark to %s</div>`+"\n", levelCls, block.Link)
	case notionapi.TypeGist:
		fmt.Fprintf(f, `<div class="gist %s">Gist for %s</div>`+"\n", levelCls, block.Source)
	case notionapi.TypeImage:
		link := block.ImageURL
		fmt.Fprintf(f, `<img class="%s" src="%s" />`+"\n", levelCls, link)
	case notionapi.TypeColumnList:
		// TODO: implement me
	case notionapi.TypeCollectionView:
		// TODO: implement me
	default:
		fmt.Printf("Unsupported block type '%s', id: %s\n", block.Type, block.ID)
		panic(fmt.Sprintf("Unsupported block type '%s'", block.Type))
	}
}

func genBlocksHTML(f io.Writer, parent *notionapi.Block, level int) {
	blocks := parent.Content
	for i, block := range blocks {
		if block == nil {
			id := parent.ContentIDs[i]
			fmt.Printf("No block at index %d with id=%s. Parent block %s of type %s\n", i, id, parent.ID, parent.Type)
		}
		genBlockHTML(f, block, level)
	}
}

type Metadata struct {
	ID          string
	Tags        []string
	Date        string
	DateParsed  time.Time
	Description string
}

// exttract metadata from blocks
func extractMetadata(pageInfo *notionapi.PageInfo) *Metadata {
	blocks := pageInfo.Page.Content
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	res := Metadata{}
	for len(blocks) > 0 {
		block := blocks[0]
		if block.Type != notionapi.TypeText {
			break
		}

		if len(block.InlineContent) > 0 {
			break
		}
		inline := block.InlineContent[0]
		// must be plain text
		if inline.AttrFlags != 0 || len(inline.Attrs) > 0 {
			break
		}
		// remove empty lines at the top
		s := strings.TrimSpace(inline.Text)
		if s == "" {
			blocks = blocks[1:]
			continue
		}
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
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
		case "id":
			res.ID = val
		case "date":
			res.Date = val
			// 2002-06-21T04:15:29-07:00
			parsed, err := time.Parse(time.RFC3339, res.Date)
			if err != nil {
				panicMsg("Failed to parse date '%s' in notion page with id '%s'. Error: %s", res.Date, pageInfo.ID, err)
			}
			res.DateParsed = parsed
		case "description":
			res.Description = val
		default:
			panicMsg("Unsupported tag '%s' in notion page with id '%s'", pageInfo.ID)
		}

	}
	pageInfo.Page.Content = blocks
	return &res
}

func panicMsg(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	fmt.Printf("%s\n", s)
	panic(s)
}

func genHTML(pageID string, pageInfo *notionapi.PageInfo) []byte {
	f := &bytes.Buffer{}
	title := pageInfo.Page.Title
	title = template.HTMLEscapeString(title)
	fmt.Fprintf(f, `<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/main.css" rel="stylesheet">
		<title>%s</title>
	</head>
	<body>`, title)

	page := pageInfo.Page
	genHTMLTitle(f, page)
	genBlocksHTML(f, page, 0)
	fmt.Fprintf(f, "</body>\n</html>\n")
	return f.Bytes()
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

func toHTML(pageID, path string) (*notionapi.PageInfo, error) {
	fmt.Printf("toHTML: pageID=%s, path=%s\n", pageID, path)
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	pageInfo, err := getPageInfoCached(pageID)
	if err != nil {
		fmt.Printf("getPageInfoCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	d := genHTML(pageID, pageInfo)
	err = ioutil.WriteFile(path, d, 0644)
	return pageInfo, err
}

func findSubPageIDs(blocks []*notionapi.Block) []string {
	var res []string
	for _, block := range blocks {
		if block.Type == notionapi.TypePage {
			res = append(res, block.ID)
		}
	}
	return res
}

func copyCSS() {
	src := filepath.Join("cmd", "tohtml", "main.css")
	dst := filepath.Join("www", "main.css")
	err := copyFile(dst, src)
	if err != nil {
		panic(err.Error())
	}
}

func importNotion() {
	os.MkdirAll("log", 0755)
	os.MkdirAll("cache", 0755)

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
		path := filepath.Join("www", name)
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
	//copyCSS()
}
