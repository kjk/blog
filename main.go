package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/url"
	"os"
	"os/exec"
	"runtime"
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
}

func main() {
	parseCmdLineFlags()
	os.MkdirAll("netlify_static", 0755)

	if true {
		testOneNotionPage()
		os.Exit(0)
	}

	if false {
		testNotionToHTML()
		os.Exit(0)
	}

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
