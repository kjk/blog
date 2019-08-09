package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi"
)

var (
	cacheDir = "notion_cache"
)

func sha1OfLink(link string) string {
	link = strings.ToLower(link)
	h := sha1.New()
	h.Write([]byte(link))
	return fmt.Sprintf("%x", h.Sum(nil))
}

var imgFiles []os.FileInfo

func findImageInDir(imgDir string, sha1 string) string {
	if len(imgFiles) == 0 {
		imgFiles, _ = ioutil.ReadDir(imgDir)
	}
	for _, fi := range imgFiles {
		if strings.HasPrefix(fi.Name(), sha1) {
			return filepath.Join(imgDir, fi.Name())
		}
	}
	return ""
}

func guessExt(fileName string, contentType string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".png", ".jpg", ".jpeg":
		return ext
	}
	switch contentType {
	case "image/png":
		return ".png"
	}
	panic(fmt.Errorf("Didn't find ext for file '%s', content type '%s'", fileName, contentType))
}

func downloadImage(c *notionapi.Client, uri string) ([]byte, string, error) {
	img, err := c.DownloadFile(uri)
	if err != nil {
		lg("\n  failed with %s\n", err)
		return nil, "", err
	}
	ext := guessExt(uri, img.Header.Get("Content-Type"))
	return img.Data, ext, nil
}

// return path of cached image on disk
func downloadAndCacheImage(c *notionapi.Client, uri string) (string, error) {
	sha := sha1OfLink(uri)

	//ext := strings.ToLower(filepath.Ext(uri))

	imgDir := filepath.Join(cacheDir, "img")
	err := os.MkdirAll(imgDir, 0755)
	panicIfErr(err)

	cachedPath := findImageInDir(imgDir, sha)
	if cachedPath != "" {
		verbose("Image %s already downloaded as %s\n", uri, cachedPath)
		return cachedPath, nil
	}

	timeStart := time.Now()
	lg("Downloading %s ... ", uri)

	imgData, ext, err := downloadImage(c, uri)

	cachedPath = filepath.Join(imgDir, sha+ext)

	err = ioutil.WriteFile(cachedPath, imgData, 0644)
	if err != nil {
		return "", err
	}
	lg("finished in %s. Wrote as '%s'\n", time.Since(timeStart), cachedPath)

	return cachedPath, nil
}

func rmFile(path string) {
	err := os.Remove(path)
	if err != nil {
		lg("os.Remove(%s) failed with %s\n", path, err)
	}
}

func rmCached(pageID string) {
	id := normalizeID(pageID)
	rmFile(filepath.Join(cacheDir, id+".txt"))
}

func loadPageAsArticle(d *CachingDownloader, pageID string) *Article {
	page, err := d.downloadAndCachePage(pageID)
	panicIfErr(err)
	lg("Downloaded %s %s\n", pageID, page.Root().Title)
	c := d.Client
	return notionPageToArticle(c, page)
}
