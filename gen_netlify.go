package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chilts/sid"
	"github.com/kjk/betterguid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	uuid "github.com/satori/go.uuid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
	atom "github.com/thomas11/atomgenerator"
)

func copyAndSortArticles(articles []*Article) []*Article {
	n := len(articles)
	res := make([]*Article, n, n)
	copy(res, articles)
	sort.Slice(res, func(i, j int) bool {
		return res[j].PublishedOn.After(res[i].PublishedOn)
	})
	return res
}

func genAtomXML(store *Articles, excludeNotes bool) ([]byte, error) {
	articles := store.getBlogNotHidden()
	if excludeNotes {
		articles = filterArticlesByTag(articles, "note", false)
	}
	articles = copyAndSortArticles(articles)
	n := 25
	if n > len(articles) {
		n = len(articles)
	}

	latest := make([]*Article, n, n)
	size := len(articles)
	for i := 0; i < n; i++ {
		latest[i] = articles[size-1-i]
	}

	pubTime := time.Now()
	if len(articles) > 0 {
		pubTime = articles[0].PublishedOn
	}

	feed := &atom.Feed{
		Title:   "Krzysztof Kowalczyk blog",
		Link:    "https://blog.kowalczyk.info/atom.xml",
		PubDate: pubTime,
	}

	for _, a := range latest {
		//id := fmt.Sprintf("tag:blog.kowalczyk.info,1999:%d", a.Id)
		e := &atom.Entry{
			Title:   a.Title,
			Link:    "https://blog.kowalczyk.info" + a.URL(),
			Content: a.BodyHTML,
			PubDate: a.PublishedOn,
		}
		feed.AddEntry(e)
	}

	return feed.GenXml()
}

func netlifyPath(fileName string) string {
	fileName = strings.TrimLeft(fileName, "/")
	path := filepath.Join("netlify_static", fileName)
	err := mkdirForFile(path)
	panicIfErr(err)
	return path
}

func netlifyWriteFile(fileName string, d []byte) {
	path := netlifyPath(fileName)
	//lg("%s\n", path)
	ioutil.WriteFile(path, d, 0644)
}

func netlifyRequestGetFullHost() string {
	return "https://blog.kowalczyk.info"
}

// https://www.linkedin.com/shareArticle?mini=true&;url=https://nodesource.com/blog/why-the-new-v8-is-so-damn-fast"
func makeLinkedinShareURL(article *Article) string {
	uri := netlifyRequestGetFullHost() + article.URL()
	uri = url.QueryEscape(uri)
	return fmt.Sprintf(`https://www.linkedin.com/shareArticle?mini=true&url=%s`, uri)
}

// https://www.facebook.com/sharer/sharer.php?u=https://nodesource.com/blog/why-the-new-v8-is-so-damn-fast
func makeFacebookShareURL(article *Article) string {
	uri := netlifyRequestGetFullHost() + article.URL()
	uri = url.QueryEscape(uri)
	return fmt.Sprintf(`https://www.facebook.com/sharer/sharer.php?u=%s`, uri)
}

// https://plus.google.com/share?url=https://nodesource.com/blog/why-the-new-v8-is-so-damn-fast
func makeGooglePlusShareURL(article *Article) string {
	uri := netlifyRequestGetFullHost() + article.URL()
	uri = url.QueryEscape(uri)
	return fmt.Sprintf(`https://plus.google.com/share?url=%s`, uri)
}

func makeTwitterShareURL(article *Article) string {
	title := url.QueryEscape(article.Title)
	uri := netlifyRequestGetFullHost() + article.URL()
	uri = url.QueryEscape(uri)
	return fmt.Sprintf(`https://twitter.com/intent/tweet?text=%s&url=%s&via=kjk`, title, uri)
}

// TagInfo represents a single tag for articles
type TagInfo struct {
	URL   string
	Name  string
	Count int
}

var (
	allTags []*TagInfo
)

func buildTags(articles []*Article) []*TagInfo {
	if allTags != nil {
		return allTags
	}

	var res []*TagInfo
	ti := &TagInfo{
		URL:   "/archives.html",
		Name:  "all",
		Count: len(articles),
	}
	res = append(res, ti)

	tagCounts := make(map[string]int)
	for _, a := range articles {
		for _, tag := range a.Tags {
			tagCounts[tag]++
		}
	}
	var tags []string
	for tag := range tagCounts {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		count := tagCounts[tag]
		ti = &TagInfo{
			URL:   "/tag/" + tag,
			Name:  tag,
			Count: count,
		}
		res = append(res, ti)
	}
	allTags = res
	return res
}

