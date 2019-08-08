package main

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

// Converter renders article as html
type Converter struct {
	article      *Article
	page         *notionapi.Page
	notionClient *notionapi.Client
	idToArticle  func(string) *Article
	galleries    [][]string

	r *tohtml.Converter
}

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// /article/${id}
func (r *Converter) rewriteURL(uri string) string {
	id := notionapi.ExtractNoDashIDFromNotionURL(uri)
	if id == "" {
		return uri
	}
	article := r.idToArticle(id)
	// this might happen when I link to some-one else's public notion pages
	if article == nil {
		return uri
	}
	return article.URL()
}

func (r *Converter) getURLAndTitleForBlock(block *notionapi.Block) (string, string) {
	id := notionapi.ToNoDashID(block.ID)
	article := r.idToArticle(id)
	if article == nil {
		title := block.Title
		lg("No article for id %s %s\n", id, title)
		url := "/article/" + id + "/" + urlify(title)
		return url, title
	}

	return article.URL(), article.Title
}

func genGalleryMainHTML(galleryID int, imageURL string) string {
	s := `
  <div class="img-wrapper-wrapper">
    <div class="img-wrapper">
      <img id="id-gallery-{galleryID}" src="{imageURL}" />
      <a class="for-nav-icon nav-icon-left" href="#" onclick="imgPrev("{galleryID}"); return false;">
        <svg viewBox="0 0 24 24" preserveAspectRatio="xMidYMid meet" focusable="false" class="nav-icon">
          <g>
            <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z" class="style-scope yt-icon">
            </path>
          </g>
        </svg>
      </a>

      <a class="for-nav-icon nav-icon-right" href="#" onclick="imgNext({galleryID}); return false;">
        <svg viewBox="0 0 24 24" preserveAspectRatio="xMidYMid meet" focusable="false" class="nav-icon" style="">
          <g>
            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" class="yt-icon"></path>
          </g>
        </svg>
      </a>
    </div>
  </div>
`
	s = strings.Replace(s, "{galleryID}", strconv.Itoa(galleryID), -1)
	s = strings.Replace(s, "{imageURL}", imageURL, -1)
	return s
}

func genGalleryThumbHTML(galleryID int, n int, im *ImageMapping) string {
	s := `
    <div id="id-thumb-{galleryID}-{imageNo}" class="pa1 ib">
      <a href="#" onclick="changeShot({galleryID}, {imageNo}); return false;">
        <img id="id-thumb-img-{galleryID}-{imageNo}" src="{imageURL}" width="80" height="60" />
      </a>
	</div>
`
	s = strings.Replace(s, "{galleryID}", strconv.Itoa(galleryID), -1)
	ns := strconv.Itoa(n)
	s = strings.Replace(s, "{imageNo}", ns, -1)
	s = strings.Replace(s, "{imageURL}", im.relativeURL, -1)
	return s
}

func (r *Converter) renderGallery(block *notionapi.Block) bool {
	imageURLS := r.article.getGalleryImages(block)
	if len(imageURLS) == 0 {
		return false
	}
	panicIf(len(imageURLS) < 2, "expected gallery to have at least 2 images, got %d", len(imageURLS))
	galleryID := len(r.galleries)
	r.galleries = append(r.galleries, imageURLS)
	var images []*ImageMapping
	for _, link := range imageURLS {
		im := r.article.findImageMappingBySource(link)
		panicIf(im == nil, "didn't find ImageMapping for %s", link)
		images = append(images, im)
	}
	firstImage := images[0]
	s := genGalleryMainHTML(galleryID, firstImage.relativeURL)
	r.r.WriteString(s)

	r.r.WriteString(`<div class="center mt3 mb6">`)
	for i, im := range images {
		s := genGalleryThumbHTML(galleryID, i, im)
		r.r.WriteString(s)
	}
	r.r.WriteString(`</div>`)
	return true
}

