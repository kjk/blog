package textiler

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// TODO: a notion of extended block (e.g. pre..) which lasts until the next
// block starts

const (
	// renderer flags
	RENDERER_XHTML = 1 << iota
)

var newline = []byte{'\n'}

type UrlRef struct {
	url  []byte
	name []byte
}

type TextileRenderer struct {
}

type TextileParser struct {
	flags int

	refs map[string]*UrlRef

	// are we inside <p> tag?
	inP bool

	out *bytes.Buffer

	// if we're parsing <ol> list, this tells us current nesting level
	olLevel int
	ulLevel int

	blockLineNo    int
	blockTags      []string
	dumpLines      bool
	dumpParagraphs bool
}

func tagSkipPre(tag string) bool {
	if tag == "pre" {
		return true
	}
	if tag == "code" {
		return true
	}
	return false
}

func (p *TextileParser) pushBlockTag(tag string) {
	if tagSkipPre(tag) {
		p.blockTags = append(p.blockTags, tag)
	}
}

func (p *TextileParser) lastBlockTagIs(tag string) bool {
	if len(p.blockTags) == 0 {
		return false
	}
	last := p.blockTags[len(p.blockTags)-1]
	return last == tag
}

func (p *TextileParser) popBlockTag(tag string) bool {
	if !tagSkipPre(tag) {
		return false
	}
	if p.lastBlockTagIs(tag) {
		p.blockTags = p.blockTags[0 : len(p.blockTags)-1]
		return true
	}
	return false
}

func (p *TextileParser) inHtmlCode() bool {
	return p.lastBlockTagIs("code")
}

func (p *TextileParser) inHtmlPre() bool {
	return p.lastBlockTagIs("pre")
}

func (p *TextileParser) inHtmlBlock() bool {
	return len(p.blockTags) > 0
}

func (p *TextileParser) isFlagSet(flag int) bool {
	return p.flags&flag != 0
}

func (p *TextileParser) isXhtml() bool {
	return p.isFlagSet(RENDERER_XHTML)
}

func NewParser(flags int) *TextileParser {
	return &TextileParser{
		flags:     flags,
		refs:      make(map[string]*UrlRef),
		out:       new(bytes.Buffer),
		blockTags: make([]string, 0),
	}
}

var punctAndSpace = []byte(".,\"'?!;:() \t")

func isValidTag(tag []byte) bool {
	_, ok := blockTags[string(tag)]
	return ok
}

func parseStartTag(l []byte) (rest, html, tag []byte) {
	if !startsWithByte(l, '<', 3) {
		return nil, nil, nil
	}
	// TODO: this is not correct, > might be inside an attribute
	htmlEnd := bytes.IndexByte(l, '>')
	if htmlEnd == -1 {
		return nil, nil, nil
	}
	html = l[:htmlEnd+1]
	// TODO: this is incorrect, space might be inside an attribute
	tagEnd := bytes.IndexByte(l, ' ')
	if tagEnd == -1 {
		tagEnd = htmlEnd
	}
	tag = l[1:tagEnd]
	if !isValidTag(tag) {
		return nil, nil, nil
	}
	return l[htmlEnd+1:], html, tag
}

func parseEndTag(l []byte) (rest, html, tag []byte) {
	if !startsWithByte(l, '<', 4) {
		return nil, nil, nil
	}
	if l[1] != '/' {
		return nil, nil, nil
	}
	htmlEnd := bytes.IndexByte(l, '>')
	if htmlEnd == -1 {
		return nil, nil, nil
	}
	tag = l[2:htmlEnd]
	if !isValidTag(tag) {
		return nil, nil, nil
	}
	return l[htmlEnd+1:], l[:htmlEnd+1], tag
}

const (
	cr = 0xd
	lf = 0xa
)

// find a end of line (cr, lf or crlf). Return the line
// and the remaining of data (without the end-of-line character(s))
func extractLine(d []byte) ([]byte, []byte) {
	if d == nil || len(d) == 0 {
		return nil, nil
	}
	wasCr := false
	pos := -1
	for i := 0; i < len(d); i++ {
		if d[i] == cr || d[i] == lf {
			wasCr = (d[i] == cr)
			pos = i
			break
		}
	}
	if pos == -1 {
		return d, nil
	}
	line := d[:pos]
	rest := d[pos+1:]
	if wasCr && len(rest) > 0 && rest[0] == lf {
		rest = rest[1:]
	}
	return line, rest
}

