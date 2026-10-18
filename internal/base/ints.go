package base

import (
	"strconv"
)

// ParseInt parses a canonical non-negative decimal integer within inclusive bounds.
func ParseInt(name, text string, lo, hi int64) (int64, error) {
	if !canonicalInteger(text) {
		return 0, invalidInteger(name, text, lo, hi)
	}

	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil || value < lo || value > hi {
		return 0, invalidInteger(name, text, lo, hi)
	}
	return value, nil
}

func canonicalInteger(text string) bool {
	if text == "0" {
		return true
	}
	if len(text) == 0 || len(text) > 10 || text[0] < '1' || text[0] > '9' {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func invalidInteger(name, text string, lo, hi int64) error {
	return Fail("invalid " + name + " " + Show(text) +
		" (expected an integer from " + strconv.FormatInt(lo, 10) +
		" to " + strconv.FormatInt(hi, 10) + ")")
}
