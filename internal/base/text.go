package base

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Show quotes text, replacing non-printable-ASCII runes and invalid UTF-8 bytes.
func Show(text string) string {
	var shown strings.Builder
	shown.Grow(len(text) + 2)
	shown.WriteByte('"')
	for _, r := range text {
		if r >= '!' && r <= '~' {
			shown.WriteRune(r)
		} else {
			shown.WriteByte('?')
		}
	}
	shown.WriteByte('"')
	return shown.String()
}

// Line formats a key padded to ten Unicode code points, two spaces, and value.
func Line(key, value string) string {
	padding := 10 - utf8.RuneCountInString(key)
	if padding < 0 {
		padding = 0
	}
	return key + strings.Repeat(" ", padding) + "  " + value
}

// Join returns the decimal values separated by single spaces.
func Join(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	var joined strings.Builder
	for i, num := range nums {
		if i > 0 {
			joined.WriteByte(' ')
		}
		joined.WriteString(strconv.Itoa(num))
	}
	return joined.String()
}