func splitIntoLines(d []byte) [][]byte {
	res := make([][]byte, 0)
	var l []byte
	for {
		l, d = extractLine(d)
		if l == nil {
			return res
		}
		res = append(res, l)
	}
	panic("")
	return res
}

func parseHtml(l []byte) (rest, html, tag []byte, start bool) {
	if rest, html, tag := parseEndTag(l); rest != nil {
		return rest, html, tag, false
	}
	if rest, html, tag := parseStartTag(l); rest != nil {
		return rest, html, tag, true
	}
	return nil, nil, nil, false
}

// ![>|<|=]{$styleOpt}($classOpt)$imgSrc($altOptional)!:$urlOptional
// TODO: should return nil for alt instead of empty slice if not found?
func parseImg(l []byte) (rest, url, imgSrc, alt []byte, attrs *AttributesOpt) {
	//fmt.Printf("l: '%s'\n", string(l))
	if len(l) < 3 {
		return nil, nil, nil, nil, nil
	}
	if l[0] != '!' {
		return nil, nil, nil, nil, nil
	}
	l = l[1:]
	l, attrs = parseAttributesOpt(l, true)
	endIdx := bytes.IndexByte(l, '!')
	if endIdx == -1 {
		return nil, nil, nil, nil, nil
	}
	imgSrc = l[:endIdx]
	l = l[endIdx+1:]
	endIdx = bytes.IndexByte(imgSrc, '(')
	if endIdx != -1 {
		alt = imgSrc[endIdx+1:]
		imgSrc = imgSrc[:endIdx]
		endIdx = bytes.IndexByte(alt, ')')
		if endIdx == -1 {
			return nil, nil, nil, nil, nil
		}
		alt = alt[:endIdx]
	}
	if len(l) > 0 && l[0] == ':' {
		l, url = extractUrlOrRefName(l[1:])
	}
	return l, url, imgSrc, alt, attrs
}

func extractUntil(l []byte, c byte) (rest, inside []byte) {
	idx := bytes.IndexByte(l, c)
	if idx == -1 {
		return nil, nil
	}
	return l[idx+1:], l[:idx]
}

func endsWithPunctOrSpace(l []byte) bool {
	n := len(l)
	if n == 0 {
		return true
	}
	c := l[n-1]
	// TODO: speed up
	return bytes.IndexByte(punctAndSpace, c) != -1
}

func isPunctOrSpace(c byte) bool {
	// TODO: speed up
	return bytes.IndexByte(punctAndSpace, c) != -1
}

// $start$inside$end$rest
// e.g. '@foo@bar'
func extractInside(l []byte, start, end byte) (rest, inside []byte) {
	if len(l) == 0 || l[0] != start {
		return nil, nil
	}
	return extractUntil(l[1:], end)
}

func startsWithByte(s []byte, b byte, minLen int) bool {
	return len(s) >= minLen && s[0] == b
}

func byteConcat(b1, b2 []byte) []byte {
	if b1 == nil && b2 == nil {
		return nil
	}
	if b1 == nil {
		return b2
	}
	if b2 == nil {
		return b1
	}
	return append(b1, b2...)
}

type PaddingInfo struct {
	alignLeft    bool
	alignRight   bool
	alignCenter  bool
	alignJustify bool
	paddingLeft  int
	paddingRight int
}

func formatPaddingInfo(pi PaddingInfo, forImg bool) []byte {
	s := ""
	if pi.paddingLeft > 0 {
		s += fmt.Sprintf("padding-left:%dem;", pi.paddingLeft)
	}
	if pi.paddingRight > 0 {
		s += fmt.Sprintf("padding-right:%dem;", pi.paddingRight)
	}
	if pi.alignLeft {
		if forImg {
			s += "float: left;"
		} else {
			s += "text-align:left;"
		}
	}
	if pi.alignRight {
		if forImg {
			s += "float: right;"
		} else {
			s += "text-align:right;"
		}
	}
	if pi.alignJustify {
		s += "text-align:justify;"
	}
	if pi.alignCenter {
		if forImg {
			s += "display: block; margin: 0 auto;"
		} else {
			s += "text-align:center;"
		}
	}
	if len(s) == 0 {
		return nil
	}
	return []byte(s)
}

func countRepeatedChars(l []byte, c byte) (rest []byte, n int) {
	for n, b := range l {
		if b != c {
			return l[n:], n
		}
	}
	return l[0:0], len(l)
}

