package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
)

// for Article.Status
const (
	statusNormal       = iota // show on main page
	statusNotImportant        // linked from archive page, but not main page
	statusHidden              // not linked from any page but accessible via url
	statusDeleted             // not shown at all
)

// URLPath describes
type URLPath struct {
	URL  string
	Name string
}

// MetaValue represents a key/value metadata
type MetaValue struct {
	key   string
	value string
}

// ImageMapping keeps track of rewritten image urls (locally cached
// images in notion)
type ImageMapping struct {
	path        string
	relativeURL string
}

type BlockInfo struct {
	shouldSkip bool
	imageURL   string
}

// Article describes a single article
type Article struct {
	page *notionapi.Page

	ID                   string
	PublishedOn          time.Time
	UpdatedOn            time.Time
	Title                string
	Tags                 []string
	BodyHTML             string
	HTMLBody             template.HTML
	HeaderImageURL       string
	Collection           string
	CollectionURL        string
	Status               int
	Description          string
	Paths                []URLPath
	Metadata             []*MetaValue
	urlOverride          string
	publishedOnOverwrite time.Time

	// if true, this belongs to blog i.e. will be present in atom.xml
	// and listed in blog section
	inBlog bool

	UpdatedAgeStr string
	Images        []ImageMapping

	blockInfos map[*notionapi.Block]*BlockInfo
}

// URL returns article's permalink
func (a *Article) URL() string {
	if a.urlOverride != "" {
		return a.urlOverride
	}
	return "/article/" + a.ID + "/" + urlify(a.Title) + ".html"
}

// PathAsText returns navigation path as text
func (a *Article) PathAsText() string {
	paths := []string{"Home"}
	for _, urlpath := range a.Paths {
		paths = append(paths, urlpath.Name)
	}
	return strings.Join(paths, " / ")
}

// TagsDisplay returns tags as html
func (a *Article) TagsDisplay() template.HTML {
	arr := make([]string, 0)
	for _, tag := range a.Tags {
		// TODO: url-quote the first tag
		escapedURL := fmt.Sprintf(`<a href="/tag/%s" class="taglink">%s</a>`, tag, tag)
		arr = append(arr, escapedURL)
	}
	s := strings.Join(arr, ", ")
	return template.HTML(s)
}

// PublishedOnShort is a short version of date
func (a *Article) PublishedOnShort() string {
	return a.PublishedOn.Format("Jan 2 2006")
}

// IsBlog returns true if this article belongs to a blog
func (a *Article) IsBlog() bool {
	return a.inBlog
}

// UpdatedAge returns when it was updated last, in days
func (a *Article) UpdatedAge() int {
	dur := time.Since(a.UpdatedOn)
	return int(dur / (time.Hour * 24))
}

// IsHidden returns true if article should not be shown in the index
func (a *Article) IsHidden() bool {
	return a.Status == statusHidden || a.Status == statusDeleted || a.Status == statusNotImportant
}

func (a *Article) getBlockInfo(block *notionapi.Block) *BlockInfo {
	bi := a.blockInfos[block]
	if bi == nil {
		bi = &BlockInfo{}
		a.blockInfos[block] = bi
	}
	return bi
}

func (a *Article) markBlockToSkip(block *notionapi.Block) {
	a.getBlockInfo(block).shouldSkip = true
}

func (a *Article) shouldSkipBlock(block *notionapi.Block) bool {
	bi := a.blockInfos[block]
	if bi == nil {
		return false
	}
	return bi.shouldSkip
}

func (a *Article) setImageBlockURL(block *notionapi.Block, uri string) {
	a.getBlockInfo(block).imageURL = uri
}

func (a *Article) getImageBlockURL(block *notionapi.Block) string {
	bi := a.blockInfos[block]
	if bi == nil {
		return ""
	}
	return bi.imageURL
}

func (a *Article) removeEmptyTextBlocksAtEnd(root *notionapi.Block) {
	n := len(root.Content)
	blocks := root.Content
	for i := 0; i < n; i++ {
		idx := n - 1 - i
		block := blocks[idx]
		if !isEmptyTextBlock(block) {
			return
		}
		a.markBlockToSkip(block)
	}
}

