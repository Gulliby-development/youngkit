package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// RunRank reports a shape's position among the partitions of its size.
func RunRank(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit rank SHAPE")
	}

	partition, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	if err := partition.Limit(400, "rank"); err != nil {
		return "", err
	}

	rank, total := count.Rank(partition)
	lines := []string{
		base.Line("shape", partition.String()),
		base.Line("size", strconv.Itoa(partition.Size())),
		base.Line("rank", rank.String()),
		base.Line("of", total.String()),
	}
	return strings.Join(lines, "\n") + "\n", nil
}