func extractShortStyle(l []byte, forImg bool) (rest, style []byte) {
	var pi PaddingInfo
	for len(l) > 0 {
		c := l[0]
		if c == '<' {
			if len(l) > 1 && l[1] == '>' {
				l = l[2:]
				pi.alignJustify = true
			} else {
				l = l[1:]
				pi.alignLeft = true
			}
		} else if c == '>' {
			l = l[1:]
			pi.alignRight = true
		} else if c == '=' {
			l = l[1:]
			pi.alignCenter = true
		} else if c == '(' {
			l, pi.paddingLeft = countRepeatedChars(l, '(')
		} else if c == ')' {
			l, pi.paddingRight = countRepeatedChars(l, ')')
		} else {
			break
		}
	}
	return l, formatPaddingInfo(pi, forImg)
}

// TODO: it's possible this list is not complete
func isStyleChar(c byte) bool {
	return isChar(c) || isDigit(c) ||
		(-1 != bytes.IndexByte([]byte{'#', '-', ':', ';'}, c))
}

// {$style}$rest
func extractStyleOpt(l []byte) (rest, styleOpt []byte) {
	if !startsWithByte(l, '{', 3) {
		return l, nil
	}
	if l[1] == '}' {
		return l, nil
	}
	for i := 1; i < len(l); i++ {
		if !isStyleChar(l[i]) {
			if l[i] == '}' {
				rest, styleOpt = l[i+1:], l[1:i]
				if !endsWithByte(styleOpt, ';') {
					styleOpt = append(styleOpt, ';')
				}
				return rest, styleOpt
			}
		}
	}
	return l, nil
}

// TODO: it's possible this list is not complete
func isLangChar(c byte) bool {
	return isChar(c) ||
		(-1 != bytes.IndexByte([]byte{'-'}, c))
}

// [$lang]$rest
func extractLangOpt(l []byte) (rest, langOpt []byte) {
	if !startsWithByte(l, '[', 3) {
		return l, nil
	}
	if l[1] == ']' {
		return l, nil
	}
	for i := 1; i < len(l); i++ {
		if !isLangChar(l[i]) {
			if l[i] == ']' {
				return l[i+1:], l[1:i]
			}
		}
	}
	return l, nil
}

func isChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// TODO: it's possible this list is not complete
func isClassChar(c byte) bool {
	return isChar(c) || isDigit(c) || c == '#' || c == '-'
}

// ($class)$rest
func extractClassOpt(l []byte) (rest []byte, classOpt []byte) {
	if !startsWithByte(l, '(', 3) {
		return l, nil
	}
	if l[1] == ')' {
		return l, nil
	}
	for i := 1; i < len(l); i++ {
		if !isClassChar(l[i]) {
			if l[i] == ')' {
				return l[i+1:], l[1:i]
			}
		}
	}
	return l, nil
}

type AttributesOpt struct {
	class []byte
	style []byte
	lang  []byte
}

// ($classOpt){$styleOpt}[$langOpt]
func parseAttributesOpt(l []byte, forImg bool) (rest []byte, attrs *AttributesOpt) {
	if len(l) == 0 {
		return l, nil
	}
	attrs = &AttributesOpt{}
	if l[0] == '(' {
		l, attrs.class = extractClassOpt(l)
	}
	l, style := extractShortStyle(l, forImg)

	for len(l) > 0 {
		n := len(l)
		switch l[0] {
		case '(':
			l, attrs.class = extractClassOpt(l)
		case '[':
			l, attrs.lang = extractLangOpt(l)
		case '{':
			l, attrs.style = extractStyleOpt(l)
		}
		if n == len(l) {
			break
		}
	}

	attrs.style = byteConcat(attrs.style, style)
	return l, attrs
}

// %($classOpt){$styleOpt}[$langOpt]$inside%$rest
func parseSpan(l []byte) (rest, inside []byte, attrs *AttributesOpt) {
	if !startsWithByte(l, '%', 3) {
		return nil, nil, nil
	}
	l = l[1:]
	l, attrs = parseAttributesOpt(l, false)
	rest, inside = extractUntil(l, '%')
	return rest, inside, attrs
}

func endsWithByte(s []byte, b byte) bool {
	if s == nil || len(s) == 0 {
		return false
	}
	n := len(s) - 1
	return s[n] == b
}

