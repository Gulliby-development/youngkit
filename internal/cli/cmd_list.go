package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
)

// RunList lists a bounded prefix of the partitions of n.
func RunList(args []string, io base.IO) (string, error) {
	specs := []base.Spec{
		{Name: "--max-part", Lo: 1, Hi: 200, Fallback: 0},
		{Name: "--limit", Lo: 1, Hi: 1000, Fallback: 20},
	}
	parsed, err := base.Parse(args, specs)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit list N [--max-part K] [--limit L]")
	}

	n, err := base.ParseInt("n", parsed.Args[0], 0, 200)
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
	limit := int(parsed.Values["--limit"])
	total := count.Partitions(int(n), int(maxPart), false)
	partitions := count.Generate(int(n), int(maxPart), limit)

	lines := []string{
		base.Line("n", strconv.FormatInt(n, 10)),
		base.Line("max-part", displayedMaxPart),
		base.Line("count", total.String()),
		base.Line("shown", strconv.Itoa(len(partitions))),
	}
	for _, partition := range partitions {
		lines = append(lines, base.Line("part", partition.String()))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
