package main

import (
	"fmt"
	"html"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

// ImageMapping keeps track of rewritten image urls (locally cached
// images in notion)
type ImageMapping struct {
	path        string
	relativeURL string
}

type BlockInfo struct {
	shouldSkip bool
}

// HTMLRenderer renders article as html
type HTMLRenderer struct {
	page         *notionapi.Page
	notionClient *notionapi.Client
	idToArticle  func(string) *Article
	images       []ImageMapping

	blockInfos map[*notionapi.Block]*BlockInfo

	r *tohtml.HTMLRenderer
}

func isEmptyTextBlock(b *notionapi.Block) bool {
	if b.Type != notionapi.BlockText {
		return false
	}
	if len(b.InlineContent) > 0 {
		return false
	}
	return true
}

func (r *HTMLRenderer) removeEmptyTextBlocksAtEnd(root *notionapi.Block) {
	n := len(root.Content)
	blocks := root.Content
	for i := 0; i < n; i++ {
		idx := n - 1 - i
		block := blocks[idx]
		if !isEmptyTextBlock(block) {
			return
		}
		r.markBlockToSkip(block)
	}
}

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// /article/${id}
func (r *HTMLRenderer) rewriteURL(uri string) string {
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

func (r *HTMLRenderer) getURLAndTitleForBlock(block *notionapi.Block) (string, string) {
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

// RenderImage renders BlockImage
func (r *HTMLRenderer) RenderImage(block *notionapi.Block, entering bool) bool {
	link := block.Source
	path, err := downloadAndCacheImage(r.notionClient, link)
	if err != nil {
		lg("genImage: downloadAndCacheImage('%s') from page https://notion.so/%s failed with '%s'\n", link, normalizeID(r.page.ID), err)
		panicIfErr(err)
		return false
	}
	relURL := "/img/" + filepath.Base(path)
	im := ImageMapping{
		path:        path,
		relativeURL: relURL,
	}
	r.images = append(r.images, im)
	attrs := []string{"class", "blog-img", "src", relURL}
	r.r.WriteElement(block, "img", attrs, "", entering)
	return true
}

// RenderPage renders BlockPage
func (r *HTMLRenderer) RenderPage(block *notionapi.Block, entering bool) bool {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		// title := html.EscapeString(block.Title)
		attrs := []string{"class", "notion-page"}
		r.r.WriteElement(block, "div", attrs, "", entering)
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
	r.r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderCode renders BlockCode
func (r *HTMLRenderer) RenderCode(block *notionapi.Block, entering bool) bool {
	// code := html.EscapeString(block.Code)
	// fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
	// <pre class="%s">
	// %s
	// </pre>`, levelCls, block.CodeLanguage, levelCls, code)
	if entering {
		htmlHighlight(r.r.Buf, string(block.Code), block.CodeLanguage, "")
	}
	return true
}

// if returns false, the block will be rendered with default
func (r *HTMLRenderer) blockRenderOverride(block *notionapi.Block, entering bool) bool {
	if r.shouldSkipBlock(block) {
		return true
	}
	switch block.Type {
	case notionapi.BlockPage:
		return r.RenderPage(block, entering)
	case notionapi.BlockCode:
		return r.RenderCode(block, entering)
	case notionapi.BlockImage:
		return r.RenderImage(block, entering)
	}
	return false
}

// NewHTMLRenderer returns new HTMLGenerator
func NewHTMLRenderer(c *notionapi.Client, page *notionapi.Page) *HTMLRenderer {
	res := &HTMLRenderer{
		notionClient: c,
		page:         page,
		blockInfos:   map[*notionapi.Block]*BlockInfo{},
	}

	r := tohtml.NewHTMLRenderer(page)
	notionapi.PanicOnFailures = true
	r.AddIDAttribute = true
	r.RenderBlockOverride = res.blockRenderOverride
	r.RewriteURL = res.rewriteURL
	res.r = r

	res.removeEmptyTextBlocksAtEnd(page.Root)

	return res
}

func (r *HTMLRenderer) getBlockInfo(block *notionapi.Block) *BlockInfo {
	bi := r.blockInfos[block]
	if bi == nil {
		bi = &BlockInfo{}
		r.blockInfos[block] = bi
	}
	return bi
}

func (r *HTMLRenderer) markBlockToSkip(block *notionapi.Block) {
	r.getBlockInfo(block).shouldSkip = true
}

func (r *HTMLRenderer) shouldSkipBlock(block *notionapi.Block) bool {
	bi := r.blockInfos[block]
	if bi == nil {
		return false
	}
	return bi.shouldSkip
}

// Gen returns generated HTML
func (r *HTMLRenderer) Gen() []byte {
	inner := string(r.r.ToHTML())
	page := r.page.Root
	f := page.FormatPage
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

func notionToHTML(c *notionapi.Client, page *notionapi.Page, articles *Articles) ([]byte, []ImageMapping) {
	r := NewHTMLRenderer(c, page)
	if articles != nil {
		r.idToArticle = func(id string) *Article {
			return articles.idToArticle[id]
		}
	}
	return r.Gen(), r.images
}
