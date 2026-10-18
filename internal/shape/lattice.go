package shape

// Adds returns each valid shape formed by adding one terminal box, in row order.
func Adds(s Shape) []Shape {
	adds := make([]Shape, 0, len(s)+1)
	for row := range s {
		if row == 0 || s[row] < s[row-1] {
			added := clone(s)
			added[row]++
			adds = append(adds, added)
		}
	}

	addedRow := clone(s)
	adds = append(adds, append(addedRow, 1))
	return adds
}

// Removes returns each valid shape formed by removing one terminal box, in row order.
func Removes(s Shape) []Shape {
	removes := make([]Shape, 0, len(s))
	for row := range s {
		if row < len(s)-1 && s[row] == s[row+1] {
			continue
		}

		removed := clone(s)
		if removed[row] == 1 {
			removed = append(removed[:row], removed[row+1:]...)
		} else {
			removed[row]--
		}
		removes = append(removes, removed)
	}
	return removes
}

func clone(s Shape) Shape {
	return append(Shape(nil), s...)
}
