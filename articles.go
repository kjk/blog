package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
)

var (
	notionBlogsStartPage      = "300db9dc27c84958a08b8d0c37f4cfe5"
	notionWebsiteStartPage    = "568ac4c064c34ef6a6ad0b8d77230681"
	notionGoCookbookStartPage = "7495260a1daa46118858ad2e049e77e6"
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

// Article describes a single article
type Article struct {
	ID             string
	PublishedOn    time.Time
	UpdatedOn      time.Time
	Title          string
	Tags           []string
	BodyHTML       string
	HTMLBody       template.HTML
	HeaderImageURL string
	Collection     string
	CollectionURL  string
	Status         int
	Description    string
	Paths          []URLPath
	urlOverride    string

	UpdatedAgeStr string
	Images        []ImageMapping

	// if true, this belongs to blog i.e. will be present in atom.xml
	// and listed in blog section
	inBlog bool

	page *notionapi.Page
}

// Articles has info about all articles downloaded from notion
type Articles struct {
	idToArticle map[string]*Article
	idToPage    map[string]*notionapi.Page
	// all downloaded articles
	articles []*Article
	// articles that are not hidden
	articlesNotHidden []*Article
	// articles that belong to a blog
	blog []*Article
	// blog articles that are not hidden
	blogNotHidden []*Article
}

func (a *Articles) getNotHidden() []*Article {
	if a.articlesNotHidden == nil {
		var arr []*Article
		for _, article := range a.articles {
			if !article.IsHidden() {
				arr = append(arr, article)
			}
		}
		a.articlesNotHidden = arr
	}
	return a.articlesNotHidden
}

func (a *Articles) getBlogNotHidden() []*Article {
	if a.blogNotHidden == nil {
		var arr []*Article
		for _, article := range a.blog {
			if !article.IsHidden() {
				arr = append(arr, article)
			}
		}
		a.blogNotHidden = arr
	}
	return a.blogNotHidden
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

func setStatusMust(article *Article, val string) {
	var err error
	article.Status, err = parseStatus(val)
	panicIfErr(err)
}

func setCollectionMust(article *Article, val string) {
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
	article.Collection = val
	article.CollectionURL = collectionURL

}
func setHeaderImageMust(article *Article, val string) {
	if val[0] != '/' {
		val = "/" + val
	}
	path := filepath.Join("www", val)
	panicIf(!u.FileExists(path), "File '%s' for @header-image doesn't exist", path)
	//fmt.Printf("Found HeaderImageURL: %s\n", fileName)
	uri := netlifyRequestGetFullHost() + val
	article.HeaderImageURL = uri
}

func notionPageToArticle(c *notionapi.Client, page *notionapi.Page) *Article {
	blocks := page.Root.Content
	//fmt.Printf("extractMetadata: %s-%s, %d blocks\n", title, id, len(blocks))
	// metadata blocks are always at the beginning. They are TypeText blocks and
	// have only one plain string as content
	root := page.Root
	title := root.Title
	id := normalizeID(root.ID)
	article := &Article{
		page:  page,
		Title: title,
	}
	nBlock := 0
	var err error
	endLoop := false

	article.PublishedOn = root.CreatedOn()
	article.UpdatedOn = root.UpdatedOn()
	var publishedOnOverwrite time.Time

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
			// PublishedOn over-writes Date and CreatedAt
			publishedOnOverwrite, err = parseDate(val)
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
		case "url":
			article.urlOverride = val
		default:
			// assume that unrecognized meta means this article doesn't have
			// proper meta tags. It might miss meta-tags that are badly named
			endLoop = true
			/*
				rmCached(page.ID)
				title := page.Page.Title
				panicMsg("Unsupported meta '%s' in notion page with id '%s', '%s'", key, normalizeID(page.ID), title)
			*/
		}
		if endLoop {
			break
		}
		blocks = blocks[1:]
		nBlock++
	}
	root.Content = blocks

	if !publishedOnOverwrite.IsZero() {
		article.PublishedOn = publishedOnOverwrite
	}

	if article.ID == "" {
		article.ID = id
	}

	if article.Collection != "" {
		path := URLPath{
			Name: article.Collection,
			URL:  article.CollectionURL,
		}
		article.Paths = append(article.Paths, path)
	}

	format := root.FormatPage
	// set image header from cover page
	if article.HeaderImageURL == "" && format != nil && format.PageCoverURL != "" {
		path, err := downloadAndCacheImage(c, format.PageCoverURL)
		panicIfErr(err)
		relURL := "/img/" + filepath.Base(path)
		im := ImageMapping{
			path:        path,
			relativeURL: relURL,
		}
		article.Images = append(article.Images, im)
		article.HeaderImageURL = relURL
	}
	return article
}

