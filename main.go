package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	_ "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kjk/u"
)

var (
	store         *ArticlesStore
	analyticsCode = "UA-194516-1"
	showDrafts    bool

	flgNewArticleTitle string
	flgWatch           bool
	flgVerbose         bool
	inProduction       bool
)

func parseCmdLineFlags() {
	flag.StringVar(&flgNewArticleTitle, "newarticle", "", "create a new article")
	flag.BoolVar(&flgWatch, "watch", false, "if true, runs caddy for preview and re-builds on changes")
	flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
	flag.Parse()
}

func logVerbose(format string, args ...interface{}) {
	if !flgVerbose {
		return
	}
	fmt.Printf(format, args...)
}

func findUniqueArticleID(articles []*Article) string {
	existingIDs := make(map[string]bool)
	for _, a := range articles {
		existingIDs[a.ID] = true
	}

	for i := 1; i < 10000; i++ {
		s := u.EncodeBase64(i)
		if !existingIDs[s] {
			return strconv.Itoa(i)
		}
	}
	u.PanicIf(true, "couldn't find unique article id")
	return ""
}

func genNewArticle(title string) {
	fmt.Printf("genNewArticle: %q\n", title)
	store, err := NewArticlesStore()
	if err != nil {
		log.Fatalf("NewStore() failed with %s", err)
	}
	newID := findUniqueArticleID(store.articles)
	t := time.Now()
	dir := "articles"
	yyyy := fmt.Sprintf("%04d", t.Year())
	month := t.Month()
	sanitizedTitle := sanitizeForFile(title)
	name := fmt.Sprintf("%02d-%s.md", month, sanitizedTitle)
	fmt.Printf("new id: %s, name: %s\n", newID, name)
	path := filepath.Join(dir, yyyy, name)
	s := fmt.Sprintf(`---
Id: %s
Title: %s
Date: %s
Format: Markdown
---`, newID, title, t.Format(time.RFC3339))
	for i := 1; i < 10; i++ {
		if !u.PathExists(path) {
			break
		}
		name = fmt.Sprintf("%02d-%s-%d.md", month, sanitizedTitle, i)
		path = filepath.Join(dir, yyyy, name)
	}
	u.PanicIf(u.PathExists(path))
	fmt.Printf("path: %s\n", path)
	u.CreateDirForFileMust(path)
	ioutil.WriteFile(path, []byte(s), 0644)
}

func loadArticles() {
	s, err := NewArticlesStore()
	u.PanicIfErr(err)
	store = s
}

func rebuildAll() {
	notesGenIDIfNecessary()
	regenMd()
	loadTemplates()
	loadArticles()
	readRedirects()
	netlifyBuild()
}

// caddy -log stdout
func runCaddy() *exec.Cmd {
	cmd := exec.Command("caddy", "-log", "stdout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	u.PanicIfErr(err)
	return cmd
}

func getDirsRecur(dir string) ([]string, error) {
	toVisit := []string{dir}
	idx := 0
	for idx < len(toVisit) {
		dir = toVisit[idx]
		idx++
		fileInfos, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, fi := range fileInfos {
			if !fi.IsDir() {
				continue
			}
			path := filepath.Join(dir, fi.Name())
			toVisit = append(toVisit, path)
		}
	}
	return toVisit, nil
}

func rebuildOnChanges() {
	dirs, err := getDirsRecur("www")
	u.PanicIfErr(err)
	dirs2, err := getDirsRecur("books")
	u.PanicIfErr(err)
	dirs3, err := getDirsRecur("articles")
	u.PanicIfErr(err)
	dirs = append(dirs, dirs2...)
	dirs = append(dirs, dirs3...)

	watcher, err := fsnotify.NewWatcher()
	u.PanicIfErr(err)
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered in rebuildOnChanges(). Error: '%s'\n", r)
				// TODO: why this doesn't seem to trigger done
				done <- true
			}
		}()

		for {
			select {
			case event := <-watcher.Events:
				// filter out events that are just chmods
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				fmt.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("modified file:", event.Name)
				}
				if isWhitelistedFromChanges(event.Name) {
					fmt.Printf("no rebuild because file whitelisted\n")
				} else {
					// TODO: could also restart caddy to pick up redirects
					rebuildAll()
				}
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()
	for _, dir := range dirs {
		fmt.Printf("Watching dir: '%s'\n", dir)
		watcher.Add(dir)
	}
	// waiting forever
	// TODO: pick up ctrl-c and cleanup and quit
	<-done
	fmt.Printf("exiting rebuildOnChanges()")
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func runCaddyAndWatch() {
	runCaddy()
	openBrowser("http://localhost:8080")
	rebuildOnChanges()
}

func main() {
	if true {
		importNotion()
		os.Exit(0)
	}

	rand.Seed(time.Now().UnixNano())
	parseCmdLineFlags()
	os.MkdirAll("netlify_static", 0755)

	if flgNewArticleTitle != "" {
		genNewArticle(flgNewArticleTitle)
		return
	}
	inProduction = true
	if flgWatch {
		showDrafts = true
		inProduction = false
	}
	rebuildAll()
	if flgWatch {
		runCaddyAndWatch()
	}
}
