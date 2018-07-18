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
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	store         *ArticlesStore
	analyticsCode = "UA-194516-1"
	showDrafts    bool

	flgRedownloadNotion bool
	flgVerbose          bool
	inProduction        bool
)

func parseCmdLineFlags() {
	flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
	flag.BoolVar(&flgRedownloadNotion, "redownload-notion", false, "if true, re-downloads content from notion")
	flag.Parse()
}

func logVerbose(format string, args ...interface{}) {
	if !flgVerbose {
		return
	}
	fmt.Printf(format, args...)
}

func loadArticlesAndNotes() {
	s, err := NewArticlesStore()
	panicIfErr(err)
	store = s
}

func rebuildAll() {
	regenMd()
	loadTemplates()
	loadArticlesAndNotes()
	readRedirects()
	netlifyBuild()
}

// caddy -log stdout
func runCaddy() *exec.Cmd {
	cmd := exec.Command("caddy", "-log", "stdout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	panicIfErr(err)
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
	panicIfErr(err)
	dirs2, err := getDirsRecur("books")
	panicIfErr(err)
	dirs3, err := getDirsRecur("articles")
	panicIfErr(err)
	dirs = append(dirs, dirs2...)
	dirs = append(dirs, dirs3...)

	watcher, err := fsnotify.NewWatcher()
	panicIfErr(err)
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
	rand.Seed(time.Now().UnixNano())
	parseCmdLineFlags()
	os.MkdirAll("netlify_static", 0755)

	if false {
		_, err := loadPageAsArticle("fa3fc358e5644f39b89c57f13d426d54")
		if err != nil {
			fmt.Printf("loadPageAsArticle() failed with '%s'\n", err)
		}
		os.Exit(0)
	}

	if flgRedownloadNotion {
		notionRedownload()
		return
	}

	inProduction = true
	rebuildAll()
}
