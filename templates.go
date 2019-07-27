package main

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/kjk/u"
)

var (
	tmplMainPage         = "mainpage.tmpl.html"
	tmplArticle          = "article.tmpl.html"
	tmplArchive          = "archive.tmpl.html"
	tmplGenerateUniqueID = "generate-unique-id.tmpl.html"
	tmplGoCookBook       = "go-cookbook.tmpl.html"
	tmplChangelog        = "changelog.tmpl.html"
	tmpl404              = "404.tmpl.html"
	templateNames        = []string{
		tmplMainPage,
		tmplArticle,
		tmplArchive,
		tmplGenerateUniqueID,
		tmplGoCookBook,
		tmplChangelog,
		tmpl404,
		"analytics.tmpl.html",
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

func execTemplateToFile(path string, templateName string, model interface{}) error {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, templateName, model)
	panicIfErr(err)
	err = ioutil.WriteFile(path, buf.Bytes(), 0644)
	return err
}

func execTemplateToWriter(name string, data interface{}, w io.Writer) error {
	loadTemplates()
	return templates.ExecuteTemplate(w, name, data)
}

func execTemplate(path string, tmplName string, d interface{}, w io.Writer) error {
	// this code path is for the preview on demand server
	if w != nil {
		return execTemplateToWriter(tmplName, d, w)
	}

	// this code path is for generating static files
	netPath := netlifyPath(path)
	err := execTemplateToFile(netPath, tmplName, d)
	panicIfErr(err)
	return nil
}