// @$inside@$rest
func parseCode(l []byte) (rest, inside []byte) {
	return extractInside(l, '@', '@')
}

// h${n}($classOpt){$styleOpt}[$langOpt]. $rest
func parseH(l []byte) (rest []byte, level int, attrs *AttributesOpt) {
	if !startsWithByte(l, 'h', 4) {
		return l, -1, nil
	}
	n := l[1] - '0'
	if n < 1 || n > 6 {
		return l, -1, nil
	}
	l = l[2:]
	l, attrs = parseAttributesOpt(l, false)
	if len(l) < 2 || l[0] != '.' || l[1] != ' ' {
		return l, -1, nil
	}
	return l[2:], int(n), attrs
}

// TODO: this is more complex
func isUrlEnd(b byte) bool {
	return -1 != bytes.IndexByte([]byte{' ', '!', ')'}, b)
}

func detectUrl(l []byte) ([]byte, []byte) {
	i := bytes.Index(l, []byte{':', '/', '/'})
	if i == -1 {
		return nil, nil
	}
	s := string(l[:i])
	if !(s == "http" || s == "https") {
		return nil, nil
	}
	i += 3
	for i < len(l) {
		if isUrlEnd(l[i]) {
			return l[:i], l[i:]
		}
		i += 1
	}
	return l, l[0:0]
}

func isNotUrlEnd(b byte) bool {
	return -1 != bytes.IndexByte([]byte{'.', ':', ','}, b)
}

func extractUrlOrRefName(l []byte) (rest, urlOrRef []byte) {
	for i, c := range l {
		// TODO: hackish. Probably should test l[:i] against a list
		// of known refs
		if isUrlEnd(c) {
			// TODO: hack, if url ends with ".", don't count it as part of url
			if i > 0 && isNotUrlEnd(l[i-1]) {
				i -= 1
			}
			return l[i:], l[:i]
		}
	}
	// TODO: hack, if url ends with ".", don't count it as part of url
	i := len(l) - 1
	if i > 0 && isNotUrlEnd(l[i]) {
		return l[i:], l[:i]
	}
	return l[0:0], l
}

// "$title":$url or "$title":$refName
func parseUrlOrRefName(l []byte) (rest, title, urlOrRefName []byte) {
	//fmt.Printf("parseUrlOrRefName: '%s'\n", string(l))
	if len(l) < 4 {
		return nil, nil, nil
	}
	l, title = extractInside(l, '"', '"')
	if title == nil || !startsWithByte(l, ':', 1) {
		return nil, nil, nil
	}
	//fmt.Printf("  title: '%s'\n", string(title))
	rest, urlOrRefName = extractUrlOrRefName(l[1:])
	//fmt.Printf("  urlOrRefName: '%s'\n", string(urlOrRefName))
	//fmt.Printf("  rest: '%s'\n", string(rest))
	//if rest == nil {
	//	fmt.Print("  rest is nil!\n")
	//}
	return rest, title, urlOrRefName
}

// [$name]$url
func isUrlRef(l []byte) ([]byte, []byte) {
	if len(l) < 4 {
		return nil, nil
	}
	if l[0] != '[' {
		return nil, nil
	}
	l = l[1:]
	endIdx := bytes.IndexByte(l, ']')
	if endIdx == -1 {
		return nil, nil
	}
	name := l[:endIdx]
	l = l[endIdx+1:]
	url, rest := detectUrl(l)
	if url == nil || len(rest) > 0 {
		return nil, nil
	}
	return name, url
}

func startsWith(l, prefix []byte) (rest []byte) {
	if bytes.HasPrefix(l, prefix) {
		return l[len(prefix):]
	}
	return nil
}

// notextile. $rest
func parseNoTextile(l []byte) (rest []byte) {
	return startsWith(l, []byte("notextile. "))
}

// bq. $rest
func parseBlockQuote(l []byte) (rest []byte) {
	return startsWith(l, []byte("bq. "))
}

// ###. $rest
func parseComment(l []byte) (rest []byte) {
	return startsWith(l, []byte("###. "))
}

// pre.. $rest
func parsePre(l []byte) (rest []byte) {
	return startsWith(l, []byte("pre.. "))
}

