package count

import (
	"math/big"

	"github.com/Gulliby-development/youngkit/internal/shape"
)

// Hooks returns the hook lengths of s, with rows and cells in diagram order.
func Hooks(s shape.Shape) [][]int {
	columns := shape.Conjugate(s)
	hooks := make([][]int, len(s))
	for row, width := range s {
		hooks[row] = make([]int, width)
		for column := 0; column < width; column++ {
			hooks[row][column] = width - column + columns[column] - row - 1
		}
	}
	return hooks
}

// SYT returns the number of standard Young tableaux of s.
func SYT(s shape.Shape) *big.Int {
	hooks := Hooks(s)
	factorial := big.NewInt(1)
	for factor := 2; factor <= s.Size(); factor++ {
		factorial.Mul(factorial, big.NewInt(int64(factor)))
	}

	denominator := big.NewInt(1)
	for _, row := range hooks {
		for _, hook := range row {
			denominator.Mul(denominator, big.NewInt(int64(hook)))
		}
	}
	return factorial.Quo(factorial, denominator)
}

// SSYT returns the number of semistandard Young tableaux of s using entries
// from 1 through entries.
func SSYT(s shape.Shape, entries int) *big.Int {
	if len(s) == 0 {
		return big.NewInt(1)
	}
	if entries < len(s) {
		return new(big.Int)
	}

	hooks := Hooks(s)
	numerator := big.NewInt(1)
	denominator := big.NewInt(1)
	for row, width := range s {
		for column := 0; column < width; column++ {
			numerator.Mul(numerator, big.NewInt(int64(entries+column-row)))
			denominator.Mul(denominator, big.NewInt(int64(hooks[row][column])))
		}
	}
	return numerator.Quo(numerator, denominator)
}
