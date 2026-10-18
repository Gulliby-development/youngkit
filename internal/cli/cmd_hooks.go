package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// RunHooks renders hook lengths and exact tableau counts for a shape.
func RunHooks(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, []base.Spec{
		{Name: "--entries", Lo: 1, Hi: 1000},
	})
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit hooks SHAPE [--entries K]")
	}

	original, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	if err := original.Limit(2000, "hooks"); err != nil {
		return "", err
	}

	hooks := count.Hooks(original)
	lines := []string{
		base.Line("shape", original.String()),
		base.Line("size", strconv.Itoa(original.Size())),
	}
	for _, row := range hooks {
		lines = append(lines, base.Line("hook", base.Join(row)))
	}
	lines = append(lines, base.Line("syt", count.SYT(original).String()))
	if entries := parsed.Values["--entries"]; entries != 0 {
		lines = append(lines,
			base.Line("entries", strconv.FormatInt(entries, 10)),
			base.Line("ssyt", count.SSYT(original, int(entries)).String()),
		)
	}
	return strings.Join(lines, "\n") + "\n", nil
}
