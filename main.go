package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/url"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	store         *ArticlesStore
	analyticsCode = "UA-194516-1"

	flgRedownloadNotion bool
	flgDeploy           bool
	flgPreview          bool
	flgVerbose          bool
	inProduction        bool
)

func parseCmdLineFlags() {
	flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
	flag.BoolVar(&flgDeploy, "deploy", false, "if true, build for deployment")
	flag.BoolVar(&flgPreview, "preview", false, "if true, runs caddy and opens a browser for preview")
	flag.BoolVar(&flgRedownloadNotion, "redownload-notion", false, "if true, re-downloads content from notion")
	flag.Parse()
}

func logVerbose(format string, args ...interface{}) {
	if !flgVerbose {
		return
	}
	fmt.Printf(format, args...)
}

func rebuildAll() {
	regenMd()
	loadTemplates()

	s, err := NewArticlesStore()
	panicIfErr(err)
	store = s

	readRedirects()
	netlifyBuild()
}

// caddy -log stdout
func runCaddy() {
	cmd := exec.Command("caddy", "-log", "stdout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func stopCaddy(cmd *exec.Cmd) {
	cmd.Process.Kill()
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

func preview() {
	go func() {
		time.Sleep(time.Second * 3)
		openBrowser("http://localhost:8080")
	}()
	runCaddy()
}

func main() {
	parseCmdLineFlags()
	os.MkdirAll("netlify_static", 0755)

	// make sure this happens first so that building for deployment is not
	// disrupted by the temporary testing code we might have below
	if flgDeploy {
		inProduction = true
		rebuildAll()
		return
	}

	if false {
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
	if flgPreview {
		preview()
	}
}
