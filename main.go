package main

import (
	"flag"
	"fmt"
	_ "net/url"
	"os"
	"os/exec"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
	"github.com/kjk/u"
)

var (
	panicIfErr = u.PanicIfErr
	fatalIf    = u.PanicIf
)

const (
	analyticsCode = "UA-194516-1"
	htmlDir       = "netlify_static" // directory where we generate html files
)

var (
	flgVerbose bool
)

func rebuildAll(d *caching_downloader.Downloader) *Articles {
	regenMd()
	loadTemplates()
	articles := loadArticles(d)
	readRedirects(articles)
	generateHTML(articles)
	return articles
}

// caddy -log stdout
func runCaddy() {
	cmd := exec.Command("caddy", "-log", "stdout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runWranglerDev() {
	cmd := exec.Command("wrangler", "dev")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	logIfError(err)
}

/*
func stopCaddy(cmd *exec.Cmd) {
	cmd.Process.Kill()
}
*/

func preview() {
	go func() {
		time.Sleep(time.Second * 1)
		u.OpenBrowser("http://localhost:8787")
	}()
	runWranglerDev()
}

var (
	nDownloadedPage = 0
)

func eventObserver(ev interface{}) {
	switch v := ev.(type) {
	case *caching_downloader.EventError:
		logf(v.Error)
	case *caching_downloader.EventDidDownload:
		nDownloadedPage++
		logf("%03d '%s' : downloaded in %s\n", nDownloadedPage, v.PageID, v.Duration)
	case *caching_downloader.EventDidReadFromCache:
		// TODO: only verbose
		nDownloadedPage++
		logf("%03d '%s' : read from cache in %s\n", nDownloadedPage, v.PageID, v.Duration)
	case *caching_downloader.EventGotVersions:
		logf("downloaded info about %d versions in %s\n", v.Count, v.Duration)
	}
}

func newNotionClient() *notionapi.Client {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		logf("must set NOTION_TOKEN env variable\n")
		//flag.Usage()
		os.Exit(1)
	}
	// TODO: verify token still valid, somehow
	client := &notionapi.Client{
		AuthToken: token,
	}
	if flgVerbose {
		client.Logger = os.Stdout
	}
	return client
}

func recreateDir(dir string) {
	err := os.RemoveAll(dir)
	must(err)
	err = os.MkdirAll(dir, 0755)
	must(err)
}

func main() {
	var (
		flgDeploy          bool
		flgPreview         bool
		flgPreviewOnDemand bool
		flgNoCache         bool
		flgWc              bool
		flgRedownload      bool
		flgRebuild         bool
		flgDiff            bool
	)

	{
		flag.BoolVar(&flgWc, "wc", false, "wc -l i.e. line count")
		flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
		flag.BoolVar(&flgNoCache, "no-cache", false, "if true, disables cache for downloading notion pages")
		flag.BoolVar(&flgDeploy, "deploy", false, "deploy to Cloudflare")
		flag.BoolVar(&flgPreview, "preview", false, "runs caddy and opens a browser for preview")
		flag.BoolVar(&flgPreviewOnDemand, "preview-on-demand", false, "runs the browser for local preview")
		flag.BoolVar(&flgRedownload, "redownload-notion", false, "download the content from Notion")
		flag.BoolVar(&flgRebuild, "rebuild", false, fmt.Sprintf("rebuild site in %s/ directory", htmlDir))
		flag.BoolVar(&flgDiff, "diff", false, "preview diff using winmerge")
		flag.Parse()
	}

	openLog()
	defer closeLog()

	if flgWc {
		doLineCount()
		return
	}

	if flgDiff {
		winmergeDiffPreview()
		return
	}

	hasCmd := flgPreview || flgPreviewOnDemand || flgRedownload || flgRebuild || flgDeploy
	if !hasCmd {
		flag.Usage()
		return
	}

	if flgRebuild {
		cache, err := caching_downloader.NewDirectoryCache(cacheDir)
		must(err)
		d := caching_downloader.New(cache, nil)
		d.EventObserver = eventObserver
		d.RedownloadNewerVersions = false
		d.NoReadCache = false
		rebuildAll(d)
		return
	}

	client := newNotionClient()
	cache, err := caching_downloader.NewDirectoryCache(cacheDir)
	must(err)
	d := caching_downloader.New(cache, client)
	d.EventObserver = eventObserver
	d.RedownloadNewerVersions = true
	d.NoReadCache = flgNoCache

	if flgRedownload {
		rebuildAll(d)
		return
	}

	if flgDeploy {
		rebuildAll(d)
		cmd := exec.Command("wrangler", "publish")
		u.RunCmdLoggedMust(cmd)
		u.OpenBrowser("https://blog.kowalczyk.info")
		return
	}

	if false {
		testNotionToHTMLOnePage(d, "dfbefe6906a943d8b554699341e997b0")
		os.Exit(0)
	}

	articles := rebuildAll(d)

	if flgPreview {
		preview()
		return
	}

	if flgPreviewOnDemand {
		startPreviewOnDemand(articles)
		return
	}
}
