package shape

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
)

// Shape is a partition written as positive, nonincreasing parts.
type Shape []int

// Parse parses a shape, validating token syntax, part ranges, ordering, and size
// in that order.
func Parse(text string) (Shape, error) {
	parts := strings.Split(text, ",")
	values := make([]int64, len(parts))

	// Validate every token's syntax before checking any semantic constraints.
	for i, part := range parts {
		value, err := base.ParseInt("part", part, 0, 9999999999)
		if err != nil {
			return nil, base.Fail("invalid shape " + base.Show(text) +
				" (expected parts such as 5,3,1, or 0 for the empty shape)")
		}
		values[i] = value
	}

	if text == "0" {
		return Shape{}, nil
	}

	for i, value := range values {
		if value < 1 || value > 5000 {
			return nil, base.Fail("invalid part " + base.Show(parts[i]) +
				" in " + base.Show(text) + " (expected 1 to 5000)")
		}
	}

	for i := 1; i < len(values); i++ {
		if values[i] > values[i-1] {
			return nil, base.Fail("parts of " + base.Show(text) +
				" must not increase (part " + strconv.Itoa(i+1) + " is " +
				parts[i] + ", after " + parts[i-1] + ")")
		}
	}

	var total big.Int
	for _, value := range values {
		total.Add(&total, big.NewInt(value))
	}
	if total.Cmp(big.NewInt(5000)) > 0 {
		return nil, base.Fail("shape " + base.Show(text) + " has size " +
			total.String() + " (above 5000)")
	}

	result := make(Shape, len(values))
	for i, value := range values {
		result[i] = int(value)
	}
	return result, nil
}

// String returns the comma-separated parts, or 0 for the empty shape.
func (s Shape) String() string {
	if len(s) == 0 {
		return "0"
	}

	var text strings.Builder
	for i, part := range s {
		if i > 0 {
			text.WriteByte(',')
		}
		text.WriteString(strconv.Itoa(part))
	}
	return text.String()
}

// Size returns the sum of the parts.
func (s Shape) Size() int {
	size := 0
	for _, part := range s {
		size += part
	}
	return size
}

// Limit reports an error when the shape's size exceeds limit.
func (s Shape) Limit(limit int, command string) error {
	size := s.Size()
	if size <= limit {
		return nil
	}
	return base.Fail("size " + strconv.Itoa(size) + " is above " +
		strconv.Itoa(limit) + " (too large for " + command + ")")
}
