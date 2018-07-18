package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
	"github.com/kr/fs"
)

// for Article.Status
const (
	statusNormal       = iota // show on main page
	statusDraft               // not shown in production but shown in dev
	statusNotImportant        // linked from archive page, but not main page
	statusHidden              // not linked from any page but accessible via url
	statusDeleted             // not shown at all
)

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

	HTMLBody template.HTML

	pageInfo *notionapi.PageInfo
}

// URL returns article's permalink
func (a *Article) URL() string {
	return "/article/" + a.ID + "/" + urlify(a.Title) + ".html"
}

// IsDraft returns true if article is a draft
func (a *Article) IsDraft() bool {
	return a.Status == statusDraft
}

// DebugIsNotImportant returns true if article is not important and we're previewing locally
func (a *Article) DebugIsNotImportant() bool {
	return !inProduction && (a.Status == statusNotImportant)
}

// DebugIsHidden returns true if article is not important and we're previewing locally
func (a *Article) DebugIsHidden() bool {
	return !inProduction && (a.Status == statusHidden)
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

// ArticlesStore is a store for articles
type ArticlesStore struct {
	articles    []*Article
	idToArticle map[string]*Article
}

func isSepLine(s string) bool {
	return strings.HasPrefix(s, "---")
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

func extractMetadataValue(d []byte, prefix string) ([]byte, string) {
	eolIdx := bytes.IndexByte(d, '\n')
	if eolIdx == -1 || eolIdx < len(prefix) {
		return d, ""
	}
	maybePrefix := strings.ToLower(string(d[:len(prefix)]))
	if maybePrefix != prefix {
		return d, ""
	}
	val := d[len(prefix):eolIdx]
	d = d[eolIdx+1:]
	return d, strings.TrimSpace(string(val))
}

func parseStatus(status string) (int, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return statusNormal, nil
	}
	switch status {
	case "hidden":
		return statusHidden, nil
	case "draft":
		return statusDraft, nil
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
		collectionURL = "/book/windows-programming-in-go.html"
		val = "Windows Programming In Go"
	}
	panicIf(collectionURL == "", "'%s' is now a known collection", val)
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

// might return nil if article is meant to be skipped (deleted or draft)
func readArticle(path string) (*Article, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	a := &Article{}
	r := bufio.NewReader(f)
	var publishedOn time.Time
	atBeginning := true
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		l = strings.TrimSpace(l)
		if isSepLine(l) {
			if atBeginning {
				continue
			}
			break
		}
		atBeginning = false
		parts := strings.SplitN(l, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Unexpected line: %q", l)
		}
		k := strings.ToLower(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "status":
			a.Status, err = parseStatus(v)
			if err != nil {
				return nil, err
			}
		case "id":
			articleSetID(a, v)
			a.OrigPath = path
		case "title":
			a.Title = v
		case "tags":
			a.Tags = parseTags(v)
		case "publishedon":
			publishedOn, err = parseDate(v)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid PublishedOn date", v)
			}
		case "date", "createdat":
			a.PublishedOn, err = parseDate(v)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid date", v)
			}
		case "updatedat":
			a.UpdatedOn, err = parseDate(v)
		case "headerimage":
			setHeaderImageMust(a, v)
		case "collection":
			setCollectionMust(a, v)
		case "description":
			a.Description = v
		case "format":
			v := strings.TrimSpace(strings.ToLower(v))
			switch v {
			case "markdown", "md":
				// do nothing
			default:
				return nil, fmt.Errorf("Unknown format '%s'", v)
			}
		default:
			return nil, fmt.Errorf("Unexpected key: %q", k)
		}
	}

	// if deleted, act as doesn't exist
	if a.Status == statusDeleted {
		return nil, nil
	}

	// PublishedOn over-writes Date and CreatedAt
	if !publishedOn.IsZero() {
		a.PublishedOn = publishedOn
	}

	if a.UpdatedOn.IsZero() {
		a.UpdatedOn = a.PublishedOn
	}

	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	a.Body = d
	a.BodyHTML = markdownToHTML(a.Body, "")
	a.HTMLBody = template.HTML(a.BodyHTML)
	return a, nil
}

func readArticles() ([]*Article, []string, error) {
	dirsToScan := []string{
		filepath.Join("books", "windows-programming-in-go"),
	}
	var allArticles []*Article
	var allDirs []string
	timeStart := time.Now()
	for _, dir := range dirsToScan {
		articles, dirs, err := readArticlesFromDir(dir)
		if err != nil {
			return nil, nil, err
		}
		allArticles = append(allArticles, articles...)
		allDirs = append(allDirs, dirs...)
	}
	fmt.Printf("read %d articles in %s\n", len(allArticles), time.Since(timeStart))
	return allArticles, allDirs, nil
}

