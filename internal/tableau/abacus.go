package tableau

import (
	"sort"

	"github.com/Gulliby-development/youngkit/internal/shape"
)

// Abacus returns the K-core and the ordered K-quotient of s.
func Abacus(s shape.Shape, k int) (shape.Shape, []shape.Shape) {
	quotient := make([]shape.Shape, k)
	if len(s) == 0 {
		return shape.Shape{}, quotient
	}

	rows := ((len(s) + k - 1) / k) * k
	runners := make([][]int, k)
	for row := 0; row < rows; row++ {
		part := 0
		if row < len(s) {
			part = s[row]
		}
		bead := part + rows - row - 1
		runner := bead % k
		runners[runner] = append(runners[runner], bead/k)
	}

	coreBeads := make([]int, 0, rows)
	for runner, levels := range runners {
		sort.Sort(sort.Reverse(sort.IntSlice(levels)))
		quotientParts := make([]int, 0, len(levels))
		for index, level := range levels {
			part := level - (len(levels) - index - 1)
			if part > 0 {
				quotientParts = append(quotientParts, part)
			}
		}
		quotient[runner] = shape.Shape(quotientParts)

		for level := 0; level < len(levels); level++ {
			coreBeads = append(coreBeads, level*k+runner)
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(coreBeads)))
	coreParts := make(shape.Shape, 0, rows)
	for index, bead := range coreBeads {
		part := bead - (rows - index - 1)
		if part > 0 {
			coreParts = append(coreParts, part)
		}
	}
	return coreParts, quotient
}
