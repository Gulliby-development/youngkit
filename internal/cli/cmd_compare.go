package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// RunCompare reports the dominance relation between two equal-size shapes.
func RunCompare(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 2 {
		return "", base.Fail("usage: youngkit compare A B")
	}

	a, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	b, err := shape.Parse(parsed.Args[1])
	if err != nil {
		return "", err
	}
	if a.Size() != b.Size() {
		return "", base.Fail("the sizes differ (" + strconv.Itoa(a.Size()) +
			" and " + strconv.Itoa(b.Size()) + ")")
	}

	comparison := shape.Compare(a, b)
	sumsA := base.Join(comparison.SumsA)
	if sumsA == "" {
		sumsA = "0"
	}
	sumsB := base.Join(comparison.SumsB)
	if sumsB == "" {
		sumsB = "0"
	}
	lines := []string{
		base.Line("a", a.String()),
		base.Line("b", b.String()),
		base.Line("size", strconv.Itoa(a.Size())),
		base.Line("sums-a", sumsA),
		base.Line("sums-b", sumsB),
		base.Line("relation", comparison.Relation),
	}
	if comparison.Relation == "incomparable" {
		lines = append(lines,
			base.Line("a-ahead", strconv.Itoa(comparison.AAhead)),
			base.Line("b-ahead", strconv.Itoa(comparison.BAhead)))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
