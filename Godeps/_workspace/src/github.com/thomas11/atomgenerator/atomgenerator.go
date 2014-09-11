// Package atomgenerator generates Atom feeds.
//
// The package is based on an implementation from Krzysztof Kowalczyk's
// https://github.com/kjk/apptranslator, with some modifications:
//
// - Generate entry ids based on a scheme described on diveintomark.org,
//   see `(e Entry) genId()`.
// - Added <author>s to Feed and Entry.
// - Added <content> field to Entry.
// - Validate() to check whether the Feed conforms to Atom.
// - Godoc.
//
// http://www.atomenabled.org/developers/syndication and RFC 4287 were
// used as a references.
//
// This code is under BSD license. See license-bsd.txt.
//
// Authors:
// - Krzysztof Kowalczyk, http://blog.kowalczyk.info/
// - Thomas Kappler, http://www.thomaskappler.net/
package atomgenerator

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const ns = "http://www.w3.org/2005/Atom"

// An Atom feed. Make an instance of this struct, add entries using
// AddEntry(), and generate your feed with GenXml().
type Feed struct {
	// Required.
	Title string
	// Required.
	PubDate time.Time
	Link    string
	// Required unless all entries have at least one Author.
	authors []Author
	entries []*Entry
}

func (f *Feed) AddEntry(e *Entry) {
	f.entries = append(f.entries, e)
}

func (f *Feed) AddAuthor(author Author) {
	f.authors = append(f.authors, author)
}

type Author struct {
	// Required.
	Name string `xml:"name"`
	// Optional.
	Email string `xml:"email,omitempty"`
	// Optional.
	Uri string `xml:"uri,omitempty"`
}

type Category struct {
	// Required
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Label  string `xml:"label,attr,omitempty"`
}

type Entry struct {
	// Required.
	Title string
	// Required.
	PubDate     time.Time
	Link        string
	Description string
	Content     string
	// Required unless the Feed has at least one Author.
	authors    []Author
	categories []Category
}

func (e *Entry) AddAuthor(author Author) {
	e.authors = append(e.authors, author)
}

func (e *Entry) AddCategory(cat Category) {
	e.categories = append(e.categories, cat)
}

type typedTag struct {
	S    string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

type entryXml struct {
	XMLName    xml.Name `xml:"entry"`
	Title      string   `xml:"title"`
	Link       *linkXml
	Updated    string     `xml:"updated"`
	Id         string     `xml:"id"`
	Summary    *typedTag  `xml:"summary"`
	Content    *typedTag  `xml:"content"`
	Authors    []Author   `xml:"author"`
	Categories []Category `xml:"category,omitempty"`
}

type linkXml struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
}

type feedXml struct {
	XMLName xml.Name `xml:"feed"`
	Ns      string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    *linkXml
	Id      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Authors []Author `xml:"author"`
	Entries []*entryXml
}

// Generate a unique global id for the Entry using the scheme described in
// http://web.archive.org/web/20110915110202/http://diveintomark.org/archives/2004/05/28/howto-atom-id.
func (e *Entry) genId() string {
	u, err := url.Parse(e.Link)
	if err != nil {
		return e.Link
	}

	var b bytes.Buffer
	b.WriteString("tag:")
	b.WriteString(u.Host)
	b.WriteString(",")
	b.WriteString(e.PubDate.Format("2006-01-02"))
	b.WriteString(":")
	b.WriteString(u.Path)
	if len(u.Fragment) > 0 {
		if !strings.HasSuffix(u.Path, "/") {
			b.WriteString("/")
		}
		b.WriteString(strings.Replace(u.Fragment, "#", "/", -1))
	}

	return b.String()
}

func newEntryXml(e *Entry) *entryXml {
	x := &entryXml{
		Id:         e.genId(),
		Title:      e.Title,
		Link:       &linkXml{Href: e.Link, Rel: "alternate"},
		Updated:    e.PubDate.Format(time.RFC3339),
		Categories: e.categories,
		Authors:    e.authors,
	}

	if len(e.Description) > 0 {
		x.Summary = &typedTag{e.Description, "html"}
	}
	if len(e.Content) > 0 {
		x.Content = &typedTag{e.Content, "html"}
	}

	return x
}

// Generate the final Atom feed in XML.
func (f *Feed) GenXml() ([]byte, error) {
	feed := &feedXml{
		Ns:      ns,
		Title:   f.Title,
		Authors: f.authors,
		Link:    &linkXml{Href: f.Link, Rel: "alternate"},
		Id:      f.Link,
		Updated: f.PubDate.Format(time.RFC3339)}
	for _, e := range f.entries {
		feed.Entries = append(feed.Entries, newEntryXml(e))
	}
	data, err := xml.MarshalIndent(feed, " ", " ")
	if err != nil {
		return []byte{}, err
	}
	s := append([]byte(xml.Header[:len(xml.Header)-1]), data...)
	return s, nil
}

// Check if the feed conforms to the Atom standard. The check is fairly ok,
// but not guaranteed to be comprehensive! Returns the list of all problems
// found. If it's empty, the feed was found to be valid.
func (f *Feed) Validate() []error {
	errs := make([]error, 0, 5)

	// Feed must have title, updated. Id is generated.
	if len(f.Title) == 0 {
		errs = append(errs, errors.New("Feed must have a Title."))
	}
	if f.PubDate.IsZero() {
		errs = append(errs, errors.New("Feed must have a PubDate."))
	}

	// Either the feed has an author, or all entries must have one.
	if len(f.authors) == 0 {
		for _, e := range f.entries {
			if len(e.authors) == 0 {
				errs = append(errs, fmt.Errorf(
					"Feed has no authors, and entry %v has none either.", e.Title))
			}
		}
	} else {
		// All authors must have a name.
		for i, author := range f.authors {
			if len(author.Name) == 0 {
				errs = append(errs, fmt.Errorf(
					"Feed author %v must have a Name.", i))
			}
		}
	}

	// Entries must have title, updated. Id is generated.
	for i, e := range f.entries {
		if len(e.Title) == 0 {
			errs = append(errs, fmt.Errorf("Entry %v must have a Title.", i))
		}
		if e.PubDate.IsZero() {
			errs = append(errs, fmt.Errorf("Entry %v must have a PubDate.", i))
		}
	}

	// Entry categories must have the term attribute.
	for i, e := range f.entries {
		for j, cat := range e.categories {
			if len(cat.Term) == 0 {
				errs = append(errs, fmt.Errorf("Category %v of entry %v must have a Term.", j, i))
			}
		}
	}

	return errs
}
