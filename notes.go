package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/kjk/u"
	"github.com/rs/xid"
	"github.com/sourcegraph/syntaxhighlight"
	"mvdan.cc/xurls"
)

const (
	noteSeparator  = "---"
	codeBlockStart = "```"
)

var (
	notesDays       []*notesForDay
	notesTagToNotes map[string][]*note
	// maps unique id of the note (from Id: ${id} metadata) to the note
	notesIDToNote  map[string]*note
	notesTagCounts []tagWithCount
	notesAllNotes  []*note

	notesWeekStartDayToNotes map[string][]*note
	notesWeekStarts          []string
	nTotalNotes              int
)

type tagWithCount struct {
	Tag   string
	Count int
}

type noteMetadata struct {
	ID    string
	Title string
}

type note struct {
	Day            time.Time
	DayStr         string // in format "2006-01-02"
	DayWithNameStr string // in format "2006-01-02 Mon"
	ID             string
	Title          string
	URL            string // in format /dailynotes/note/${id}-${title}
	HTMLBody       template.HTML
	Tags           []string
	rawLines       []string
}

type notesForDay struct {
	Day    time.Time
	DayStr string
	Notes  []*note
	lines  []string
}

type modelNotesForWeek struct {
	Notes         []*note
	TotalNotes    int
	TagCounts     []tagWithCount
	WeekStartDay  string
	NextWeek      string
	PrevWeek      string
	AnalyticsCode string
}

func buildBodyFromLines(lines []string) (string, []string) {
	var resTags []string

	var tags []string
	for i, line := range lines {
		line, tags = removeHashTags(line)
		lines[i] = line
		resTags = append(resTags, tags...)
	}
	resTags = u.RemoveDuplicateStrings(resTags)

	// collapse multiple empty lines into single empty line
	// and remove lines that are just #hashtags
	currWrite := 1
	for i := 1; i < len(lines); i++ {
		prev := lines[currWrite-1]
		curr := lines[i]
		if len(prev) == 0 && len(curr) == 0 {
			// skips the current line because we don't advance currWrite
			continue
		}

		lines[currWrite] = curr
		currWrite++
	}
	lines = lines[:currWrite]

	if len(lines) == 0 {
		return "", resTags
	}

	// remove empty lines from beginning
	for len(lines[0]) == 0 {
		lines = lines[1:]
	}

	// remove empty lines from end
	for lastLineEmpty(lines) {
		lines = removeLastLine(lines)
	}
	return strings.Join(lines, "\n"), resTags
}

// given lines, extracts metadata information from lines that are:
// Id: $id
// Title: $title
// Returns new lines with metadata lines removed
func extractMetaDataFromLines(lines []string) ([]string, noteMetadata) {
	var res noteMetadata
	writeIdx := 0
	for i, s := range lines {
		idx := strings.Index(s, ":")
		skipLine := false
		if -1 != idx {
			name := strings.ToLower(s[:idx])
			val := strings.TrimSpace(s[idx+1:])
			switch name {
			case "id":
				res.ID = val
				skipLine = true
			case "title":
				res.Title = val
				skipLine = true
			}
		}
		if skipLine || writeIdx == i {
			continue
		}
		lines[writeIdx] = lines[i]
		writeIdx++
	}
	//panicIf(res.ID == "", "note has no Id:. Note: %s\n", strings.Join(lines, "\n"))
	return lines[:writeIdx], res
}

func noteToHTML(s string) string {
	urls := xurls.Relaxed.FindAllString(s, -1)
	urls = u.RemoveDuplicateStrings(urls)

	// sort by length, longest first, so that we correctly convert
	// urls to hrefs when there are 2 urls like http://foo.com
	// and http://foo.com/longer
	sort.Slice(urls, func(i, j int) bool {
		return len(urls[i]) > len(urls[j])
	})
	// this is a two-step url -> random_unique_string,
	// random_unique_string -> url replacement to prevent
	// double-escaping if we have 2 urls like: foo.bar.com and bar.com
	urlToAnchor := make(map[string]string)

	for _, url := range urls {
		anchor := genRandomString()
		urlToAnchor[url] = anchor
		s = strings.Replace(s, url, anchor, -1)
	}

	for _, url := range urls {
		replacement := fmt.Sprintf(`<a href="%s">%s</a>`, url, url)
		anchor := urlToAnchor[url]
		s = strings.Replace(s, anchor, replacement, -1)
	}
	//fmt.Printf("%s\n", s)
	s, _ = sanitize.HTMLAllowing(s)
	//panicIfErr(err)
	//fmt.Printf("%s\n\n\n", s)
	return s
}

