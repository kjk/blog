package main

import (
	"fmt"
	"html/template"
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

func genAtomXML(excludeNotes bool) ([]byte, error) {
	articles := store.GetArticles(articlesNormal)
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
	//fmt.Printf("%s\n", path)
	ioutil.WriteFile(path, d, 0644)
}

func netlifyRequestGetFullHost() string {
	return "https://blog.kowalczyk.info"
}

func makeShareHTML(article *Article) string {
	title := url.QueryEscape(article.Title)
	uri := netlifyRequestGetFullHost() + article.URL()
	uri = url.QueryEscape(uri)
	shareURL := fmt.Sprintf(`https://twitter.com/intent/tweet?text=%s&url=%s&via=kjk`, title, uri)
	followURL := `https://twitter.com/intent/follow?user_id=3194001`
	return fmt.Sprintf(`Hey there. You've read the whole thing. Let others know about this article by <a href="%s">sharing on Twitter</a>. <br>To be notified about new articles, <a href="%s">follow @kjk</a> on Twitter.`, shareURL, followURL)
}

func getArticleByID(articleID string) *Article {
	articles := store.GetArticles(articlesWithHidden)
	for _, a := range articles {
		if a.ID == articleID {
			return a
		}
	}
	return nil
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

func netlifyWriteArticlesArchiveForTag(tag string) {
	path := "/archives.html"
	articles := store.GetArticles(articlesWithLessVisible)
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
		path = fmt.Sprintf("/blog/archives-by-tag-%s.html", tagInPath)
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

func netlifyBuild() {
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
	fmt.Printf("Copied %d files\n", nCopied)

	netlifyAddStaticRedirects()
	netlifyAddRewrite("/favicon.ico", "/static/favicon.ico")
	netlifyAddRewrite("/articles/", "/documents.html")
	netlifyAddRewrite("/articles/index.html", "/documents.html")
	//netlifyAddRewrite("/book/", "/static/documents.html")
	//netflifyAddTempRedirect("/book/*", "/article/:splat")
	netflifyAddTempRedirect("/software/sumatrapdf*", "https://www.sumatrapdfreader.org/:splat")

	netlifyExecTemplate("/documents.html", tmplDocuments, nil)
	netflifyAddTempRedirect("/static/documents.html", "/documents.html")

	{
		// url: /book/go-cookbook.html
		model := struct {
			InProduction bool
		}{
			InProduction: true,
		}
		netlifyExecTemplate("/book/go-cookbook.html", tmplGoCookBook, model)
		netlifyAddRewrite("/articles/go-cookbook.html", "/book/go-cookbook.html")
	}

	{
		// url: /book/windows-programming-in-go.html
		model := struct {
			InProduction bool
		}{
			InProduction: true,
		}
		netlifyExecTemplate("/book/go-cookbook.html", tmplGoCookBook, model)
		netlifyAddRewrite("/articles/go-cookbook.html", "/book/go-cookbook.html")
	}

	{
		// /
		articles := store.GetArticles(articlesNormal)
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
		netlifyExecTemplate("/index.html", tmplMainPage, model)
	}

	{
		// /atom.xml
		d, err := genAtomXML(true)
		panicIfErr(err)
		netlifyWriteFile("/atom.xml", d)
	}

	{
		// /atom-all.xml
		d, err := genAtomXML(false)
		panicIfErr(err)
		netlifyWriteFile("/atom-all.xml", d)
	}

	{
		// /blog/ and /kb/ are only for redirects, we only handle /article/ at this point
		articles := store.GetArticles(articlesWithHidden)
		logVerbose("%d articles\n", len(articles))
		for _, a := range articles {
			article := getArticleByID(a.ID)
			panicIf(article == nil, "No article for id '%s'", a.ID)
			shareHTML := makeShareHTML(article)

			coverImage := ""
			if article.HeaderImageURL != "" {
				coverImage = netlifyRequestGetFullHost() + article.HeaderImageURL
			}

			canonicalURL := netlifyRequestGetFullHost() + article.URL()
			model := struct {
				AnalyticsCode  string
				Article        *Article
				CanonicalURL   string
				CoverImage     string
				PageTitle      string
				ShareHTML      template.HTML
				TagsDisplay    string
				HeaderImageURL string
				GitHubEditURL  string
				NotionEditURL  string
			}{
				AnalyticsCode: analyticsCode,
				Article:       article,
				CanonicalURL:  canonicalURL,
				CoverImage:    coverImage,
				PageTitle:     article.Title,
				ShareHTML:     template.HTML(shareHTML),
			}
			if a.pageInfo != nil {
				id := normalizeID(a.pageInfo.ID)
				model.NotionEditURL = "https://notion.so/" + id
			} else {
				model.GitHubEditURL = "https://github.com/kjk/blog/edit/master/" + article.OrigPath
			}

			path := fmt.Sprintf("/blog/%s.html", article.ID)
			logVerbose("%s, %s => %s, %s, %s\n", article.OrigID, article.ID, path, article.URL(), article.Title)
			netlifyExecTemplate(path, tmplArticle, model)
			netlifyAddRewrite(article.URL(), path)
		}
	}

	{
		// /archives.html
		netlifyWriteArticlesArchiveForTag("")
		seenTags := make(map[string]bool)
		articles := store.GetArticles(articlesWithLessVisible)
		for _, article := range articles {
			for _, tag := range article.Tags {
				if !seenTags[tag] {
					netlifyWriteArticlesArchiveForTag(tag)
					seenTags[tag] = true
					continue
				}
			}
		}
	}

	{
		// /dailynotes (index page)
		weekStart := notesWeekStarts[0]
		notes := notesWeekStartDayToNotes[weekStart]
		var nextWeek string
		if len(notesWeekStarts) > 1 {
			nextWeek = notesWeekStarts[1]
		}
		model := &modelNotesForWeek{
			Notes:         notes,
			TagCounts:     notesTagCounts,
			TotalNotes:    nTotalNotes,
			WeekStartDay:  weekStart,
			AnalyticsCode: analyticsCode,
			NextWeek:      nextWeek,
		}
		path := "/dailynotes/dailynotes.html"
		netlifyExecTemplate(path, tmplNotesWeek, model)
		netlifyAddRewrite("/dailynotes", path)
	}

	{
		// /dailynotes/week/${day} : week starting with a given day
		for weekStart, notes := range notesWeekStartDayToNotes {
			panicIf(len(notes) == 0, "no notes for week '%s'", weekStart)
			var nextWeek, prevWeek string
			for idx, ws := range notesWeekStarts {
				if ws != weekStart {
					continue
				}
				if idx > 0 {
					prevWeek = notesWeekStarts[idx-1]
				}
				lastIdx := len(notesWeekStarts) - 1
				if idx+1 <= lastIdx {
					nextWeek = notesWeekStarts[idx+1]
				}
				break
			}
			model := &modelNotesForWeek{
				Notes:         notes,
				TagCounts:     notesTagCounts,
				WeekStartDay:  weekStart,
				NextWeek:      nextWeek,
				PrevWeek:      prevWeek,
				AnalyticsCode: analyticsCode,
			}
			path := fmt.Sprintf("/dailynotes/dailynotes-week-%s.html", weekStart)
			netlifyExecTemplate(path, tmplNotesWeek, model)
			from := "/dailynotes/week/" + weekStart
			netlifyAddRewrite(from, path)
		}
	}

	{
		// /dailynotes/note/${id}-${title}
		for _, aNote := range notesAllNotes {
			weekStartTime := calcWeekStart(aNote.Day)
			weekStartDay := weekStartTime.Format("2006-01-02")
			model := struct {
				WeekStartDay  string
				Note          *note
				AnalyticsCode string
			}{
				WeekStartDay:  weekStartDay,
				Note:          aNote,
				AnalyticsCode: analyticsCode,
			}
			path := fmt.Sprintf("/dailynotes/dailynotes-note-%s.html", aNote.ID)
			netlifyExecTemplate(path, tmplNotesNote, model)
			from := fmt.Sprintf("/dailynotes/note/%s", aNote.ID)
			if aNote.Title != "" {
				from += "-" + urlify(aNote.Title)
			}
			netlifyAddRewrite(from, path)
		}
	}

	{
		// /dailynotes/tag/${tag}
		seenTags := make(map[string]bool)
		for tag, notes := range notesTagToNotes {
			panicIf(len(notes) == 0, "no notes for tag '%s'", tag)
			model := struct {
				Notes         []*note
				TagCounts     []tagWithCount
				Tag           string
				AnalyticsCode string
			}{
				Notes:         notes,
				TagCounts:     notesTagCounts,
				Tag:           tag,
				AnalyticsCode: analyticsCode,
			}
			// TODO: this tag can be
			tag2 := urlify(tag)
			panicIf(seenTags[tag2], "already seen tag: '%s' '%s'", tag, tag2)
			path := fmt.Sprintf("/dailynotes/dailynotes-tag-%s.html", tag2)
			netlifyExecTemplate(path, tmplNotesTag, model)
			from := fmt.Sprintf("/dailynotes/tag/%s", tag)
			netlifyAddRewrite(from, path)
		}
	}

	{
		// /dailynotes-atom.xml
		notes := notesAllNotes
		if len(notes) > 25 {
			notes = notes[:25]
		}

		pubTime := time.Now()
		if len(notes) > 0 {
			pubTime = notes[0].Day
		}

		feed := &atom.Feed{
			Title:   "Krzysztof Kowalczyk daily notes",
			Link:    "https://blog.kowalczyk.info/dailynotes-atom.xml",
			PubDate: pubTime,
		}

		for _, n := range notes {
			//id := fmt.Sprintf("tag:blog.kowalczyk.info,1999:%d", a.Id)
			title := n.Title
			if title == "" {
				title = n.ID
			}
			e := &atom.Entry{
				Title:   title,
				Link:    "https://blog.kowalczyk.info/" + n.URL,
				Content: string(n.HTMLBody),
				PubDate: n.Day,
			}
			feed.AddEntry(e)
		}

		data, err := feed.GenXml()
		panicIfErr(err)
		netlifyWriteFile("/dailynotes-atom.xml", data)
	}

	{
		// /sitemap.xml
		data, err := genSiteMap("https://blog.kowalczyk.info")
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

	netlifyAddArticleRedirects()
	netlifyWriteRedirects()
	writeCaddyConfig()
}
