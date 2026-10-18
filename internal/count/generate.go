package count

import "github.com/Gulliby-development/youngkit/internal/shape"

// Generate returns at most limit partitions of n, in reverse-lexicographic
// order, with every part at most k.
func Generate(n, k, limit int) []shape.Shape {
	if n < 0 || limit <= 0 {
		return nil
	}
	if n == 0 {
		return []shape.Shape{{}}
	}
	if k <= 0 {
		return nil
	}
	if k > n {
		k = n
	}

	partitions := make([]shape.Shape, 0)
	path := make(shape.Shape, 0)

	var visit func(remaining, maximum int) bool
	visit = func(remaining, maximum int) bool {
		if remaining == 0 {
			partition := make(shape.Shape, len(path))
			copy(partition, path)
			partitions = append(partitions, partition)
			return len(partitions) == limit
		}

		if maximum > remaining {
			maximum = remaining
		}
		for part := maximum; part > 0; part-- {
			path = append(path, part)
			stop := visit(remaining-part, part)
			path = path[:len(path)-1]
			if stop {
				return true
			}
		}
		return false
	}

	visit(n, k)
	return partitions
}
