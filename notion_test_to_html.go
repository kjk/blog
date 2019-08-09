package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	destDir = "netlify_static"
)

func copyCSS() {
	src := filepath.Join("www", "css", "main.css")
	dst := filepath.Join(destDir, "main.css")
	err := copyFile(dst, src)
	panicIfErr(err)
}

func createDestDir() {
	err := os.MkdirAll(destDir, 0755)
	panicIfErr(err)
}

func createNotionDirs() {
	err := os.MkdirAll(cacheDir, 0755)
	panicIfErr(err)
}

// downloads and html
func testNotionToHTMLOnePage(d *CachingDownloader, id string) {

	//id := "c9bef0f1c8fe40a2bc8b06ace2bd7d8f" // tools page, columns
	//id := "0a66e6c0c36f4de49417a47e2c40a87e" // mono-spaced page with toggle, devlog 2018
	//id := "484919a1647144c29234447ce408ff6b" // test toggle
	//id := "88aee8f43620471aa9dbcad28368174c" // test image and gist
	loadTemplates()
	createNotionDirs()
	createDestDir()

	id = normalizeID(id)
	article := loadPageAsArticle(d, id)

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

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, tmplArticle, model)
	panicIfErr(err)
	data := buf.Bytes()
	data = bytes.Replace(data, []byte("/css/main.css"), []byte("/main.css"), -1)

	path := filepath.Join(destDir, "index.html")
	err = ioutil.WriteFile(path, data, 0644)
	panicIfErr(err)
	copyCSS()

	err = os.Chdir(destDir)
	panicIfErr(err)

	go func() {
		time.Sleep(time.Second * 1)
		openBrowser("http://localhost:2015")
	}()
	runCaddy()
}
