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
	analyticsCode = "UA-194516-1"

	flgRedownloadNotion bool
	flgDeploy           bool
	flgPreview          bool
	flgVerbose          bool
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
	loadAllArticles()
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
		time.Sleep(time.Second * 1)
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
		rebuildAll()
		return
	}

	if false {
		notionRedownloadOne("88aee8f43620471aa9dbcad28368174c")
		os.Exit(0)
	}

	if false {
		testNotionToHTMLOnePage("dd5c0a813dfe4487a6cd432f82c0c2fc")
		os.Exit(0)
	}

	if flgRedownloadNotion {
		notionRedownloadAll()
		return
	}

	rebuildAll()
	if flgPreview {
		preview()
	}
}
