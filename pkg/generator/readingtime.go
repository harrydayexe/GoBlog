package generator

import (
	"math"
	"regexp"
	"strings"
)

const wordsPerMinute = 220

var tagRE = regexp.MustCompile(`<[^>]*>`)

// wordCount strips HTML tags from rendered content and counts whitespace-separated tokens.
// Tags are replaced with a space so adjacent elements don't merge their words.
func wordCount(html []byte) int {
	stripped := tagRE.ReplaceAll(html, []byte(" "))
	return len(strings.Fields(string(stripped)))
}

// minutesFromWords converts a word count to an estimated reading time in minutes.
// Uses a ceiling division at wordsPerMinute WPM with a 1-minute floor.
func minutesFromWords(n int) int {
	if n <= 0 {
		return 1
	}
	m := int(math.Ceil(float64(n) / float64(wordsPerMinute)))
	if m < 1 {
		return 1
	}
	return m
}
