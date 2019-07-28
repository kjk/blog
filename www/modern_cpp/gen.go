package main

import (
	"fmt"
	"strings"
)

const c17lang = `
template argument deduction for class templates
declaring non-type template parameters with auto
folding expressions
new rules for auto deduction from braced-init-list
constexpr lambda
lambda capture this by value
inline variables
nested namespaces
structured bindings
selection statements with initializer
constexpr if
utf-8 character literals
direct-list-initialization of enums
`

const c17lib = `
std::variant
std::optional
std::any
std::string_view
std::invoke
std::apply
splicing for maps and sets
`

const c14lang = `
binary literals
generic lambda expressions
lambda capture initializers
return type deduction
decltype(auto)
relaxing constraints on constexpr functions
variable templates
`

const c14lib = `
user-defined literals for standard library types
compile-time integer sequences
std::make_unique
`

const c11lang = `
move semantics
variadic templates
rvalue references
initializer lists
static assertions
auto
lambda expressions
decltype
template aliases
nullptr
strongly-typed enums
attributes
constexpr
delegating constructors
user-defined literals
explicit virtual overrides
final specifier
default functions
deleted functions
range-based for loops
special member functions for move semantics
converting constructors
explicit conversion functions
inline-namespaces
non-static data member initializers
right angle brackets
`

const c11lib = `
std::move
std::forward
std::to_string
type traits
smart pointers
std::chrono
tuples
std::tie
std::array
unordered containers
std::make_shared
memory model
`

func toLines(s string) []string {
	lines := strings.Split(s, "\n")
	var res []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			res = append(res, line)
		}
	}
	return res
}

func dumpLines(lines []string) {
	fmt.Printf("%d lines:\n", len(lines))
	for _, line := range lines {
		fmt.Printf("%s\n", line)
	}
}

func a(text string, id string) string {
	format := `<a href="#" onclick='return g("%s")'>%s</a>`
	return fmt.Sprintf(format, id, text)
}

func li(inner string) string {
	return `<li>` + inner + `</li>` + "\n"
}

func ul(inner string) string {
	return "<ul>\n" + inner + "</ul>"
}

func genUL(lines []string, idPrefix string) string {
	s := ""
	for id, line := range lines {
		idStr := fmt.Sprintf("%s%d", idPrefix, id+1)
		s += li(a(line, idStr))
	}
	return ul(s)
}

func dumpAsLines(s string, idPrefix string) {
	lines := toLines(s)
	// dumpLines(lines)
	s = genUL(lines, idPrefix) + "\n\n"
	fmt.Printf(s)
}

func calcLines() {
	strings := []string{c17lang, c17lib, c14lang, c14lib, c11lang, c11lib}
	nLines := 0
	for _, s := range strings {
		lines := toLines(s)
		nLines += len(lines)
		nLines++ // for header
	}
	fmt.Printf("%d lines\n", nLines)
	for _, nColumns := range []int{2, 3, 4} {
		nPerColumn := nLines / nColumns
		if nLines%nColumns > 0 {
			nPerColumn++
		}
		fmt.Printf("%d per column if %d columsn\n", nPerColumn, nColumns)
	}
}

func dump() {
	dumpAsLines(c17lang, "c17lang-")
	dumpAsLines(c17lib, "c17lib-")

	dumpAsLines(c14lang, "c14lang-")
	dumpAsLines(c14lib, "c14lib-")

	dumpAsLines(c11lang, "c11lang-")
	dumpAsLines(c11lib, "c11lib-")
}

func main() {
	//dump()
	calcLines()
}