// p($classOpt){$styleOpt}[$langOpt]. $rest
func parseP(l []byte) (rest []byte, attrs *AttributesOpt) {
	if !startsWithByte(l, 'p', 3) {
		return nil, nil
	}
	l = l[1:]
	l, attrs = parseAttributesOpt(l, false)
	if len(l) < 2 || l[0] != '.' || l[1] != ' ' {
		return nil, nil
	}
	return l[2:], attrs
}

func needsHtmlCodeEscaping(b byte) []byte {
	switch b {

	/*	case '"':
		return []byte("&quot;")*/
	case '&':
		return []byte("&amp;")
	case '<':
		return []byte("&lt;")
	case '>':
		return []byte("&gt;")
	case '\'':
		return []byte("&#8217;")
	}
	return nil
}

func (p *TextileParser) serHtmlRaw(before, htmlRaw, rest []byte) {
	p.serEscaped(before)
	p.out.Write(htmlRaw)
	p.parseInline(rest)
}

func needsEscaping(b byte) []byte {
	switch b {
	case '\'':
		return []byte("&#8217;")
	}
	return nil
}

func (p *TextileParser) serEscapedInContext(l []byte) {
	if p.inHtmlCode() || p.inHtmlPre() {
		p.serAsHtmlCode(l)
	} else {
		p.serEscaped(l)
	}
}

func (p *TextileParser) serEscaped(l []byte) {
	for _, b := range l {
		if esc := needsEscaping(b); esc != nil {
			p.out.Write(esc)
		} else {
			p.out.WriteByte(b)
		}
	}
}

func (p *TextileParser) serAsHtmlCode(l []byte) {
	for _, b := range l {
		if esc := needsHtmlCodeEscaping(b); esc != nil {
			p.out.Write(esc)
		} else {
			p.out.WriteByte(b)
		}
	}
}

// s is "$class[#$id]", we return ' class="$class" id="$id"'
func serClassOrIdOpt(s []byte) string {
	if s == nil || len(s) == 0 {
		return ""
	}
	idx := bytes.IndexByte(s, '#')
	if -1 == idx {
		return fmt.Sprintf(` class="%s"`, string(s))
	}
	if 0 == idx {
		return fmt.Sprintf(` id="%s"`, string(s[1:]))
	}
	return fmt.Sprintf(` class="%s" id="%s"`, string(s[:idx]), string(s[idx+1:]))
}

func serStyleOpt(s []byte) string {
	if s == nil || len(s) == 0 {
		return ""
	}
	s = prettyPrintStyle(s)
	return fmt.Sprintf(` style="%s"`, string(s))
}

func serLangOpt(s []byte) string {
	if s == nil || len(s) == 0 {
		return ""
	}
	return fmt.Sprintf(` lang="%s"`, string(s))
}

func serAttributesOpt(attrs *AttributesOpt) string {
	if attrs == nil {
		return ""
	}
	s1 := serClassOrIdOpt(attrs.class)
	s2 := serStyleOpt(attrs.style)
	s3 := serLangOpt(attrs.lang)
	return s1 + s2 + s3
}

func (p *TextileParser) serTag(tag string, attrs *AttributesOpt, before, inside, rest []byte) {
	p.serAsHtmlCode(before)
	p.out.WriteString(fmt.Sprintf("<%s%s>", tag, serAttributesOpt(attrs)))
	p.parseInline(inside)
	p.out.WriteString(fmt.Sprintf("</%s>", tag))
	p.parseInline(rest)
}

// TODO: change to serTag("span", ...) ?
func (p *TextileParser) serSpan(before, inside []byte, attrs *AttributesOpt, rest []byte) {
	p.serEscaped(before)
	attrsStr := serAttributesOpt(attrs)
	p.out.WriteString(fmt.Sprintf(`<span%s>`, attrsStr))
	p.parseInline(inside)
	p.out.WriteString("</span>")
	p.parseInline(rest)
}

func (p *TextileParser) serUrl(before, title, url, rest []byte) {
	p.serEscaped(before)
	p.out.WriteString(fmt.Sprintf(`<a href="%s">`, string(url)))
	p.serEscaped(title)
	p.out.WriteString("</a>")
	p.parseInline(rest)
}

func (p *TextileParser) serCode(before, inside, rest []byte) {
	p.serEscaped(before)
	p.out.WriteString(fmt.Sprintf(`<code>%s</code>`, string(inside)))
	p.parseInline(rest)
}