// RenderImage renders BlockImage
func (r *Converter) RenderImage(block *notionapi.Block) bool {
	link := block.Source
	im := r.article.findImageMappingBySource(link)
	relURL := im.relativeURL
	imgURL := r.article.getImageBlockURL(block)
	if imgURL != "" {
		attrs := []string{"href", imgURL, "target", "_blank"}
		r.r.WriteElement(block, "a", attrs, "", true)
		{
			attrs2 := []string{"class", "blog-img", "src", relURL}
			r.r.WriteElement(block, "img", attrs2, "", true)
			r.r.WriteElement(block, "img", attrs2, "", false)
		}
		r.r.WriteElement(block, "a", attrs, "", false)
	} else {
		attrs := []string{"class", "blog-img", "src", relURL}
		r.r.WriteElement(block, "img", attrs, "", false)
		r.r.WriteElement(block, "img", attrs, "", true)
	}
	return true
}

// RenderPage renders BlockPage
func (r *Converter) RenderPage(block *notionapi.Block) bool {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		// title := html.EscapeString(block.Title)
		attrs := []string{"class", "notion-page"}
		r.r.WriteElement(block, "div", attrs, "", true)
		r.r.RenderChildren(block)
		r.r.WriteElement(block, "div", attrs, "", false)
		return true
	}

	var cls string
	if tp == notionapi.BlockPageSubPage {
		cls = "page"
	} else if tp == notionapi.BlockPageLink {
		cls = "page-link"
	} else {
		panic("unexpected page type")
	}

	url, title := r.getURLAndTitleForBlock(block)
	title = html.EscapeString(title)
	content := fmt.Sprintf(`<a href="%s">%s</a>`, url, title)
	attrs := []string{"class", cls}
	title = html.EscapeString(title)
	r.r.WriteElement(block, "div", attrs, content, true)
	r.r.WriteElement(block, "div", attrs, content, false)
	return true
}

// RenderCode renders BlockCode
func (r *Converter) RenderCode(block *notionapi.Block) bool {
	// code := html.EscapeString(block.Code)
	// fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
	// <pre class="%s">
	// %s
	// </pre>`, levelCls, block.CodeLanguage, levelCls, code)
	err := htmlHighlight(r.r.Buf, string(block.Code), block.CodeLanguage, "")
	panicIfErr(err)
	return true
}

// if returns false, the block will be rendered with default
func (r *Converter) blockRenderOverride(block *notionapi.Block) bool {
	if r.article.shouldSkipBlock(block) {
		return true
	}
	if r.renderGallery(block) {
		return true
	}
	switch block.Type {
	case notionapi.BlockPage:
		return r.RenderPage(block)
	case notionapi.BlockCode:
		return r.RenderCode(block)
	case notionapi.BlockImage:
		return r.RenderImage(block)
	}
	return false
}

// NewHTMLConverter returns new HTMLGenerator
func NewHTMLConverter(c *notionapi.Client, article *Article) *Converter {
	res := &Converter{
		notionClient: c,
		article:      article,
		page:         article.page,
	}

	r := tohtml.NewConverter(article.page)
	notionapi.PanicOnFailures = true
	r.AddIDAttribute = true
	r.RenderBlockOverride = res.blockRenderOverride
	r.RewriteURL = res.rewriteURL
	res.r = r

	return res
}

// Gen returns generated HTML
func (r *Converter) GenereateHTML() []byte {
	inner := string(r.r.ToHTML())
	page := r.page.Root()
	f := page.FormatPage()
	isMono := f != nil && f.PageFont == "mono"

	s := `<p></p>`
	if isMono {
		s += `<div style="font-family: monospace">`
	}
	s += inner
	if isMono {
		s += `</div>`
	}
	return []byte(s)
}

func notionToHTML(c *notionapi.Client, article *Article, articles *Articles) ([]byte, []*ImageMapping) {
	//fmt.Printf("notionToHTML: %s\n", notionapi.ToNoDashID(article.ID))
	r := NewHTMLConverter(c, article)
	if articles != nil {
		r.idToArticle = func(id string) *Article {
			return articles.idToArticle[id]
		}
	}
	return r.GenereateHTML(), r.article.Images
}
