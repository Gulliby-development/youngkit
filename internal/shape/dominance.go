package shape

// Dominance describes the prefix-sum relation between two equal-size shapes.
type Dominance struct {
	Relation string
	SumsA    []int
	SumsB    []int
	AAhead   int
	BAhead   int
}

// Compare computes padded prefix sums and their dominance relation.
func Compare(a, b Shape) Dominance {
	length := len(a)
	if len(b) > length {
		length = len(b)
	}

	result := Dominance{
		SumsA: make([]int, length),
		SumsB: make([]int, length),
	}
	aSum, bSum := 0, 0
	aGreater, bGreater := false, false
	for i := 0; i < length; i++ {
		if i < len(a) {
			aSum += a[i]
		}
		if i < len(b) {
			bSum += b[i]
		}
		result.SumsA[i] = aSum
		result.SumsB[i] = bSum

		if aSum > bSum {
			aGreater = true
			if result.AAhead == 0 {
				result.AAhead = i + 1
			}
		} else if bSum > aSum {
			bGreater = true
			if result.BAhead == 0 {
				result.BAhead = i + 1
			}
		}
	}

	switch {
	case !aGreater && !bGreater:
		result.Relation = "equal"
	case aGreater && !bGreater:
		result.Relation = "a dominates b"
	case !aGreater && bGreater:
		result.Relation = "b dominates a"
	default:
		result.Relation = "incomparable"
	}
	return result
}