func (p *TextileParser) serImg(before []byte, imgSrc []byte, alt []byte, attrs *AttributesOpt, url []byte, rest []byte) {
	p.serEscaped(before)
	if len(url) > 0 {
		p.out.WriteString(fmt.Sprintf(`<a href="%s" class="img">`, string(url)))
	}
	s := ""
	if len(attrs.style) > 0 {
		s += fmt.Sprintf(` style="%s"`, string(attrs.style))
	}
	if len(attrs.class) > 0 {
		s += fmt.Sprintf(` class="%s"`, string(attrs.class))
	}
	if len(alt) > 0 {
		s += fmt.Sprintf(` title="%s"`, string(alt))
	}
	s += fmt.Sprintf(` alt="%s"`, string(alt))
	p.out.WriteString(fmt.Sprintf(`<img src="%s"%s`, string(imgSrc), s))
	if p.isXhtml() {
		p.out.WriteString(" />")
	} else {
		p.out.WriteString(">")
	}
	if len(url) > 0 {
		p.out.WriteString("</a>")
	}
	p.parseInline(rest)
}

func (p *TextileParser) serNoTextile(s []byte) {
	p.out.Write(s)
}

func prettyPrintStyle(s []byte) []byte {
	res := make([]byte, 0)
	state := 0 // 0 - regular, 1 - after ';'
	for _, b := range s {
		if state == 0 {
			res = append(res, b)
			if b == ';' {
				state = 1
			}
		} else {
			if b != ' ' {
				res = append(res, ' ')
				res = append(res, b)
				state = 0
			}
		}
	}
	if !endsWithByte(res, ';') {
		res = append(res, ';')
	}
	return res
}

func (p *TextileParser) serP(s []byte, attrs *AttributesOpt) {
	attrsStr := serAttributesOpt(attrs)
	p.out.WriteString(fmt.Sprintf("\t<p%s>", attrsStr))
	p.parseInline(s)
	p.out.WriteString("</p>")
}

func (p *TextileParser) serPre(s []byte) {
	p.out.WriteString(fmt.Sprintf("<pre>"))
	p.parseInline(s)
	p.out.WriteString(fmt.Sprintf("</pre>"))
}

func (p *TextileParser) serBlockQuote(s []byte) {
	p.out.WriteString("\t<blockquote>\n\t")
	p.serP(s, nil)
	p.out.WriteString("\n\t</blockquote>")
}

func (p *TextileParser) serH(rest []byte, n int, attrs *AttributesOpt) {
	s := serAttributesOpt(attrs)
	p.out.WriteString(fmt.Sprintf("\t<h%d%s>", n, s))
	p.out.Write(rest) // TODO: escape?
	p.out.WriteString(fmt.Sprintf("</h%d>", n))
}

func parseListEl(l []byte, r rune) (rest []byte, level int) {
	level = 0
	for {
		rune, size := utf8.DecodeRune(l)
		if rune == utf8.RuneError {
			return nil, 0
		}
		if rune != r {
			if rune == ' ' && level > 0 {
				return l[1:], level
			}
			return nil, 0
		}
		level += 1
		l = l[size:]
	}
	return nil, 0
}

func (p *TextileParser) closeOlIfNecessary() {
	for p.olLevel > 0 {
		p.olLevel -= 1
		// TODO: write me
		p.out.WriteString("</li>\n")
		p.out.WriteString("\t</ol>")
	}
}

func (p *TextileParser) closeUlIfNecessary() {
	for p.ulLevel > 0 {
		p.ulLevel -= 1
		// TODO: if it's only one list element, it should be:
		// </li></ul>, if more than one: </li>\n\t</ul>
		p.out.WriteString("</li>\n")
		p.out.WriteString("\t</ul>")
	}
}

func (p *TextileParser) serOl(l []byte, level int) {
	if p.olLevel > 0 {
		p.out.WriteString("</li>\n")
	}
	if level > p.olLevel {
		n := level - p.olLevel
		for n > 0 {
			p.out.WriteString("\t<ol>\n")
			n -= 1
		}
	}
	p.olLevel = level
	p.out.WriteString("\t\t<li>")
	p.parseInline(l)
}

func (p *TextileParser) serUl(l []byte, level int) {
	if p.ulLevel > 0 {
		p.out.WriteString("</li>\n")
	}
	if level > p.ulLevel {
		n := level - p.ulLevel
		for n > 0 {
			p.out.WriteString("\t<ul>\n")
			n -= 1
		}
	}
	p.ulLevel = level
	p.out.WriteString("\t\t<li>")
	p.parseInline(l)
}