func articleSetID(a *Article, v string) {
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

func addIDToBlock(block *notionapi.Block, idToBlock map[string]*notionapi.Block) {
	id := normalizeID(block.ID)
	idToBlock[id] = block
	for _, block := range block.Content {
		if block == nil {
			continue
		}
		addIDToBlock(block, idToBlock)
	}
}

func buildArticleNavigation(article *Article, isRootPage func(string) bool, idToBlock map[string]*notionapi.Block) {
	// some already have path (e.g. those that belong to a collection)
	if len(article.Paths) > 0 {
		return
	}

	page := article.page.Root
	currID := normalizeID(page.ParentID)

	var paths []URLPath
	for !isRootPage(currID) {
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

	// set in reverse order
	n := len(paths)
	for i := 1; i <= n; i++ {
		path := paths[n-i]
		article.Paths = append(article.Paths, path)
	}
}

// build navigation bread-crumbs for articles
func buildArticlesNavigation(articles *Articles) {
	idToBlock := map[string]*notionapi.Block{}
	for _, a := range articles.articles {
		page := a.page
		if page == nil {
			continue
		}
		addIDToBlock(page.Root, idToBlock)
	}

	isRoot := func(id string) bool {
		id = normalizeID(id)
		switch id {
		case notionBlogsStartPage, notionWebsiteStartPage, notionGoCookbookStartPage:
			return true
		}
		return false
	}

	for _, article := range articles.articles {
		buildArticleNavigation(article, isRoot, idToBlock)
	}
}

func loadArticles(c *notionapi.Client) *Articles {
	res := &Articles{}
	startIDs := []string{notionWebsiteStartPage}
	res.idToPage = loadAllPages(c, startIDs, useCacheForNotion)

	res.idToArticle = map[string]*Article{}
	for id, page := range res.idToPage {
		panicIf(id != normalizeID(id), "bad id '%s' sneaked in", id)
		article := notionPageToArticle(c, page)
		if article.urlOverride != "" {
			fmt.Printf("url override: %s => %s\n", article.urlOverride, article.ID)
		}
		res.idToArticle[id] = article
		// this might be legacy, short id. If not, we just set the same value twice
		articleID := article.ID
		res.idToArticle[articleID] = article
		if article.IsBlog() {
			res.blog = append(res.blog, article)
		}
		res.articles = append(res.articles, article)
	}

	for _, article := range res.articles {
		html, images := notionToHTML(c, article.page, res)
		article.BodyHTML = string(html)
		article.HTMLBody = template.HTML(article.BodyHTML)
		article.Images = append(article.Images, images...)
	}

	buildArticlesNavigation(res)

	sort.Slice(res.blog, func(i, j int) bool {
		return res.blog[i].PublishedOn.After(res.blog[j].PublishedOn)
	})

	return res
}

// MonthArticle combines article and a month
type MonthArticle struct {
	*Article
	DisplayMonth string
}

// Year describes articles in a given year
type Year struct {
	Name     string
	Articles []MonthArticle
}

// DisplayTitle returns a title for an article
func (a *MonthArticle) DisplayTitle() string {
	if a.Title != "" {
		return a.Title
	}
	return "no title"
}

// NewYear creates a new Year
func NewYear(name string) *Year {
	return &Year{Name: name, Articles: make([]MonthArticle, 0)}
}

func buildYearsFromArticles(articles []*Article) []Year {
	res := make([]Year, 0)
	var currYear *Year
	var currMonthName string
	n := len(articles)
	for i := 0; i < n; i++ {
		a := articles[i]
		yearName := a.PublishedOn.Format("2006")
		if currYear == nil || currYear.Name != yearName {
			if currYear != nil {
				res = append(res, *currYear)
			}
			currYear = NewYear(yearName)
			currMonthName = ""
		}
		ma := MonthArticle{Article: a}
		monthName := a.PublishedOn.Format("01")
		if monthName != currMonthName {
			ma.DisplayMonth = a.PublishedOn.Format("January 2")
		} else {
			ma.DisplayMonth = a.PublishedOn.Format("2")
		}
		currMonthName = monthName
		currYear.Articles = append(currYear.Articles, ma)
	}
	if currYear != nil {
		res = append(res, *currYear)
	}
	return res
}

func filterArticlesByTag(articles []*Article, tag string, include bool) []*Article {
	res := make([]*Article, 0)
	for _, a := range articles {
		hasTag := false
		for _, t := range a.Tags {
			if tag == t {
				hasTag = true
				break
			}
		}
		if include && hasTag {
			res = append(res, a)
		} else if !include && !hasTag {
			res = append(res, a)
		}
	}
	return res
}
