package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var (
	gPreviewArticles *Articles
)

func serve404(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	path := filepath.Join("www", "404.html")

	parts := strings.Split(uri[1:], "/")
	if len(parts) > 2 && parts[0] == "essential" {
		bookName := parts[1]
		maybePath := filepath.Join("www", "essential", bookName, "404.html")
		if fileExists(maybePath) {
			fmt.Printf("'%s' exists\n", maybePath)
			path = maybePath
		} else {
			fmt.Printf("'%s' doesn't exist\n", maybePath)
		}
	}
	fmt.Printf("Serving 404 from '%s' for '%s'\n", path, uri)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		d = []byte(fmt.Sprintf("URL '%s' not found!", uri))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	w.Write(d)
}

func writeHTMLHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
}

// tempRedirect gives a Moved Permanently response.
func doRedirect(w http.ResponseWriter, r *http.Request, newPath string, code int) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	if code == 0 {
		code = http.StatusFound
	}
	w.WriteHeader(code)
}

func handleIndexOnDemand(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("uri: %s\n", r.URL.Path)
	uri := r.URL.Path
	nAgain := 0
again:
	nAgain++
	if nAgain > 3 {
		fmt.Printf("Too many 200 redirects for '%s'\n", r.URL.Path)
		serve404(w, r)
		return
	}
	if tryServeKnown(w, r) {
		return
	}

	if tryServeNotionCacheImg(w, r) {
		return
	}

	if tryServeTagArchive(w, r) {
		return
	}

	if tryServeArticle(w, r) {
		return
	}

	for _, redir := range netlifyRedirects {
		if redir.from == uri {
			// this is internal rewrite
			if redir.code == 200 {
				fmt.Printf("Rewrite %s => %s\n", uri, redir.to)
				uri = redir.to
				r.URL.Path = uri
				goto again
			}
			fmt.Printf("Redirect %s => %s, code: %d\n", uri, redir.to, redir.code)
			doRedirect(w, r, redir.to, redir.code)
		}
	}

	isDir := strings.HasSuffix(uri, "/")
	fileName := filepath.FromSlash(uri[1:])
	if isDir {
		fileName = filepath.Join(fileName, "index.html")
	}
	path := filepath.Join("www", fileName)
	if !fileExists(path) {
		fmt.Printf("path '%s' for url '%s' doesn't exist\n", path, uri)
		serve404(w, r)
		return
	}
	fmt.Printf("Serving file: %s\n", path)
	http.ServeFile(w, r, path)
}

var knownURLs = map[string]func(*Articles, io.Writer) error{
	"/":               genIndex,
	"/changelog.html": genChangelog,
	"/blogindex.html": genBlogIndex,
	"/archives.html":  genArchives,
	"/sitemap.xml":    genSitemap,
	"/atom.xml":       genAtom,
	"/atom-all.xml":   genAtomAll,
}

func tryServeKnown(w http.ResponseWriter, r *http.Request) bool {
	uri := r.URL.Path
	if fn := knownURLs[uri]; fn != nil {
		writeHTMLHeaders(w)
		logIfError(fn(gPreviewArticles, w))
		return true
	}
	return false
}

func findArticleByID(store *Articles, id string) *Article {
	for _, article := range store.articles {
		if article.ID == id {
			return article
		}
	}
	return nil
}

func tryServeArticle(w http.ResponseWriter, r *http.Request) bool {
	uri := r.URL.Path
	articlePath := strings.TrimPrefix(uri, "/article/")
	if uri == articlePath {
		return false
	}
	articlePath = strings.Split(articlePath, "/")[0]
	articleID := strings.TrimSuffix(articlePath, ".html")
	article := findArticleByID(gPreviewArticles, articleID)
	if article == nil {
		fmt.Printf("Didn't find article with id '%s' for uri '%s'\n", articleID, uri)
		return false
	}
	genArticle(article, w)
	return true
}

// tag/<tagname>
func tryServeTagArchive(w http.ResponseWriter, r *http.Request) bool {
	uri := r.URL.Path
	tag := strings.TrimPrefix(uri, "/tag/")
	if uri == tag {
		return false
	}
	netlifyWriteArticlesArchiveForTag(gPreviewArticles, tag, w)
	return true
}

// for /img/<name> check notion_cache/img
func tryServeNotionCacheImg(w http.ResponseWriter, r *http.Request) bool {
	uri := r.URL.Path
	imgName := strings.TrimPrefix(uri, "/img/")
	if uri == imgName {
		return false
	}
	path := filepath.Join("notion_cache", "img", imgName)
	if !fileExists(path) {
		return false
	}
	http.ServeFile(w, r, path)
	return true
}

// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
func makeHTTPServerOnDemand() *http.Server {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", handleIndexOnDemand)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      mux,
	}
	return srv
}

func startPreviewOnDemand(articles *Articles) {
	gPreviewArticles = articles

	addAllRedirects(gPreviewArticles)

	httpSrv := makeHTTPServerOnDemand()
	httpSrv.Addr = "127.0.0.1:8183"

	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		panicIfErr(err)
		fmt.Printf("HTTP server shutdown gracefully\n")
	}()
	fmt.Printf("Started listening on %s\n", httpSrv.Addr)
	openBrowser("http://" + httpSrv.Addr)

	waitForCtrlC()
}
