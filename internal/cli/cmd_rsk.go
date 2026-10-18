package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/tableau"
)

// RunRsk computes the Robinson-Schensted pair of a permutation.
func RunRsk(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	words, err := base.Words(io, parsed.Args)
	if err != nil {
		return "", err
	}
	if len(words) == 0 {
		return "", base.Fail("usage: youngkit rsk ENTRY... | youngkit rsk -")
	}
	if len(words) > 5000 {
		return "", base.Fail("too many entries (" + strconv.Itoa(len(words)) + " is above 5000)")
	}

	perm := make([]int, len(words))
	seen := make([]bool, len(words)+1)
	for index, word := range words {
		entry, err := base.ParseInt("entry", word, 1, int64(len(words)))
		if err != nil {
			return "", err
		}
		if seen[entry] {
			return "", base.Fail("entry " + strconv.FormatInt(entry, 10) + " is repeated")
		}
		seen[entry] = true
		perm[index] = int(entry)
	}

	p, q := tableau.RSK(perm)
	lis := 0
	if len(p) > 0 {
		lis = len(p[0])
	}
	lines := []string{
		base.Line("perm", base.Join(perm)),
		base.Line("shape", p.Shape().String()),
		base.Line("lis", strconv.Itoa(lis)),
		base.Line("lds", strconv.Itoa(len(p))),
	}
	for _, row := range p {
		lines = append(lines, base.Line("p", base.Join(row)))
	}
	for _, row := range q {
		lines = append(lines, base.Line("q", base.Join(row)))
	}
	return strings.Join(lines, "\n") + "\n", nil
}