// find qtag (or two) followed by punctuation
func parseQtagInside(l []byte, qtag byte, two bool) (rest, inside []byte) {
	for i, c := range l {
		if c != qtag {
			continue
		}
		rest = l[i+1:]
		if two {
			if !startsWithByte(rest, qtag, 1) {
				continue
			}
			rest = rest[1:]
		} else {
			if i == 0 {
				// this is for cases like __r __
				return nil, nil
			}
		}
		if len(rest) == 0 || isPunctOrSpace(rest[0]) {
			inside = l[:i]
			if len(inside) == 0 || inside[len(inside)-1] == ' ' {
				return nil, nil
			}
			return rest, inside
		}
	}
	return nil, nil
}

func (p *TextileParser) parseQtag(before, rest []byte, qtag byte, tag string) bool {
	if !endsWithPunctOrSpace(before) {
		return false
	}
	rest = rest[1:] // we know the first byte is qtag
	rest, attrs := parseAttributesOpt(rest, false)
	if rest, inside := parseQtagInside(rest, qtag, false); rest != nil {
		p.serTag(tag, attrs, before, inside, rest)
		return true
	}
	return false
}

func (p *TextileParser) parseQtag2(before, rest []byte, qtag byte, tag string) bool {
	if !endsWithPunctOrSpace(before) {
		return false
	}
	rest = rest[1:] // we know the first byte is qtag
	if !startsWithByte(rest, qtag, 1) {
		return false
	}
	rest = rest[1:]
	rest, attrs := parseAttributesOpt(rest, false)
	if rest, inside := parseQtagInside(rest, qtag, true); rest != nil {
		p.serTag(tag, attrs, before, inside, rest)
		return true
	}
	return false
}

var qtagToTag = map[byte]string{
	'-': "del",
	'+': "ins",
	'^': "sup",
	'~': "sub",
}

func (p *TextileParser) parseInline(l []byte) {
	for i := 0; i < len(l); i++ {
		b := l[i]

		switch b {
		case '-', '+', '^', '~':
			if p.parseQtag(l[:i], l[i:], b, qtagToTag[b]) {
				return
			}

		case '?':
			if p.parseQtag2(l[:i], l[i:], b, "cite") {
				return
			}

		case '_':
			if p.parseQtag2(l[:i], l[i:], b, "i") {
				return
			}
			if p.parseQtag(l[:i], l[i:], b, "em") {
				return
			}

		case '*':
			if p.parseQtag2(l[:i], l[i:], b, "b") {
				return
			}
			if p.parseQtag(l[:i], l[i:], b, "strong") {
				return
			}

		case '"':
			if rest, title, urlOrRefName := parseUrlOrRefName(l[i:]); rest != nil {
				url := urlOrRefName
				if urlRef, ok := p.refs[string(urlOrRefName)]; ok {
					url = urlRef.url
				}
				p.serUrl(l[:i], title, url, rest)
				return
			}

		case '!':
			if rest, url, imgSrc, alt, attrs := parseImg(l[i:]); rest != nil {
				p.serImg(l[:i], imgSrc, alt, attrs, url, rest)
				return
			}

		case '@':
			if rest, inside := parseCode(l[i:]); rest != nil {
				p.serCode(l[:i], inside, rest)
				return
			}

		case '%':
			if rest, inside, attrs := parseSpan(l[i:]); rest != nil {
				p.serSpan(l[:i], inside, attrs, rest)
				return
			}

		case '<':
			if rest, html, _, _ := parseHtml(l[i:]); rest != nil {
				p.parseInline(l[:i])
				p.serEscapedInContext(html)
				p.parseInline(rest)
				return
			}
		}
	}
	p.serAsHtmlCode(l)
	//p.serEscapedInContext(l)
	//p.serEscaped(l)
}

func (p *TextileParser) startNewLine() {
	if p.inHtmlBlock() {
		if p.blockLineNo > 1 {
			p.out.WriteString("\n")
		}
		return
	}
	if !p.inP {
		p.out.WriteString("\t<p>")
		p.inP = true
	} else {
		if p.isXhtml() {
			p.out.WriteString("<br />\n")
		} else {
			p.out.WriteString("<br>\n")
		}
	}
}

func (p *TextileParser) closeP() {
	if p.inP {
		p.out.WriteString("</p>")
		p.inP = false
	}
	p.out.WriteString("\n\n")
}