func netlifyWriteArticlesArchiveForTag(store *Articles, tag string) {
	path := "/archives.html"
	articles := store.getBlogNotHidden()
	if tag != "" {
		articles = filterArticlesByTag(articles, tag, true)
		// must manually resolve conflict due to urlify
		tagInPath := tag
		if tag == "c#" {
			tagInPath = "csharp"
		} else if tag == "c++" {
			tagInPath = "cplusplus"
		}
		tagInPath = urlify(tagInPath)
		path = fmt.Sprintf("/article/archives-by-tag-%s.html", tagInPath)
		from := "/tag/" + tag
		netlifyAddRewrite(from, path)
	}

	model := struct {
		AnalyticsCode string
		Article       *Article
		PostsCount    int
		Tag           string
		Years         []Year
		Tags          []*TagInfo
	}{
		AnalyticsCode: analyticsCode,
		PostsCount:    len(articles),
		Years:         buildYearsFromArticles(articles),
		Tag:           tag,
		Tags:          buildTags(articles),
	}

	netlifyExecTemplate(path, tmplArchive, model)
}

func skipTmplFiles(path string) bool {
	if strings.Contains(path, ".tmpl.") {
		return true
	}
	return false
}

func copyImages() {
	srcDir := filepath.Join("notion_cache", "img")
	dstDir := filepath.Join("netlify_static", "img")
	dirCopyRecur(dstDir, srcDir, nil)
}

func genIndex(store *Articles, w io.Writer) error {
	// /
	articles := store.getBlogNotHidden()
	if len(articles) > 5 {
		articles = articles[:5]
	}
	articleCount := len(articles)
	websiteIndexPage := store.idToArticle[notionWebsiteStartPage]
	model := struct {
		AnalyticsCode string
		Article       *Article
		Articles      []*Article
		ArticleCount  int
		WebsiteHTML   template.HTML
	}{
		AnalyticsCode: analyticsCode,
		Article:       nil, // always nil
		ArticleCount:  articleCount,
		Articles:      articles,
		WebsiteHTML:   websiteIndexPage.HTMLBody,
	}
	return execTemplate("/index.html", tmplMainPage, model, w)
}