type codeSnippetInfo struct {
	anchor   string
	codeHTML []byte
}

// returns new lines and a mapping of string => html as flattened string array
func extractCodeSnippets(lines []string) ([]string, []*codeSnippetInfo) {
	var resLines []string
	var codeSnippets []*codeSnippetInfo
	codeLineStart := -1
	for i, s := range lines {
		isCodeLine := strings.HasPrefix(s, codeBlockStart)
		if isCodeLine {
			if codeLineStart == -1 {
				// this is a beginning of new code block
				codeLineStart = i
			} else {
				// end of the code block
				//lang := strings.TrimPrefix(lines[codeLineStart], codeBlockStart)
				codeLines := lines[codeLineStart+1 : i]
				codeLineStart = -1
				code := strings.Join(codeLines, "\n")
				codeHTML, err := syntaxhighlight.AsHTML([]byte(code))
				panicIfErr(err)
				anchor := genRandomString()
				resLines = append(resLines, anchor)
				snippetInfo := &codeSnippetInfo{
					anchor:   anchor,
					codeHTML: codeHTML,
				}
				codeSnippets = append(codeSnippets, snippetInfo)
			}
		} else {
			if codeLineStart == -1 {
				resLines = append(resLines, s)
			}
		}
	}
	// TODO: could append unclosed lines
	panicIf(codeLineStart != -1)

	return resLines, codeSnippets
}

func newNote(lines []string) *note {
	nTotalNotes++
	rawLines := dupStringArray(lines)
	lines, meta := extractMetaDataFromLines(lines)
	lines, codeSnippets := extractCodeSnippets(lines)

	s, tags := buildBodyFromLines(lines)

	body := noteToHTML(s)
	for _, codeSnippet := range codeSnippets {
		anchor := codeSnippet.anchor
		codeHTML := `<pre class="note-code">` + string(codeSnippet.codeHTML) + `</pre>`
		body = strings.Replace(body, anchor, codeHTML, -1)
	}
	return &note{
		Tags:     tags,
		HTMLBody: template.HTML(body),
		ID:       meta.ID,
		Title:    meta.Title,
		rawLines: rawLines,
	}
}

func linesToNotes(lines []string) []*note {
	// parts are separated by "---" line
	var res []*note
	var curr []string
	for _, line := range lines {
		if line == noteSeparator {
			if len(curr) > 0 {
				part := newNote(curr)
				res = append(res, part)
			}
			curr = nil
		} else {
			curr = append(curr, line)
		}
	}
	if len(curr) > 0 {
		part := newNote(curr)
		res = append(res, part)
	}
	return res
}

func readNotesFoDay(path string) ([]*notesForDay, error) {
	seenDays := make(map[string]bool)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var notes []*notesForDay
	var curr *notesForDay
	var lines []string

	for scanner.Scan() {
		s := strings.TrimRight(scanner.Text(), "\n\r\t ")
		day, err := time.Parse("2006-01-02", s)

		if err == nil {
			// this is a new day
			dayStr := s
			panicIf(seenDays[dayStr], "duplicate day: %s", dayStr)
			seenDays[dayStr] = true

			if curr != nil {
				curr.lines = lines
				curr.Notes = linesToNotes(lines)
				notes = append(notes, curr)
				if false && len(notes) == 1 {
					s = strings.Join(lines, "\n")
					fmt.Printf("First day:\n%s\n", s)
					s = strings.Join(curr.Notes[0].rawLines, "\n")
					fmt.Printf("First note:\n%s\n", s)
				}
			}
			curr = &notesForDay{
				DayStr: dayStr,
				Day:    day,
			}
			lines = nil
		} else {
			lines = append(lines, s)
		}
	}
	curr.lines = lines
	curr.Notes = linesToNotes(lines)
	notes = append(notes, curr)
	return notes, scanner.Err()
}