func readArticlesFromDir(dir string) ([]*Article, []string, error) {
	walker := fs.Walk(dir)
	var res []*Article
	var dirs []string
	for walker.Step() {
		if walker.Err() != nil {
			fmt.Printf("readArticles: walker.Step() failed with %s\n", walker.Err())
			return nil, nil, walker.Err()
		}
		st := walker.Stat()
		path := walker.Path()
		if st.IsDir() {
			dirs = append(dirs, path)
			continue
		}
		name := filepath.Base(path)
		if name == "notes.txt" {
			continue
		}

		a, err := readArticle(path)
		if err != nil {
			fmt.Printf("readArticle() of %s failed with %s\n", path, err)
			return nil, nil, err
		}
		if a != nil {
			res = append(res, a)
		}
	}
	return res, dirs, nil
}

// NewArticlesStore returns a store of articles
func NewArticlesStore() (*ArticlesStore, error) {
	articles, _, err := readArticles()
	if err != nil {
		return nil, err
	}
	articles2 := loadArticlesFromNotion()
	if len(articles2) > 0 {
		articles = append(articles, articles2...)
	}
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PublishedOn.After(articles[j].PublishedOn)
	})

	res := &ArticlesStore{
		articles: articles,
	}
	res.idToArticle = make(map[string]*Article)
	for _, a := range articles {
		curr := res.idToArticle[a.ID]
		if curr != nil {
			log.Fatalf("2 articles with the same id %s\n%s\n%s\n", a.ID, curr.OrigPath, a.OrigPath)
		}
		res.idToArticle[a.ID] = a
	}
	return res, nil
}

const (
	articlesNormal          = 0
	articlesWithLessVisible = 1
	articlesWithHidden      = 2
)

func isNormal(a *Article) bool {
	if a.Status == statusNormal {
		return true
	}
	return false
}

func shouldGetArticle(a *Article, typ int) bool {
	if typ == articlesNormal {
		return isNormal(a)
	}

	if typ == articlesWithLessVisible {
		return isNormal(a) || (a.Status == statusNotImportant)
	}

	panicIf(typ != articlesWithHidden, "unknown typ: %d", typ)
	return isNormal(a) || (a.Status == statusNotImportant) || (a.Status == statusHidden)
}

// GetArticles returns articles of a given type
func (s *ArticlesStore) GetArticles(typ int) []*Article {
	var res []*Article
	for _, a := range s.articles {
		if shouldGetArticle(a, typ) {
			res = append(res, a)
		}
	}
	return res
}

// GetArticleByID returns an article given its id
func (s *ArticlesStore) GetArticleByID(id string) *Article {
	return s.idToArticle[id]
}

// TODO: this is simplistic but works for me, http://net.tutsplus.com/tutorials/other/8-regular-expressions-you-should-know/
// has more elaborate regex for extracting urls
var urlRx = regexp.MustCompile(`https?://[[:^space:]]+`)
var notURLEndChars = []byte(".),")

func notURLEndChar(c byte) bool {
	return -1 != bytes.IndexByte(notURLEndChars, c)
}

var disableUrlization = false

func strToHTML(s string) string {
	matches := urlRx.FindAllStringIndex(s, -1)
	if nil == matches || disableUrlization {
		s = template.HTMLEscapeString(s)
		s = strings.Replace(s, "\n", "<br>", -1)
		return "<p>" + s + "</p>"
	}

	urlMap := make(map[string]string)
	ns := ""
	prevEnd := 0
	for n, match := range matches {
		start, end := match[0], match[1]
		for end > start && notURLEndChar(s[end-1]) {
			end--
		}
		url := s[start:end]
		ns += s[prevEnd:start]

		// placeHolder is meant to be an unlikely string that doesn't exist in
		// the message, so that we can replace the string with it and then
		// revert the replacement. A more robust approach would be to remember
		// offsets
		placeHolder, ok := urlMap[url]
		if !ok {
			placeHolder = fmt.Sprintf("a;dfsl;a__lkasjdfh1234098;lajksdf_%d", n)
			urlMap[url] = placeHolder
		}
		ns += placeHolder
		prevEnd = end
	}
	ns += s[prevEnd:len(s)]

	ns = template.HTMLEscapeString(ns)
	for url, placeHolder := range urlMap {
		url = fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, url, url)
		ns = strings.Replace(ns, placeHolder, url, -1)
	}
	ns = strings.Replace(ns, "\n", "<br>", -1)
	return "<p>" + ns + "</p>"
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
