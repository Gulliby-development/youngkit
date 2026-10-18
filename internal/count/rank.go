package count

import (
	"math/big"

	"github.com/Gulliby-development/youngkit/internal/shape"
)

// Rank returns the zero-based reverse-lexicographic position of s and the
// total number of partitions of its size.
func Rank(s shape.Shape) (rank, total *big.Int) {
	n := s.Size()
	counts := partitionCounts(n)
	rank = new(big.Int)

	remaining := n
	bound := n
	for _, part := range s {
		upper := bound
		if remaining < upper {
			upper = remaining
		}
		for candidate := part + 1; candidate <= upper; candidate++ {
			rank.Add(rank, &counts[remaining-candidate][candidate])
		}
		remaining -= part
		bound = part
	}

	total = new(big.Int).Set(&counts[n][n])
	return rank, total
}

// Unrank returns the partition at the zero-based reverse-lexicographic rank.
// The caller's rank is never modified.
func Unrank(n int, rank *big.Int) shape.Shape {
	counts := partitionCounts(n)
	remainder := new(big.Int).Set(rank)
	result := make(shape.Shape, 0)

	remaining := n
	bound := n
	for remaining > 0 {
		upper := bound
		if remaining < upper {
			upper = remaining
		}
		for candidate := upper; candidate > 0; candidate-- {
			block := &counts[remaining-candidate][candidate]
			if remainder.Cmp(block) < 0 {
				result = append(result, candidate)
				remaining -= candidate
				bound = candidate
				break
			}
			remainder.Sub(remainder, block)
		}
	}
	return result
}

// partitionCounts builds C(r,b), the number of partitions of r with parts at
// most b. Values are local to one Rank or Unrank operation and are immutable
// after the table is built.
func partitionCounts(n int) [][]big.Int {
	if n < 0 {
		return nil
	}

	counts := make([][]big.Int, n+1)
	for remaining := range counts {
		counts[remaining] = make([]big.Int, n+1)
	}
	for bound := 0; bound <= n; bound++ {
		counts[0][bound].SetInt64(1)
	}

	for remaining := 1; remaining <= n; remaining++ {
		for bound := 1; bound <= n; bound++ {
			counts[remaining][bound].Set(&counts[remaining][bound-1])
			if bound <= remaining {
				counts[remaining][bound].Add(
					&counts[remaining][bound],
					&counts[remaining-bound][bound],
				)
			}
		}
	}
	return counts
}
