package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/shape"
	"github.com/Gulliby-development/youngkit/internal/tableau"
)

// RunCore reports the K-core and K-quotient of a shape.
func RunCore(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 2 {
		return "", base.Fail("usage: youngkit core K SHAPE")
	}

	k, err := base.ParseInt("K", parsed.Args[0], 2, 50)
	if err != nil {
		return "", err
	}
	original, err := shape.Parse(parsed.Args[1])
	if err != nil {
		return "", err
	}

	core, quotient := tableau.Abacus(original, int(k))
	weight := 0
	lines := []string{
		base.Line("shape", original.String()),
		base.Line("k", strconv.Itoa(int(k))),
		base.Line("core", core.String()),
	}
	for _, part := range quotient {
		weight += part.Size()
	}
	lines = append(lines, base.Line("weight", strconv.Itoa(weight)))
	for runner, part := range quotient {
		lines = append(lines, base.Line("quotient", strconv.Itoa(runner)+"="+part.String()))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
