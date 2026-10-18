package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
)

// RunCount reports the exact number of partitions of n.
func RunCount(args []string, io base.IO) (string, error) {
	specs := []base.Spec{
		{Name: "--max-part", Lo: 1, Hi: 5000, Fallback: 0},
		{Name: "--distinct", Flag: true},
	}
	parsed, err := base.Parse(args, specs)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit count N [--max-part K] [--distinct]")
	}

	n, err := base.ParseInt("n", parsed.Args[0], 0, 5000)
	if err != nil {
		return "", err
	}

	displayedMaxPart := "any"
	maxPart := parsed.Values["--max-part"]
	if maxPart != 0 {
		displayedMaxPart = strconv.FormatInt(maxPart, 10)
	} else {
		maxPart = n
	}
	distinct := parsed.Values["--distinct"] != 0
	distinctText := "no"
	if distinct {
		distinctText = "yes"
	}

	total := count.Partitions(int(n), int(maxPart), distinct)
	lines := []string{
		base.Line("n", strconv.FormatInt(n, 10)),
		base.Line("max-part", displayedMaxPart),
		base.Line("distinct", distinctText),
		base.Line("count", total.String()),
	}
	return strings.Join(lines, "\n") + "\n", nil
}
