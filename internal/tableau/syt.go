package tableau

import (
	"math/big"

	"github.com/Gulliby-development/youngkit/internal/count"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// Unrank returns the standard tableau at index in the corner-removal order.
func Unrank(s shape.Shape, index *big.Int) Tableau {
	remaining := append(shape.Shape(nil), s...)
	result := make(Tableau, len(s))
	for row, width := range s {
		result[row] = make([]int, width)
	}
	residual := new(big.Int).Set(index)

	for value := s.Size(); value > 0; value-- {
		for row := 0; row < len(remaining); row++ {
			if row+1 < len(remaining) && remaining[row] == remaining[row+1] {
				continue
			}

			branch := make(shape.Shape, 0, len(remaining))
			branch = append(branch, remaining[:row]...)
			if remaining[row] > 1 {
				branch = append(branch, remaining[row]-1)
			}
			branch = append(branch, remaining[row+1:]...)

			branchCount := count.SYT(branch)
			if residual.Cmp(branchCount) >= 0 {
				residual.Sub(residual, branchCount)
				continue
			}

			result[row][remaining[row]-1] = value
			remaining = branch
			break
		}
	}

	return result
}
