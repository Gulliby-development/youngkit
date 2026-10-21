package shape

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
)

// Shape is a partition written as positive, nonincreasing parts.
type Shape []int

// Parse parses a shape, validating token syntax, part and count ranges,
// ordering, and size in that order. A token may be a run PARTxCOUNT.
func Parse(text string) (Shape, error) {
	tokens := strings.Split(text, ",")
	runs := make([]run, len(tokens))

	// Validate every token's syntax before checking any semantic constraints.
	for i, token := range tokens {
		r, ok := readRun(token)
		if !ok {
			return nil, base.Fail("invalid shape " + base.Show(text) +
				" (expected parts such as 5,3,1, or 0 for the empty shape)")
		}
		runs[i] = r
	}

	if text == "0" {
		return Shape{}, nil
	}

	for _, r := range runs {
		if r.part < 1 || r.part > 5000 {
			return nil, base.Fail("invalid part " + base.Show(r.partText) +
				" in " + base.Show(text) + " (expected 1 to 5000)")
		}
		if r.count < 1 || r.count > 5000 {
			return nil, base.Fail("invalid count " + base.Show(r.countText) +
				" in " + base.Show(text) + " (expected 1 to 5000)")
		}
	}

	position := int64(0)
	for i, r := range runs {
		if i > 0 && r.part > runs[i-1].part {
			return nil, base.Fail("parts of " + base.Show(text) +
				" must not increase (part " + strconv.FormatInt(position+1, 10) +
				" is " + r.partText + ", after " + runs[i-1].partText + ")")
		}
		position += r.count
	}

	var total big.Int
	for _, r := range runs {
		var product big.Int
		product.Mul(big.NewInt(r.part), big.NewInt(r.count))
		total.Add(&total, &product)
	}
	if total.Cmp(big.NewInt(5000)) > 0 {
		return nil, base.Fail("shape " + base.Show(text) + " has size " +
			total.String() + " (above 5000)")
	}

	result := Shape{}
	for _, r := range runs {
		for j := int64(0); j < r.count; j++ {
			result = append(result, int(r.part))
		}
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