func (p *TextileParser) parseBlockStart(l []byte) (parsed bool) {
	rune, _ := utf8.DecodeRune(l)
	if rune == utf8.RuneError {
		return false
	}
	parsed = true
	switch rune {
	case 'h':
		if rest, n, attrs := parseH(l); n != -1 {
			p.serH(rest, n, attrs)
			return
		}
	case '<':
		if rest, html, tag, startTag := parseHtml(l); rest != nil {
			tagStr := string(tag)
			if startTag {
				p.pushBlockTag(tagStr)
			}
			p.startNewLine()
			p.serHtmlRaw(nil, html, rest)
			if !startTag {
				p.popBlockTag(tagStr)
			}
			return
		}
	case 'n':
		if rest := parseNoTextile(l); rest != nil {
			p.serNoTextile(rest)
			return
		}
	case 'p':
		if rest := parsePre(l); rest != nil {
			p.serPre(rest)
			return
		}
		if rest, attrs := parseP(l); rest != nil {
			p.serP(rest, attrs)
			return
		}
	case 'b':
		if rest := parseBlockQuote(l); rest != nil {
			p.serBlockQuote(rest)
			return
		}
	case '#':
		// TODO: not fully correct
		if rest := parseComment(l); rest != nil {
			return
		}
		if rest, level := parseListEl(l, '#'); rest != nil {
			p.serOl(rest, level)
			return
		}
	case '*':
		if rest, level := parseListEl(l, '*'); rest != nil {
			p.serUl(rest, level)
			return
		}
	case '•':
		if rest, level := parseListEl(l, '•'); rest != nil {
			p.serUl(rest, level)
			return
		}
	}
	return false
}

func (p *TextileParser) closePrevBlock() {
	p.closeOlIfNecessary()
	p.closeUlIfNecessary()
}

func (p *TextileParser) parseBlock(l []byte) {
	if len(l) == 0 {
		p.closePrevBlock()
		p.closeP()
		p.blockLineNo = 0
		return
	}
	p.blockLineNo += 1

	if p.parseBlockStart(l) {
		return
	}

	p.closePrevBlock()
	if p.inHtmlCode() {
		p.startNewLine()
		p.serAsHtmlCode(l)
		return
	}
	p.startNewLine()
	p.parseInline(l)
}

func dumpLines(lines [][]byte, out *bytes.Buffer) {
	for _, l := range lines {
		out.WriteString("'")
		out.Write(l)
		out.WriteString("'")
		out.Write(newline)
	}
}

func (p *TextileParser) parseRef(line []byte) bool {
	if name, url := isUrlRef(line); name != nil {
		p.refs[string(name)] = &UrlRef{name: name, url: url}
		return true
	}
	return false
}

// do a pass over all lines, extract references, remove the lines with
// references. In addition, collapses multipe consecutive empty lines
func (p *TextileParser) firstPass(lines [][]byte) [][]byte {
	res := make([][]byte, 0)
	for _, l := range lines {
		if !p.parseRef(l) {
			if len(l) > 0 {
				res = append(res, l)
			} else {
				// collapse multiple consecuitve empty lines
				if len(res) > 0 {
					last := res[len(res)-1]
					if len(last) > 0 {
						res = append(res, l)
					}
				}
			}
		}
	}
	return res
}

func (p *TextileParser) toHtml(d []byte) []byte {
	lines := splitIntoLines(d)
	if p.dumpLines {
		var buf bytes.Buffer
		fmt.Print("----------\n")
		dumpLines(lines, &buf)
		fmt.Printf("%s", string(buf.Bytes()))
	}

	lines = p.firstPass(lines)
	for _, l := range lines {
		p.parseBlock(l)
	}
	p.closeOlIfNecessary()
	p.closeUlIfNecessary()
	p.closeP()
	res := p.out.Bytes()
	return bytes.TrimRight(res, "\n")
}

func ToHtml(d []byte, dumpLines, dumpParagraphs bool) []byte {
	p := NewParser(0)
	p.dumpLines = dumpLines
	p.dumpParagraphs = dumpParagraphs
	return p.toHtml(d)
}

func ToXhtml(d []byte, dumpLines, dumpParagraphs bool) []byte {
	p := NewParser(RENDERER_XHTML)
	p.dumpLines = dumpLines
	p.dumpParagraphs = dumpParagraphs
	return p.toHtml(d)
}