func netlifyBuild(store *Articles) {
	// verify we're in the right directory
	_, err := os.Stat("netlify_static")
	panicIfErr(err)
	outDir := filepath.Join("netlify_static")
	err = os.RemoveAll(outDir)
	panicIfErr(err)
	err = os.MkdirAll(outDir, 0755)
	panicIfErr(err)
	nCopied, err := dirCopyRecur(outDir, "www", skipTmplFiles)
	panicIfErr(err)
	lg("Copied %d files\n", nCopied)

	netlifyAddStaticRedirects()
	netlifyAddRewrite("/favicon.ico", "/static/favicon.ico")
	//netlifyAddRewrite("/book/", "/static/documents.html")
	//netflifyAddTempRedirect("/book/*", "/article/:splat")
	netflifyAddTempRedirect("/software/sumatrapdf*", "https://www.sumatrapdfreader.org/:splat")

	netflifyAddTempRedirect("/articles/", "/documents.html")
	netflifyAddTempRedirect("/articles/index.html", "/documents.html")
	netflifyAddTempRedirect("/static/documents.html", "/documents.html")
	netflifyAddTempRedirect("/software/index.html", "/software/")

	{
		// url: /book/go-cookbook.html
		model := struct {
		}{}
		netlifyExecTemplate("/book/go-cookbook.html", tmplGoCookBook, model)
		netlifyAddRewrite("/articles/go-cookbook.html", "/book/go-cookbook.html")
	}

	{
		// url: /book/windows-programming-in-go.html
		model := struct {
		}{}
		netlifyExecTemplate("/book/go-cookbook.html", tmplGoCookBook, model)
		netlifyAddRewrite("/articles/go-cookbook.html", "/book/go-cookbook.html")
	}

	genIndex(store, nil)

	// TODO: maybe just use /archive.html
	{
		// /blogindex.html
		articles := store.getBlogNotHidden()
		articleCount := len(articles)
		model := struct {
			AnalyticsCode string
			Article       *Article
			Articles      []*Article
			ArticleCount  int
		}{
			AnalyticsCode: analyticsCode,
			Article:       nil, // always nil
			ArticleCount:  articleCount,
			Articles:      articles,
		}
		netlifyExecTemplate("/blogindex.html", tmplBlogIndex, model)
	}

	{
		// /changelog.html
		articles := append([]*Article{}, store.articles...)
		sort.Slice(articles, func(i, j int) bool {
			a1 := articles[i]
			a2 := articles[j]
			return a1.UpdatedOn.After(a2.UpdatedOn)
		})
		if len(articles) > 64 {
			articles = articles[:64]
		}
		prevAge := -1
		for _, a := range articles {
			age := a.UpdatedAge()
			if prevAge != age {
				a.UpdatedAgeStr = fmt.Sprintf("%d d", a.UpdatedAge())
			}
			prevAge = age
		}

		model := struct {
			AnalyticsCode string
			Article       *Article
			Articles      []*Article
		}{
			AnalyticsCode: analyticsCode,
			Article:       nil, // always nil
			Articles:      articles,
		}
		netlifyExecTemplate("/changelog.html", tmplChangelog, model)
	}

	copyImages()

	{
		// /atom.xml
		d, err := genAtomXML(store, true)
		panicIfErr(err)
		netlifyWriteFile("/atom.xml", d)
	}

	{
		// /atom-all.xml
		d, err := genAtomXML(store, false)
		panicIfErr(err)
		netlifyWriteFile("/atom-all.xml", d)
	}

	{
		// /blog/ and /kb/ are only for redirects, we only handle /article/ at this point
		logVerbose("%d articles\n", len(store.idToPage))
		for _, article := range store.articles {
			canonicalURL := netlifyRequestGetFullHost() + article.URL()
			model := struct {
				AnalyticsCode      string
				Article            *Article
				CanonicalURL       string
				CoverImage         string
				PageTitle          string
				TagsDisplay        string
				HeaderImageURL     string
				NotionEditURL      string
				Description        string
				TwitterShareURL    string
				FacebookShareURL   string
				LinkedInShareURL   string
				GooglePlusShareURL string
			}{
				AnalyticsCode:      analyticsCode,
				Article:            article,
				CanonicalURL:       canonicalURL,
				CoverImage:         article.HeaderImageURL,
				PageTitle:          article.Title,
				Description:        article.Description,
				TwitterShareURL:    makeTwitterShareURL(article),
				FacebookShareURL:   makeFacebookShareURL(article),
				LinkedInShareURL:   makeLinkedinShareURL(article),
				GooglePlusShareURL: makeGooglePlusShareURL(article),
			}
			if article.page != nil {
				id := normalizeID(article.page.ID)
				model.NotionEditURL = "https://notion.so/" + id
			}
			path := fmt.Sprintf("/article/%s.html", article.ID)
			logVerbose("%s => %s, %s, %s\n", article.ID, path, article.URL(), article.Title)
			netlifyExecTemplate(path, tmplArticle, model)
			if article.urlOverride != "" {
				//lg("url override: %s => %s\n", article.urlOverride, path)
				netlifyAddRewrite(article.urlOverride, path)
			}
		}
	}

	{
		// /archives.html
		netlifyWriteArticlesArchiveForTag(store, "")
		tags := map[string]struct{}{}
		for _, article := range store.getBlogNotHidden() {
			for _, tag := range article.Tags {
				tags[tag] = struct{}{}
			}
		}
		for tag := range tags {
			netlifyWriteArticlesArchiveForTag(store, tag)
		}
	}

	{
		// /sitemap.xml
		data, err := genSiteMap(store, "https://blog.kowalczyk.info")
		panicIfErr(err)
		netlifyWriteFile("/sitemap.xml", data)
	}

	{
		// /tools/generate-unique-id
		idXid := xid.New()
		idKsuid := ksuid.New()

		t := time.Now().UTC()
		entropy := rand.New(rand.NewSource(t.UnixNano()))
		idUlid := ulid.MustNew(ulid.Timestamp(t), entropy)
		betterGUID := betterguid.New()
		uuid := uuid.NewV4()

		flake := sonyflake.NewSonyflake(sonyflake.Settings{})
		sfid, err := flake.NextID()
		sfidstr := fmt.Sprintf("%x", sfid)
		if err != nil {
			sfidstr = err.Error()
		}

		model := struct {
			Xid           string
			Ksuid         string
			Ulid          string
			BetterGUID    string
			Sonyflake     string
			Sid           string
			UUIDv4        string
			AnalyticsCode string
		}{
			Xid:           idXid.String(),
			Ksuid:         idKsuid.String(),
			Ulid:          idUlid.String(),
			BetterGUID:    betterGUID,
			Sonyflake:     sfidstr,
			Sid:           sid.Id(),
			UUIDv4:        uuid.String(),
			AnalyticsCode: analyticsCode,
		}

		// make sure /tools/generate-unique-id is served as html
		path := "/tools/generate-unique-id.html"
		netlifyExecTemplate(path, tmplGenerateUniqueID, model)
		netlifyAddRewrite("/tools/generate-unique-id", path)
	}

	// /ping
	netlifyWriteFile("/ping", []byte("pong"))

	// no longer care about /worklog

	netlifyAddArticleRedirects(store)
	netlifyWriteRedirects()
	writeCaddyConfig()
}
