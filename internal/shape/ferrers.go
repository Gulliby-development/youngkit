package shape

import "strings"

// Conjugate reflects a shape across its main diagonal.
func Conjugate(s Shape) Shape {
	if len(s) == 0 {
		return Shape{}
	}

	conjugate := make(Shape, s[0])
	for _, rowLength := range s {
		for column := 0; column < rowLength; column++ {
			conjugate[column]++
		}
	}
	return conjugate
}

// Diagram returns one row of # characters for each part.
func Diagram(s Shape) []string {
	diagram := make([]string, len(s))
	for i, part := range s {
		diagram[i] = strings.Repeat("#", part)
	}
	return diagram
}
