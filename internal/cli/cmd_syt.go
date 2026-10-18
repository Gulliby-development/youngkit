package cli

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
	"github.com/Gulliby-development/youngkit/internal/shape"
	"github.com/Gulliby-development/youngkit/internal/tableau"
)

// RunSyt lists a bounded sequence of standard tableaux at their positions.
func RunSyt(args []string, _ base.IO) (string, error) {
	parsed, err := base.Parse(args, []base.Spec{
		{Name: "--index", Lo: 0, Hi: 9999999999, Fallback: 0},
		{Name: "--limit", Lo: 1, Hi: 50, Fallback: 10},
	})
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 1 {
		return "", base.Fail("usage: youngkit syt SHAPE [--index K] [--limit L]")
	}

	s, err := shape.Parse(parsed.Args[0])
	if err != nil {
		return "", err
	}
	if err := s.Limit(200, "syt"); err != nil {
		return "", err
	}

	total := count.SYT(s)
	index := big.NewInt(parsed.Values["--index"])
	if index.Cmp(total) >= 0 {
		return "", base.Fail("index " + index.String() + " is not below " +
			total.String() + " (the number of tableaux of " + base.Show(s.String()) + ")")
	}

	shown := int(parsed.Values["--limit"])
	remaining := new(big.Int).Sub(new(big.Int).Set(total), index)
	if remaining.Cmp(big.NewInt(int64(shown))) < 0 {
		shown = int(remaining.Int64())
	}

	lines := []string{
		base.Line("shape", s.String()),
		base.Line("size", strconv.Itoa(s.Size())),
		base.Line("count", total.String()),
		base.Line("index", index.String()),
		base.Line("shown", strconv.Itoa(shown)),
	}
	position := new(big.Int).Set(index)
	for i := 0; i < shown; i++ {
		lines = append(lines, base.Line("tableau", tableau.Unrank(s, position).String()))
		position.Add(position, big.NewInt(1))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
