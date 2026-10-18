package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// RunConjugate reflects a shape and renders its Ferrers diagram.
func RunConjugate(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit conjugate SHAPE")
	}

	original, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	if err := original.Limit(200, "conjugate"); err != nil {
		return "", err
	}

	conjugate := shape.Conjugate(original)
	self := "no"
	if original.String() == conjugate.String() {
		self = "yes"
	}

	lines := []string{
		base.Line("shape", original.String()),
		base.Line("size", strconv.Itoa(original.Size())),
		base.Line("conjugate", conjugate.String()),
		base.Line("self", self),
	}
	for _, row := range shape.Diagram(original) {
		lines = append(lines, base.Line("row", row))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
