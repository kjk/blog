// This code is under BSD license. See license-bsd.txt
package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"unicode/utf8"
)

// whitelisted characters valid in url
func validateRune(c rune) byte {
	if c >= 'a' && c <= 'z' {
		return byte(c)
	}
	if c >= '0' && c <= '9' {
		return byte(c)
	}
	if c == '-' || c == '_' || c == '.' {
		return byte(c)
	}
	if c == ' ' {
		return '-'
	}
	return 0
}

func charCanRepeat(c byte) bool {
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	return false
}

// urlify generates safe url from tile by removing hazardous characters
func urlify(title string) string {
	s := strings.TrimSpace(title)
	s = strings.ToLower(s)
	var res []byte
	for _, r := range s {
		c := validateRune(r)
		if c == 0 {
			continue
		}
		// eliminute duplicate consequitive characters
		var prev byte
		if len(res) > 0 {
			prev = res[len(res)-1]
		}
		if c == prev && !charCanRepeat(c) {
			continue
		}
		res = append(res, c)
	}
	s = string(res)
	if len(s) > 128 {
		s = s[:128]
	}
	return s
}

func httpErrorf(w http.ResponseWriter, format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func trimEmptyLines(a []string) []string {
	var res []string

	// remove empty lines from beginning and duplicated empty lines
	prevWasEmpty := true
	for _, s := range a {
		currIsEmpty := (len(s) == 0)
		if currIsEmpty && prevWasEmpty {
			continue
		}
		res = append(res, s)
		prevWasEmpty = currIsEmpty
	}
	// remove empty lines from end
	for len(res) > 0 {
		lastIdx := len(res) - 1
		if len(res[lastIdx]) != 0 {
			break
		}
		res = res[:lastIdx]
	}
	return res
}

func lastLineEmpty(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	lastIdx := len(lines) - 1
	line := lines[lastIdx]
	return len(line) == 0
}

func removeLastLine(lines []string) []string {
	lastIdx := len(lines) - 1
	return lines[:lastIdx]
}

func findWordEnd(s string, start int) int {
	for i := start; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			return i + 1
		}
	}
	return -1
}

// TODO: must not remove spaces from start
func collapseMultipleSpaces(s string) string {
	for {
		s2 := strings.Replace(s, "  ", " ", -1)
		if s2 == s {
			return s

		}
		s = s2
	}
}

// remove #tag from start and end
func removeHashTags(s string) (string, []string) {
	var tags []string
	defer func() {
		for i, tag := range tags {
			tags[i] = strings.ToLower(tag)
		}
	}()

	// remove hashtags from start
	for strings.HasPrefix(s, "#") {
		idx := findWordEnd(s, 0)
		if idx == -1 {
			tags = append(tags, s[1:])
			return "", tags
		}
		tags = append(tags, s[1:idx-1])
		s = strings.TrimLeft(s[idx:], " ")
	}

	// remove hashtags from end
	s = strings.TrimRight(s, " ")
	for {
		idx := strings.LastIndex(s, "#")
		if idx == -1 {
			return s, tags
		}
		// tag from the end must not have space after it
		if -1 != findWordEnd(s, idx) {
			return s, tags
		}
		// tag from the end must start at the beginning of line
		// or be proceded by space
		if idx > 0 && s[idx-1] != ' ' {
			return s, tags
		}
		tags = append(tags, s[idx+1:])
		s = strings.TrimRight(s[:idx], " ")
	}
}

// there are no guarantees in life, but this should be pretty unique string
func genRandomString() string {
	var a [20]byte
	_, err := rand.Read(a[:])
	if err == nil {
		return hex.EncodeToString(a[:])
	}
	return fmt.Sprintf("__--##%d##--__", rand.Int63())
}

func dupStringArray(a []string) []string {
	return append([]string{}, a...)
}

func reverseStringArray(a []string) {
	n := len(a) / 2
	for i := 0; i < n; i++ {
		end := len(a) - i - 1
		a[i], a[end] = a[end], a[i]
	}
}

func sanitizeForFile(s string) string {
	var res []byte
	toRemove := "/\\#()[]{},?+.'\""
	var prev rune
	buf := make([]byte, 3)
	for _, c := range s {
		if strings.ContainsRune(toRemove, c) {
			continue
		}
		switch c {
		case ' ', '_':
			c = '-'
		}
		if c == prev {
			continue
		}
		prev = c
		n := utf8.EncodeRune(buf, c)
		for i := 0; i < n; i++ {
			res = append(res, buf[i])
		}
	}
	if len(res) > 32 {
		res = res[:32]
	}
	s = string(res)
	s = strings.Trim(s, "_- ")
	s = strings.ToLower(s)
	return s
}
