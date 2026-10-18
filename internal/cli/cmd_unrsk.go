package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/tableau"
)

// RunUnrsk recovers the permutation represented by a pair of standard tableaux.
func RunUnrsk(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 2 {
		return "", base.Fail("usage: youngkit unrsk P Q")
	}

	p, err := tableau.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	q, err := tableau.Parse(parsed.Args[1])
	if err != nil {
		return "", err
	}
	shape := p.Shape()
	if shape.String() != q.Shape().String() {
		return "", base.Fail("the shapes differ (" + shape.String() + " and " + q.Shape().String() + ")")
	}

	perm := tableau.Inverse(p, q)
	permText := base.Join(perm)
	if len(perm) == 0 {
		permText = "0"
	}
	lines := []string{
		base.Line("shape", shape.String()),
		base.Line("size", strconv.Itoa(shape.Size())),
		base.Line("perm", permText),
	}
	return strings.Join(lines, "\n") + "\n", nil
}
