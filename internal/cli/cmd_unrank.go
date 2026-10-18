package cli

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/count"
)

// RunUnrank returns the partition at a zero-based reverse-lexicographic rank.
func RunUnrank(args []string, io base.IO) (string, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return "", err
	}
	if len(parsed.Args) != 2 {
		return "", base.Fail("usage: youngkit unrank N RANK")
	}

	n, err := base.ParseInt("n", parsed.Args[0], 0, 400)
	if err != nil {
		return "", err
	}
	total := count.Partitions(int(n), int(n), false)
	maximum := new(big.Int).Sub(new(big.Int).Set(total), big.NewInt(1))

	rankText := parsed.Args[1]
	if !canonicalRank(rankText) {
		return "", invalidRank(rankText, maximum)
	}
	rank, ok := new(big.Int).SetString(rankText, 10)
	if !ok || rank.Cmp(total) >= 0 {
		return "", invalidRank(rankText, maximum)
	}

	partition := count.Unrank(int(n), new(big.Int).Set(rank))
	lines := []string{
		base.Line("n", strconv.FormatInt(n, 10)),
		base.Line("rank", rank.String()),
		base.Line("shape", partition.String()),
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func canonicalRank(text string) bool {
	if text == "0" {
		return true
	}
	if len(text) == 0 || len(text) > 200 || text[0] < '1' || text[0] > '9' {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func invalidRank(text string, maximum *big.Int) error {
	return base.Fail("invalid rank " + base.Show(text) +
		" (expected an integer from 0 to " + maximum.String() + ")")
}
