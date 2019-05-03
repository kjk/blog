package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlify(t *testing.T) {
	tests := []struct {
		s    string
		sExp string
	}{
		{
			s:    "Laws of marketing #22 (resources) ",
			sExp: "laws-of-marketing-22-resources",
		},
		{
			s:    "t  -_",
			sExp: "t-_",
		},
		{
			s:    "foo.htML  ",
			sExp: "foo.html",
		},
	}
	for _, test := range tests {
		sGot := urlify(test.s)
		assert.Equal(t, test.sExp, sGot)
	}
}

func TestTrimEmptyLines(t *testing.T) {
	tests := []struct {
		a   []string
		exp []string
	}{
		{
			a:   []string{"a"},
			exp: []string{"a"},
		},
		{
			a:   []string{"a", "", "", "b"},
			exp: []string{"a", "", "b"},
		},
		{
			a:   []string{"", "a", ""},
			exp: []string{"a"},
		},
		{
			a:   []string{"", "", "a", "", "b", "", ""},
			exp: []string{"a", "", "b"},
		},
	}
	for _, test := range tests {
		got := trimEmptyLines(test.a)
		assert.Equal(t, test.exp, got)
	}
}

func TestRemoveHashtags(t *testing.T) {
	tests := []struct {
		s    string
		tags []string
		sExp string
	}{
		{
			s:    "#idea Build a web service  ",
			sExp: "Build a web service",
			tags: []string{"idea"},
		},
		{
			s:    "#foo   #BAr and #me",
			sExp: "and",
			tags: []string{"foo", "bar", "me"},
		},
		{
			s:    "not #found here",
			sExp: "not #found here",
			tags: nil,
		},
		{
			s:    "#foo   not a#hash",
			sExp: "not a#hash",
			tags: []string{"foo"},
		},
	}
	for _, test := range tests {
		sGot, tags := removeHashTags(test.s)
		assert.Equal(t, test.sExp, sGot)
		assert.Equal(t, test.tags, tags)
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{
			s:   "foo",
			exp: "Foo",
		},
		{
			s:   "FOO",
			exp: "Foo",
		},
		{
			s:   "FOO baR",
			exp: "Foo bar",
		},
	}
	for _, test := range tests {
		got := capitalize(test.s)
		assert.Equal(t, test.exp, got)
	}
}
