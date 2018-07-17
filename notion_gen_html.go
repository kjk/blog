package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/kjk/notionapi"
)

// HTMLGenerator is for notion -> HTML generation
type HTMLGenerator struct {
	f        *bytes.Buffer
	pageInfo *notionapi.PageInfo
	level    int
	err      error
}

// NewHTMLGenerator returns new HTMLGenerator
func NewHTMLGenerator(pageInfo *notionapi.PageInfo) *HTMLGenerator {
	return &HTMLGenerator{
		f:        &bytes.Buffer{},
		pageInfo: pageInfo,
	}
}

// Gen returns generated HTML
func (g *HTMLGenerator) Gen() []byte {
	g.genContent(g.pageInfo.Page)
	return g.f.Bytes()
}

func (g *HTMLGenerator) genInlineBlock(b *notionapi.InlineBlock) error {
	var start, close string
	if b.AttrFlags&notionapi.AttrBold != 0 {
		start += "<b>"
		close += "</b>"
	}
	if b.AttrFlags&notionapi.AttrItalic != 0 {
		start += "<i>"
		close += "</i>"
	}
	if b.AttrFlags&notionapi.AttrStrikeThrought != 0 {
		start += "<strike>"
		close += "</strike>"
	}
	if b.AttrFlags&notionapi.AttrCode != 0 {
		start += "<code>"
		close += "</code>"
	}
	skipText := false
	if b.Link != "" {
		start += fmt.Sprintf(`<a href="%s">%s</a>`, b.Link, b.Text)
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="user">@%s</span>`, b.UserID)
		skipText = true
	}
	if b.Date != nil {
		// TODO: serialize date properly
		start += fmt.Sprintf(`<span class="date">@TODO: date</span>`)
		skipText = true
	}
	if !skipText {
		start += b.Text
	}
	_, err := io.WriteString(g.f, start+close)
	return err
}

func (g *HTMLGenerator) genInlineBlocks(blocks []*notionapi.InlineBlock) error {
	for _, block := range blocks {
		err := g.genInlineBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *HTMLGenerator) genBlockSurrouded(block *notionapi.Block, start, close string) {
	io.WriteString(g.f, start+"\n")
	g.genInlineBlocks(block.InlineContent)
	g.level++
	g.genContent(block)
	g.level--
	io.WriteString(g.f, close+"\n")
}

func (g *HTMLGenerator) genBlock(block *notionapi.Block) {
	levelCls := ""
	if g.level > 0 {
		levelCls = fmt.Sprintf(" lvl%d", g.level)
	}

	switch block.Type {
	case notionapi.BlockText:
		start := fmt.Sprintf(`<p>`)
		close := `</p>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockHeader:
		start := fmt.Sprintf(`<h1 class="hdr%s">`, levelCls)
		close := `</h1>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockSubHeader:
		start := fmt.Sprintf(`<h2 class="hdr%s">`, levelCls)
		close := `</h2>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockTodo:
		clsChecked := ""
		if block.IsChecked {
			clsChecked = " todo-checked"
		}
		start := fmt.Sprintf(`<div class="todo%s%s">`, levelCls, clsChecked)
		close := `</div>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockToggle:
		start := fmt.Sprintf(`<div class="toggle%s">`, levelCls)
		close := `</div>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockQuote:
		start := fmt.Sprintf(`<blockquote class="%s">`, levelCls)
		close := `</blockquote>`
		g.genBlockSurrouded(block, start, close)
	case notionapi.BlockDivider:
		fmt.Fprintf(g.f, `<hr class="%s"/>`+"\n", levelCls)
	case notionapi.BlockPage:
		id := strings.TrimSpace(block.ID)
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page-link"
		}
		title := template.HTMLEscapeString(block.Title)
		url := normalizeID(id) + ".html"
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		fmt.Fprintf(g.f, "%s\n", html)
	case notionapi.BlockCode:
		/*
			code := template.HTMLEscapeString(block.Code)
			fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
			<pre class="%s">
			%s
			</pre>`, levelCls, block.CodeLanguage, levelCls, code)
		*/
		htmlHighlight(g.f, string(block.Code), block.CodeLanguage, "")
	case notionapi.BlockBookmark:
		fmt.Fprintf(g.f, `<div class="bookmark %s">Bookmark to %s</div>`+"\n", levelCls, block.Link)
	case notionapi.BlockGist:
		fmt.Fprintf(g.f, `<div class="gist %s">Gist for %s</div>`+"\n", levelCls, block.Source)
	case notionapi.BlockImage:
		link := block.ImageURL
		fmt.Fprintf(g.f, `<img class="%s" src="%s" />`+"\n", levelCls, link)
	case notionapi.BlockColumnList:
		// TODO: implement me
	case notionapi.BlockCollectionView:
		// TODO: implement me
	default:
		fmt.Printf("Unsupported block type '%s', id: %s\n", block.Type, block.ID)
		panic(fmt.Sprintf("Unsupported block type '%s'", block.Type))
	}
}

func (g *HTMLGenerator) genBlocks(blocks []*notionapi.Block) {
	for len(blocks) > 0 {
		block := blocks[0]
		if block == nil {
			fmt.Printf("Missing block\n")
			blocks = blocks[1:]
			continue
		}

		if block.Type == notionapi.BlockNumberedList {
			fmt.Fprintf(g.f, `<ol>`)
			for len(blocks) > 0 {
				block := blocks[0]
				if block.Type != notionapi.BlockNumberedList {
					break
				}
				g.genBlockSurrouded(block, `<li>`, `</li>`)
				blocks = blocks[1:]
			}
			fmt.Fprintf(g.f, `</ol>`)
		} else if block.Type == notionapi.BlockBulletedList {
			fmt.Fprintf(g.f, `<ul>`)
			for len(blocks) > 0 {
				block := blocks[0]
				if block.Type != notionapi.BlockBulletedList {
					break
				}
				g.genBlockSurrouded(block, `<li>`, `</li>`)
				blocks = blocks[1:]
			}
			fmt.Fprintf(g.f, `</ul>`)
		} else {
			g.genBlock(block)
			blocks = blocks[1:]
		}
	}
}

func (g *HTMLGenerator) genContent(parent *notionapi.Block) {
	g.genBlocks(parent.Content)
}
