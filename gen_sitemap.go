package main

import (
	"encoding/xml"
	"path"
	"time"
)

// SiteMapURLSet represents <urlset>
type SiteMapURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Ns      string   `xml:"xmlns,attr"`
	URLS    []SiteMapURL
}

func makeSiteMapURLSet() *SiteMapURLSet {
	return &SiteMapURLSet{
		Ns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}
}

// SiteMapURL represents a single url
type SiteMapURL struct {
	XMLName      xml.Name `xml:"url"`
	URL          string   `xml:"loc"`
	LastModified string   `xml:"lastmod"`
}

// There are more static pages, but those are the important ones
var staticURLS = []string{
	"/book/go-cookbook.html",
	"/articles/cbz-cbr-comic-book-reader-viewer-for-windows.html",
	"/articles/chm-reader-viewer-for-windows.html",
	"/articles/mobi-ebook-reader-viewer-for-windows.html",
	"/articles/epub-ebook-reader-viewer-for-windows.html",
	"/articles/where-to-get-free-ebooks-epub-mobi.html",
	"/software/",
	"/documents.html",
}

func genSiteMap(store *Articles, host string) ([]byte, error) {
	articles := store.getNotHidden()
	urlset := makeSiteMapURLSet()
	var urls []SiteMapURL
	for _, article := range articles {
		pageURL := path.Join(host, article.URL())
		uri := SiteMapURL{
			URL:          pageURL,
			LastModified: article.UpdatedOn.Format("2006-01-02"),
		}
		urls = append(urls, uri)
	}

	now := time.Now()
	for _, staticURL := range staticURLS {
		pageURL := path.Join(host, staticURL)
		uri := SiteMapURL{
			URL:          pageURL,
			LastModified: now.Format("2006-01-02"),
		}
		urls = append(urls, uri)
	}

	urlset.URLS = urls

	xmlData, err := xml.MarshalIndent(urlset, "", "")
	if err != nil {
		return nil, err
	}
	d := append([]byte(xml.Header), xmlData...)
	return d, nil
}
