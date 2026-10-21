package shape

import (
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
)

// run is a part repeated count times, with the texts it was read from.
type run struct {
	part, count         int64
	partText, countText string
}

// readRun reads one token PART or PARTxCOUNT; ok is false on a syntax error.
func readRun(token string) (run, bool) {
	partText, countText, found := strings.Cut(token, "x")
	if !found {
		countText = "1"
	}
	part, err := base.ParseInt("part", partText, 0, 9999999999)
	if err != nil {
		return run{}, false
	}
	count, err := base.ParseInt("count", countText, 0, 9999999999)
	if err != nil {
		return run{}, false
	}
	return run{part: part, count: count, partText: partText, countText: countText}, true
}
