package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

func handleIndexOnDemand(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("uri: %s\n", r.URL.Path)
	uri := r.URL.Path
	if uri == "/" {
		writeHTMLHeaders(w)
		err := genIndex(gPreviewArticles, w)
		logIfError(err)
		return
	}

	serve404(w, r)
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
	httpSrv := makeHTTPServerOnDemand()
	httpSrv.Addr = "127.0.0.1:8173"

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

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("Got signal %s\n", sig)
}
