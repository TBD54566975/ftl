// Package strcase provides programming case conversion functions for strings.
//
// These case conversion functions are used to deterministically convert strings
// to various programming cases.
package strcase

// NOTE: This code is from https://github.com/fatih/camelcase. MIT license.

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func title(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + strings.ToLower(s[n:])
}

func ToLowerCamel(s string) string {
	parts := split(s)
	for i := range parts {
		parts[i] = title(parts[i])
	}
	return strings.ToLower(parts[0]) + strings.Join(parts[1:], "")
}

func ToUpperCamel(s string) string {
	parts := split(s)
	for i := range parts {
		parts[i] = title(parts[i])
	}
	return strings.Join(parts, "")
}

func ToLowerSnake(s string) string {
	parts := split(s)
	out := make([]string, 0, len(parts)*2)
	for i := range parts {
		if parts[i] == "_" {
			continue
		}
		out = append(out, strings.ToLower(parts[i]))
	}
	return strings.Join(out, "_")
}

func ToUpperSnake(s string) string {
	parts := split(s)
	out := make([]string, 0, len(parts)*2)
	for i := range parts {
		if parts[i] == "_" {
			continue
		}
		out = append(out, strings.ToUpper(parts[i]))
	}
	return strings.Join(out, "_")
}

func ToLowerKebab(s string) string {
	parts := split(s)
	out := make([]string, 0, len(parts)*2)
	for i := range parts {
		if parts[i] == "-" || parts[i] == "_" {
			continue
		}
		out = append(out, strings.ToLower(parts[i]))
	}
	return strings.Join(out, "-")
}

func ToUpperKebab(s string) string {
	parts := split(s)
	out := make([]string, 0, len(parts)*2)
	for i := range parts {
		if parts[i] == "-" || parts[i] == "_" {
			continue
		}
		out = append(out, strings.ToUpper(parts[i]))
	}
	return strings.Join(out, "-")
}

// Splits a camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//	"" =>                     [""]
//	"lowercase" =>            ["lowercase"]
//	"Class" =>                ["Class"]
//	"MyClass" =>              ["My", "Class"]
//	"MyC" =>                  ["My", "C"]
//	"HTML" =>                 ["HTML"]
//	"PDFLoader" =>            ["PDF", "Loader"]
//	"AString" =>              ["A", "String"]
//	"SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//	"vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//	"GL11Version" =>          ["GL", "11", "Version"]
//	"99Bottles" =>            ["99", "Bottles"]
//	"May5" =>                 ["May", "5"]
//	"BFG9000" =>              ["BFG", "9000"]
//	"BöseÜberraschung" =>     ["Böse", "Überraschung"]
//	"Two  spaces" =>          ["Two", "  ", "spaces"]
//	"BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1. If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2. Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3. Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4. Iterate through array of split strings, and if a given string
//     is upper case:
//     if subsequent string is lower case:
//     move last character of upper case string to beginning of
//     lower case string
func split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		var class int
		switch {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := range len(runes) - 1 {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return entries
}
