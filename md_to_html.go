package main

import (
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(), html.TabWidth(2))
	panicIf(htmlFormatter == nil, "couldn't create html formatter")
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
	panicIf(highlightStyle == nil, "didn't find style '%s'", styleName)

}

// based on https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func makeRenderHookCodeBlock(defaultLang string) mdhtml.RenderNodeFunc {
	return func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		codeBlock, ok := node.(*ast.CodeBlock)
		if !ok {
			return ast.GoToNext, false
		}
		lang := string(codeBlock.Info)
		if false {
			verbose("lang: '%s', code: %s\n", lang, string(codeBlock.Literal[:16]))
			io.WriteString(w, "\n<pre class=\"chroma\"><code>")
			mdhtml.EscapeHTML(w, codeBlock.Literal)
			io.WriteString(w, "</code></pre>\n")
		} else {
			htmlHighlight(w, string(codeBlock.Literal), lang, defaultLang)
		}
		return ast.GoToNext, true
	}
}

func markdownToUnsafeHTML(md []byte, defaultLang string) []byte {
	extensions := parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings |
		parser.NoEmptyLineBeforeBlock
	parser := parser.NewWithExtensions(extensions)

	htmlFlags := mdhtml.Smartypants |
		mdhtml.SmartypantsFractions |
		mdhtml.SmartypantsDashes |
		mdhtml.SmartypantsLatexDashes
	htmlOpts := mdhtml.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: makeRenderHookCodeBlock(defaultLang),
	}
	renderer := mdhtml.NewRenderer(htmlOpts)
	return markdown.ToHTML(md, parser, renderer)
}

func markdownToHTML(d []byte, defaultLang string) string {
	unsafe := markdownToUnsafeHTML(d, defaultLang)
	policy := bluemonday.UGCPolicy()
	policy.AllowStyling()
	policy.RequireNoFollowOnFullyQualifiedLinks(false)
	policy.RequireNoFollowOnLinks(false)
	res := policy.SanitizeBytes(unsafe)
	return string(res)
}
