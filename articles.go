package main

import (
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
)

var (
	// maps notion page ID to *Article
	notionIDToArticle map[string]*Article
	// maps id to *Article. id can either be notion page ID
	// or shorter, legacy id
	idToArticle   map[string]*Article
	storeArticles []*Article
	blogArticles  []*Article
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
	OrigID         string
	PublishedOn    time.Time
	UpdatedOn      time.Time
	Title          string
	Tags           []string
	OrigPath       string // path of the markdown file with content
	Body           []byte
	BodyHTML       string
	HeaderImageURL string
	Collection     string
	CollectionURL  string
	Status         int
	Description    string
	Paths          []URLPath
	// if true, this belongs to blog i.e. will be present in atom.xml
	// and listed in blog section
	inBlog bool

	HTMLBody template.HTML

	pageInfo *notionapi.PageInfo
}

// URL returns article's permalink
func (a *Article) URL() string {
	return "/article/" + a.ID + "/" + urlify(a.Title) + ".html"
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
	article.HeaderImageURL = val
}

func articleSetID(a *Article, v string) {
	// we handle 2 types of ids:
	// - blog posts from articles/ directory have integer id
	// - blog posts imported from quicknotes have id that are strings
	a.OrigID = strings.TrimSpace(v)
	a.ID = a.OrigID
	id, err := strconv.Atoi(a.ID)
	if err == nil {
		a.ID = u.EncodeBase64(id)
	}
}

func loadAllArticles() {
	notionIDToArticle = make(map[string]*Article)

	{
		articles := loadNotionPages(notionBlogsStartPage)
		fmt.Printf("Loaded %d blog articles\n\n", len(articles))
	}

	{
		articles := loadNotionPages(notionGoCookbookStartPage)
		fmt.Printf("Loaded %d go cookbook articles\n\n", len(articles))
	}

	{
		articles := loadNotionPages(notionWebsiteStartPage)
		fmt.Printf("Loaded %d website articles\n", len(articles))
	}

	for _, a := range notionIDToArticle {
		if a.IsBlog() {
			blogArticles = append(blogArticles, a)
		}
	}
	sort.Slice(blogArticles, func(i, j int) bool {
		return blogArticles[i].PublishedOn.After(blogArticles[j].PublishedOn)
	})

	var articles []*Article
	for _, a := range notionIDToArticle {
		articles = append(articles, a)
	}

	storeArticles = articles

	panicIf(idToArticle != nil, "idToArticle not nil")
	idToArticle = make(map[string]*Article)
	for _, a := range articles {
		curr := idToArticle[a.ID]
		if curr != nil {
			log.Fatalf("2 articles with the same id %s\n%s\n%s\n", a.ID, curr.OrigPath, a.OrigPath)
		}
		idToArticle[a.ID] = a
	}
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
