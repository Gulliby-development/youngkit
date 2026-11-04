package count

import "math/big"

// PartitionsWithLargest returns the number of partitions of n whose largest
// part is exactly k. When distinct is true, no part may be repeated. The empty
// partition of 0 has no largest part, so it is never counted.
func PartitionsWithLargest(n, k int, distinct bool) *big.Int {
	if n <= 0 || k <= 0 {
		return new(big.Int)
	}
	return new(big.Int).Sub(Partitions(n, k, distinct), Partitions(n, k-1, distinct))
}