func parseTags(s string) []string {
	tags := strings.Split(s, ",")
	var res []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		// skip the tag I use in quicknotes.io to tag notes for the blog
		if tag == "for-blog" || tag == "published" || tag == "draft" {
			continue
		}
		res = append(res, tag)
	}
	return res
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	// TODO: more formats?
	return time.Now(), err
}

func parseStatus(status string) (int, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return statusNormal, nil
	}
	switch status {
	case "hidden":
		return statusHidden, nil
	case "notimportant":
		return statusNotImportant, nil
	case "deleted":
		return statusDeleted, nil
	default:
		return 0, fmt.Errorf("'%s' is not a valid status", status)
	}
}

func isEmptyTextBlock(b *notionapi.Block) bool {
	if b.Type != notionapi.BlockText {
		return false
	}
	if len(b.InlineContent) > 0 {
		return false
	}
	return true
}

func (a *Article) SetID(v string) {
	// we handle 3 types of ids:
	// - blog posts from articles/ directory have integer id
	// - blog posts imported from quicknotes have id that are strings
	// - articles written in notion, have notion string id
	a.ID = strings.TrimSpace(v)
	id, err := strconv.Atoi(a.ID)
	if err == nil {
		a.ID = u.EncodeBase64(id)
	}
}

func (a *Article) setStatusMust(val string) {
	var err error
	a.Status, err = parseStatus(val)
	panicIfErr(err)
}

func (a *Article) setCollectionMust(val string) {
	collectionURL := ""
	switch val {
	case "go-cookbook":
		collectionURL = "/book/go-cookbook.html"
		val = "Go Cookbook"
	case "go-windows":
		// ignore
		return
	}
	panicIf(collectionURL == "", "'%s' is not a known collection", val)
	a.Collection = val
	a.CollectionURL = collectionURL

}

func (a *Article) setHeaderImageMust(val string) {
	if val[0] != '/' {
		val = "/" + val
	}
	path := filepath.Join("www", val)
	panicIf(!u.FileExists(path), "File '%s' for @header-image doesn't exist", path)
	uri := netlifyRequestGetFullHost() + val
	// fmt.Printf("Found HeaderImageURL: %s\n", uri)
	a.HeaderImageURL = uri
}

func getInlineBlocksText(blocks []*notionapi.InlineBlock) string {
	s := ""
	for _, b := range blocks {
		s += b.Text
	}
	return s
}

// parse:
// #url ${url}
// followed by an image block
func (a *Article) maybeParseImageURL(nBlock int, block *notionapi.Block) bool {
	if block.Type != notionapi.BlockText {
		return false
	}
	blocks := a.page.Root.Content
	if nBlock == len(blocks)-1 {
		return false
	}
	nextBlock := blocks[nBlock+1]
	if nextBlock.Type != notionapi.BlockImage {
		return false
	}
	s := getInlineBlocksText(block.InlineContent)
	s = strings.TrimSpace(s)
	uri := strings.TrimPrefix(s, "#url")
	if uri == s {
		return false
	}
	uri = strings.TrimSpace(uri)
	a.markBlockToSkip(block)
	a.setImageBlockURL(nextBlock, uri)
	return false
}

