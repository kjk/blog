package main

import (
	"bytes"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var redirects = map[string]string{
	"/index.html":                  "/",
	"/blog":                        "/",
	"/blog/":                       "/",
	"/kb/serialization-in-c#.html": "/article/Serialization-in-C.html",
	"/extremeoptimizations":        "/extremeoptimizations/index.html",
	"/extremeoptimizations/":       "/extremeoptimizations/index.html",
	"/feed/rss2/atom.xml":          "/atom.xml",
	"/feed/rss2/":                  "/atom.xml",
	"/feed/rss2":                   "/atom.xml",
	"/feed/":                       "/atom.xml",
	"/feed":                        "/atom.xml",
	"/articles/cocoa-objectivec-reference.html":     "/articles/cocoa-reference.html",
	"/forum_sumatra/rss.php":                        "http://forums.fofou.org/sumatrapdf/rss",
	"/forum_sumatra":                                "http://forums.fofou.org/sumatrapdf",
	"/google6dba371684d43cd6.html":                  "/static/google6dba371684d43cd6.html",
	"/software/15minutes/index.html":                "/software/15minutes.html",
	"/software/15minutes/":                          "/software/15minutes.html",
	"/software/fofou":                               "/software/fofou/index.html",
	"/software/patheditor":                          "/software/patheditor/for-windows.html",
	"/software/patheditor/":                         "/software/patheditor/for-windows.html",
	"/software/scdiff/":                             "/software/scdiff.html",
	"/software/scdiff/index.html":                   "/software/scdiff.html",
	"/software/sumatra":                             "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf":                          "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/":                         "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/index.html":               "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/download.html _blank":     "/software/sumatrapdf/download-free-pdf-viewer.html",
	"/software/sumatrapdf/download.html":            "/software/sumatrapdf/download-free-pdf-viewer.html",
	"/software/sumatrapdf/prerelase.html":           "/software/sumatrapdf/prerelease.html",
	"/software/sumatrapdf/sumatra-shot-00.gif":      "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-00.gif",
	"/software/sumatrapdf/sumatra-shot-01.gif":      "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-01.gif",
	"/software/sumatrapdf/sumatra-shot-00-full.gif": "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-00-full.gif",
	"/software/sumatrapdf/sumatra-shot-01-full.gif": "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-01-full.gif",
	"/software/sumatrapdf/SumatraSplash.png":        "http://kjkpub.s3.amazonaws.com/blog/sumatra/SumatraSplash.png",
	"/software/volante":                             "/software/volante/database.html",
	"/software/volante/":                            "/software/volante/database.html",
	"/software/volante/index.html":                  "/software/volante/database.html",
	"/static/software.html":                         "/software/index.html",
}

var articleRedirects = make(map[string]int)
var articleRedirectsMutex sync.Mutex

func readRedirects() {
	fname := filepath.Join(getDataDir(), "data", "article_redirects.txt")
	d, err := ReadFileAll(fname)
	if err != nil {
		return
	}
	lines := bytes.Split(d, []byte{'\n'})
	for _, l := range lines {
		if 0 == len(l) {
			continue
		}
		parts := strings.Split(string(l), "|")
		if len(parts) != 2 {
			panic("malformed article_redirects.txt")
		}
		idStr := parts[0]
		url := strings.TrimSpace(parts[1])
		if id, err := strconv.Atoi(idStr); err != nil {
			panic("malformed line in article_redirects.txt")
		} else {
			a := store.GetArticleById(id)
			if a == nil {
				panic("bad article id article_redirects.txt")
			}
			articleRedirects[url] = id
		}
	}
	logger.Noticef("loaded %d article redirects\n", len(articleRedirects))
}

// return -1 if there's no redirect for this urls
func getRedirectArticleId(url string) int {
	url = url[1:] // remove '/' from the beginning
	articleRedirectsMutex.Lock()
	defer articleRedirectsMutex.Unlock()
	if articleId, ok := articleRedirects[url]; ok {
		return articleId
	}
	return -1
}

func redirectIfNeeded(w http.ResponseWriter, r *http.Request) bool {
	url := r.URL.Path
	//logger.Noticef("redirectIfNeeded(): '%s'", url)
	if redirUrl, ok := redirects[url]; ok {
		//logger.Noticef("Redirecting '%s' => '%s'", url, redirUrl)
		http.Redirect(w, r, redirUrl, 302)
		return true
	}

	redirectArticleId := getRedirectArticleId(url)
	if redirectArticleId == -1 {
		return false
	}
	article := store.GetArticleById(redirectArticleId)
	if article != nil {
		redirUrl := "/" + article.Permalink()
		//logger.Noticef("Redirecting '%s' => '%s'", url, redirUrl)
		http.Redirect(w, r, redirUrl, 302)
		return true
	}

	return false
}

// url: /forum_sumatra/${rest}
func forumRedirect(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/forum_sumatra/"):]
	redirUrl := "http://forums.fofou.org/sumatrapdf/" + url + "?" + r.URL.RawQuery
	//logger.Noticef("Redirecting '%s' => '%s'", r.URL.Path, redirUrl)
	http.Redirect(w, r, redirUrl, 302)
}
