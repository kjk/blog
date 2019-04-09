package main

import (
	"fmt"
	"html"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

// ImageMapping keeps track of rewritten image urls (locally cached
// images in notion)
type ImageMapping struct {
	path        string
	relativeURL string
}

// HTMLRenderer keeps data
type HTMLRenderer struct {
	page         *notionapi.Page
	notionClient *notionapi.Client
	idToArticle  func(string) *Article
	images       []ImageMapping

	r *tohtml.HTMLRenderer
}

// only hex chars seem to be valid
func isValidNotionIDChar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		// currently not used but just in case they change their minds
		return true
	}
	return false
}

func isValidNotionID(id string) bool {
	// len("ea07db1b9bff415ab180b0525f3898f6")
	if len(id) != 32 {
		return false
	}
	for i := range id {
		if !isValidNotionIDChar(id[i]) {
			return false
		}
	}
	return true
}

// https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// https://www.notion.so/c674bebe8adf44d18c3a36cc18c131e2
// returns "" if didn't detect valid notion id in the url
func extractNotionIDFromURL(uri string) string {
	trimmed := strings.TrimPrefix(uri, "https://www.notion.so/")
	if uri == trimmed {
		return ""
	}
	// could be c674bebe8adf44d18c3a36cc18c131e2 from https://www.notion.so/c674bebe8adf44d18c3a36cc18c131e2
	id := trimmed
	parts := strings.Split(trimmed, "-")
	n := len(parts)
	if n >= 2 {
		// could be ea07db1b9bff415ab180b0525f3898f6 from Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
		id = parts[n-1]
	}
	id = notionapi.ToNoDashID(id)
	if !isValidNotionID(id) {
		return ""
	}
	return id
}

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// /article/${id}
func (r *HTMLRenderer) maybeReplaceNotionLink(uri string) string {
	id := extractNotionIDFromURL(uri)
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

// renderInlineLink renders a link in inline block
// we replace inter-notion urls to inter-blog urls
func (r *HTMLRenderer) renderInlineLink(b *notionapi.InlineBlock) (string, bool) {
	link := r.maybeReplaceNotionLink(b.Link)
	text := html.EscapeString(b.Text)
	s := fmt.Sprintf(`<a href="%s">%s</a>`, link, text)
	return s, true
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
		// title := template.HTMLEscapeString(block.Title)
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
	title = template.HTMLEscapeString(title)
	r.r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderCode renders BlockCode
func (r *HTMLRenderer) RenderCode(block *notionapi.Block, entering bool) bool {
	// code := template.HTMLEscapeString(block.Code)
	// fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
	// <pre class="%s">
	// %s
	// </pre>`, levelCls, block.CodeLanguage, levelCls, code)
	if entering {
		htmlHighlight(r.r.Buf, string(block.Code), block.CodeLanguage, "")
	}
	return true
}

var (
	toggleEntering = `
<div style="width: 100%%; margin-top: 2px; margin-bottom: 1px;">
    <div style="display: flex; align-items: flex-start; width: 100%%; padding-left: 2px; color: rgb(66, 66, 65);">

        <div style="margin-right: 4px; width: 24px; flex-grow: 0; flex-shrink: 0; display: flex; align-items: center; justify-content: center; min-height: calc((1.5em + 3px) + 3px); padding-right: 2px;">
            <div id="toggle-toggle-{{id}}" onclick="javascript:onToggleClick(this)" class="toggler" style="align-items: center; user-select: none; display: flex; width: 1.25rem; height: 1.25rem; justify-content: center; flex-shrink: 0;">

                <svg id="toggle-closer-{{id}}" width="100%%" height="100%%" viewBox="0 0 100 100" style="fill: currentcolor; display: none; width: 0.6875em; height: 0.6875em; transition: transform 300ms ease-in-out; transform: rotateZ(180deg);">
                    <polygon points="5.9,88.2 50,11.8 94.1,88.2 "></polygon>
                </svg>

                <svg id="toggle-opener-{{id}}" width="100%%" height="100%%" viewBox="0 0 100 100" style="fill: currentcolor; display: block; width: 0.6875em; height: 0.6875em; transition: transform 300ms ease-in-out; transform: rotateZ(90deg);">
                    <polygon points="5.9,88.2 50,11.8 94.1,88.2 "></polygon>
                </svg>
            </div>
        </div>

        <div style="flex: 1 1 0px; min-width: 1px;">
            <div style="display: flex;">
                <div style="padding-top: 3px; padding-bottom: 3px">{{inline}}</div>
            </div>

            <div style="margin-left: -2px; display: none" id="toggle-content-{{id}}">
                <div style="display: flex; flex-direction: column;">
                    <div style="width: 100%%; margin-top: 2px; margin-bottom: 0px;">
                        <div style="color: rgb(66, 66, 65);">
							<div style="">
`
	toggleClosing = `
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
`
)

// RenderToggle renders BlockToggle blocks
func (r *HTMLRenderer) RenderToggle(block *notionapi.Block, entering bool) bool {
	panicIf(block.Type != notionapi.BlockToggle, "unexpected block type '%s'", block.Type)

	if entering {
		// TODO: could do it without pushing buffers
		r.r.PushNewBuffer()
		r.r.RenderInlines(block.InlineContent)
		inline := r.r.PopBuffer().String()
		id := notionapi.ToNoDashID(block.ID)
		s := strings.Replace(toggleEntering, "{{id}}", id, -1)
		s = strings.Replace(s, "{{inline}}", inline, -1)
		r.r.WriteString(s)

	} else {
		r.r.WriteString(toggleClosing)
	}
	// we handled it
	return true
}

func (r *HTMLRenderer) blockRenderOverride(block *notionapi.Block, entering bool) bool {
	switch block.Type {
	case notionapi.BlockPage:
		return r.RenderPage(block, entering)
	case notionapi.BlockCode:
		return r.RenderCode(block, entering)
	case notionapi.BlockToggle:
		return r.RenderToggle(block, entering)
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
	}

	r := tohtml.NewHTMLRenderer(page)
	r.PanicOnFailures = true
	r.AddIDAttribute = true
	r.Data = res
	r.RenderBlockOverride = res.blockRenderOverride
	r.RenderInlineLinkOverride = res.renderInlineLink

	res.r = r
	return res
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