func (a *Article) maybeParseMeta(nBlock int, block *notionapi.Block) bool {
	var err error

	if block.Type != notionapi.BlockText {
		logTemp("extractMetadata: ending look because block %d is of type %s\n", nBlock, block.Type)
		return false
	}

	if len(block.InlineContent) == 0 {
		logTemp("block %d of type %s and has no InlineContent\n", nBlock, block.Type)
		return true
	} else {
		logTemp("block %d has %d InlineContent\n", nBlock, len(block.InlineContent))
	}

	inline := block.InlineContent[0]
	// must be plain text
	if !inline.IsPlain() {
		logTemp("block: %d of type %s: inline has attributes\n", nBlock, block.Type)
		return false
	}

	// remove empty lines at the top
	s := strings.TrimSpace(inline.Text)
	if s == "" {
		logTemp("block: %d of type %s: inline.Text is empty\n", nBlock, block.Type)
		return true
	}
	logTemp("  %d %s '%s'\n", nBlock, block.Type, s)

	// parse generic metadata like "@foo: bar" or "@foo bar"
	if s[0] == '@' {
		s := s[1:]
		idx := strings.Index(s, ":")
		if idx == -1 {
			idx = strings.Index(s, " ")
		}
		key := s
		value := ""
		if idx != -1 {
			key = s[:idx]
			value = s[idx+1:]
		}
		meta := &MetaValue{
			key:   key,
			value: value,
		}
		a.Metadata = append(a.Metadata, meta)
		return true
	}

	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		logTemp("block: %d of type %s: inline.Text is not key/value. s='%s'\n", nBlock, block.Type, s)
		return false
	}
	key := strings.ToLower(strings.TrimSpace(parts[0]))
	val := strings.TrimSpace(parts[1])
	switch key {
	case "tags":
		a.Tags = parseTags(val)
		logTemp("Tags: %v\n", a.Tags)
	case "id":
		a.SetID(val)
		logTemp("ID: %s\n", a.ID)
	case "publishedon":
		// PublishedOn over-writes Date and CreatedAt
		a.publishedOnOverwrite, err = parseDate(val)
		panicIfErr(err)
		a.inBlog = true
		logTemp("got publishedon")
	case "date", "createdat":
		a.PublishedOn, err = parseDate(val)
		panicIfErr(err)
		a.inBlog = true
		logTemp("got date or createdat")
	case "updatedat":
		a.UpdatedOn, err = parseDate(val)
		panicIfErr(err)
	case "status":
		a.setStatusMust(val)
	case "description":
		a.Description = val
		logTemp("Description: %s\n", a.Description)
	case "headerimage":
		a.setHeaderImageMust(val)
	case "collection":
		a.setCollectionMust(val)
	case "url":
		a.urlOverride = val
	default:
		// assume that unrecognized meta means this article doesn't have
		// proper meta tags. It might miss meta-tags that are badly named
		return false
		/*
			rmCached(page.ID)
			title := page.Page.Title
			panicMsg("Unsupported meta '%s' in notion page with id '%s', '%s'", key, normalizeID(page.ID), title)
		*/
	}
	return true
}

func notionPageToArticle(c *notionapi.Client, page *notionapi.Page) *Article {
	blocks := page.Root.Content
	//fmt.Printf("extractMetadata: %s-%s, %d blocks\n", title, id, len(blocks))
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	root := page.Root
	title := root.Title
	id := normalizeID(root.ID)
	a := &Article{
		page:       page,
		Title:      title,
		blockInfos: map[*notionapi.Block]*BlockInfo{},
	}

	// allow debugging for specific pages
	if true && id == "39a15945117440d99a9ef0f7de1b618a" {
		doTempLog = true
		defer func() {
			doTempLog = false
		}()
		logTemp("Temp logging article %s %s\n", id, title)
	}

	a.PublishedOn = root.CreatedOn()
	a.UpdatedOn = root.UpdatedOn()

	parsingMeta := true
	for nBlock, block := range blocks {
		logTemp("  %d %s '%s'\n", nBlock, block.Type, block.Title)

		if parsingMeta {
			parsingMeta = a.maybeParseMeta(nBlock, block)
			if parsingMeta {
				a.markBlockToSkip(block)
			}
		}
		if !parsingMeta {
			a.maybeParseImageURL(nBlock, block)
		}
	}

	if !a.publishedOnOverwrite.IsZero() {
		a.PublishedOn = a.publishedOnOverwrite
	}

	if a.ID == "" {
		a.ID = id
	}

	if a.Collection != "" {
		path := URLPath{
			Name: a.Collection,
			URL:  a.CollectionURL,
		}
		a.Paths = append(a.Paths, path)
	}

	format := root.FormatPage
	// set image header from cover page
	if a.HeaderImageURL == "" && format != nil && format.PageCover != "" {
		path, err := downloadAndCacheImage(c, format.PageCover)
		panicIfErr(err)
		relURL := "/img/" + filepath.Base(path)
		im := ImageMapping{
			path:        path,
			relativeURL: relURL,
		}
		a.Images = append(a.Images, im)
		uri := netlifyRequestGetFullHost() + relURL
		a.HeaderImageURL = uri
	}

	a.removeEmptyTextBlocksAtEnd(page.Root)
	return a
}
