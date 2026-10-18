package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// RunNeighbors lists the shapes immediately above and below a shape in Young's lattice.
func RunNeighbors(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit neighbors SHAPE")
	}

	current, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}

	adds := shape.Adds(current)
	removes := shape.Removes(current)
	lines := []string{
		base.Line("shape", current.String()),
		base.Line("size", strconv.Itoa(current.Size())),
		base.Line("adds", strconv.Itoa(len(adds))),
	}
	for _, neighbor := range adds {
		lines = append(lines, base.Line("add", neighbor.String()))
	}
	lines = append(lines, base.Line("removes", strconv.Itoa(len(removes))))
	for _, neighbor := range removes {
		lines = append(lines, base.Line("remove", neighbor.String()))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
