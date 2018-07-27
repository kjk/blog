package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/kjk/u"
)

var (
	tmplMainPage         = "mainpage.tmpl.html"
	tmplBlogIndex        = "blog_index.tmpl.html"
	tmplArticle          = "article.tmpl.html"
	tmplArchive          = "archive.tmpl.html"
	tmplGenerateUniqueID = "generate-unique-id.tmpl.html"
	tmplGoCookBook       = "go-cookbook.tmpl.html"
	tmplChangelog        = "changelog.tmpl.html"
	tmpl404              = "404.tmpl.html"
	templateNames        = []string{
		tmplMainPage,
		tmplBlogIndex,
		tmplArticle,
		tmplArchive,
		tmplGenerateUniqueID,
		tmplGoCookBook,
		tmplChangelog,
		tmpl404,
		"analytics.tmpl.html",
		"page_navbar.tmpl.html",
	}
	templatePaths []string
	templates     *template.Template

	// dirs to search when looking for templates
	tmplDirs = []string{
		"www",
		filepath.Join("www", "tmpl"),
		filepath.Join("www", "tools"),
		filepath.Join("www", "static"),
	}
)

func findTemplate(name string) string {
	for _, dir := range tmplDirs {
		path := filepath.Join(dir, name)
		if u.FileExists(path) {
			return path
		}
	}
	panicIf(true, "didn't find tamplate %s in dirs %v", name, tmplDirs)
	return ""
}

func loadTemplates() {
	for _, name := range templateNames {
		path := findTemplate(name)
		templatePaths = append(templatePaths, path)
	}
	templates = template.Must(template.ParseFiles(templatePaths...))
}

func netlifyExecTemplate(fileName string, templateName string, model interface{}) {
	path := netlifyPath(fileName)
	execTemplate(path, templateName, model)
}

func execTemplate(path string, templateName string, model interface{}) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, templateName, model)
	panicIfErr(err)
	err = ioutil.WriteFile(path, buf.Bytes(), 0644)
	panicIfErr(err)
}