func notesGenIDIfNecessary() {
	path := filepath.Join("articles", "notes.txt")
	notesPerDay, err := readNotesFoDay(path)
	panicIfErr(err)
	var lines []string
	var updatedNotes []*note
	for _, dayNotes := range notesPerDay {
		lines = append(lines, dayNotes.DayStr)
		notes := dayNotes.Notes
		lastNoteIdx := len(notes) - 1
		for idx, note := range notes {
			if note.ID == "" {
				note.ID = xid.New().String()
				fmt.Printf("Generated id %s for note from %s\n", note.ID, dayNotes.DayStr)
				idLine := fmt.Sprintf("Id: %s", note.ID)
				lines = append(lines, idLine)
				updatedNotes = append(updatedNotes, note)
			}
			rawLines := trimEmptyLines(note.rawLines)
			lines = append(lines, rawLines...)
			if idx != lastNoteIdx {
				lines = append(lines, "", noteSeparator, "")
			}
		}
		lines = append(lines, "")
	}

	if len(updatedNotes) > 0 {
		s := strings.Join(lines, "\n")
		err := ioutil.WriteFile(path, []byte(s), 0644)
		panicIfErr(err)
		fmt.Printf("Generated id for %d notes\n", len(updatedNotes))
		fmt.Printf("Need to checkin.\n")
		os.Exit(0)
	}
}

// given time, return time on start of week (monday)
func calcWeekStart(t time.Time) time.Time {
	// wd is 1 to 7
	wd := t.Weekday()
	dayOffset := time.Duration((wd - 1)) * time.Hour * -24
	return t.Add(dayOffset)
}

func readNotes(path string) error {
	var err error
	if len(notesAllNotes) > 0 {
		return nil
	}

	notesDays = nil
	notesTagToNotes = make(map[string][]*note)
	notesIDToNote = make(map[string]*note)
	notesTagCounts = nil
	notesAllNotes = nil
	notesWeekStartDayToNotes = make(map[string][]*note)
	notesWeekStarts = nil

	notesDays, err = readNotesFoDay(path)
	if err != nil {
		return err
	}

	// verify they are in chronological order
	for i := 1; i < len(notesDays); i++ {
		notesForDay := notesDays[i-1]
		notesForPrevDay := notesDays[i]
		diff := notesForDay.Day.Sub(notesForPrevDay.Day)
		if diff < 0 {
			return fmt.Errorf("Note '%s' should be later than '%s'", notesForDay.DayStr, notesForPrevDay.DayStr)
		}
	}

	nNotes := 0
	// update date and id on notes
	for _, day := range notesDays {
		weekStartTime := calcWeekStart(day.Day)
		weekStartDay := weekStartTime.Format("2006-01-02")
		for _, note := range day.Notes {
			notesAllNotes = append(notesAllNotes, note)
			nNotes++
			id := note.ID
			panicIf(notesIDToNote[id] != nil, "duplicate note id: %s", id)
			notesIDToNote[id] = note
			note.Day = day.Day
			note.DayStr = day.Day.Format("2006-01-02")
			note.DayWithNameStr = day.Day.Format("2006-01-02 Mon")
			note.URL = "/dailynotes/note/" + id
			if note.Title != "" {
				note.URL += "-" + urlify(note.Title)
			}
			for _, tag := range note.Tags {
				a := notesTagToNotes[tag]
				a = append(a, note)
				notesTagToNotes[tag] = a
			}
			a := notesWeekStartDayToNotes[weekStartDay]
			a = append(a, note)
			notesWeekStartDayToNotes[weekStartDay] = a
		}
	}
	for day := range notesWeekStartDayToNotes {
		notesWeekStarts = append(notesWeekStarts, day)
	}
	var tags []string
	for tag := range notesTagToNotes {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		count := len(notesTagToNotes[tag])
		tc := tagWithCount{
			Tag:   tag,
			Count: count,
		}
		notesTagCounts = append(notesTagCounts, tc)
	}

	sort.Strings(notesWeekStarts)
	reverseStringArray(notesWeekStarts)
	fmt.Printf("Read %d notes in %d days and %d weeks\n", nNotes, len(notesDays), len(notesWeekStarts))
	return nil
}
