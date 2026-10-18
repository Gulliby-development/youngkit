package count

import "math/big"

// Partitions returns the number of partitions of n whose parts are at most k.
// When distinct is true, no part may be repeated.
func Partitions(n, k int, distinct bool) *big.Int {
	if n < 0 {
		return new(big.Int)
	}
	if n == 0 {
		return big.NewInt(1)
	}
	if k <= 0 {
		return new(big.Int)
	}
	if k > n {
		k = n
	}

	// Each entry represents the number of ways to form its index using the
	// parts processed so far. Ascending sums allow a part to be reused;
	// descending sums ensure a part is used at most once.
	ways := make([]big.Int, n+1)
	ways[0].SetInt64(1)
	for part := 1; part <= k; part++ {
		if distinct {
			for sum := n; sum >= part; sum-- {
				ways[sum].Add(&ways[sum], &ways[sum-part])
			}
			continue
		}
		for sum := part; sum <= n; sum++ {
			ways[sum].Add(&ways[sum], &ways[sum-part])
		}
	}

	return new(big.Int).Set(&ways[n])
}
