package cli_test

import "testing"

func TestCountLargestMatchesSmallValues(t *testing.T) {
	requireCommandSuccess(t, []string{"count", "6", "--largest", "3"},
		"n           6\nmax-part    any\ndistinct    no\nlargest     3\ncount       3\n")
	requireCommandSuccess(t, []string{"count", "6", "--largest=6"},
		"n           6\nmax-part    any\ndistinct    no\nlargest     6\ncount       1\n")
	requireCommandSuccess(t, []string{"count", "6", "--largest", "1"},
		"n           6\nmax-part    any\ndistinct    no\nlargest     1\ncount       1\n")
}

func TestCountLargestWithDistinctParts(t *testing.T) {
	requireCommandSuccess(t, []string{"count", "10", "--distinct", "--largest", "5"},
		"n           10\nmax-part    any\ndistinct    yes\nlargest     5\ncount       2\n")
	requireCommandSuccess(t, []string{"count", "6", "--largest", "1", "--distinct"},
		"n           6\nmax-part    any\ndistinct    yes\nlargest     1\ncount       0\n")
}

func TestCountLargestAboveNOrOfEmptyPartitionIsZero(t *testing.T) {
	requireCommandSuccess(t, []string{"count", "4", "--largest", "5"},
		"n           4\nmax-part    any\ndistinct    no\nlargest     5\ncount       0\n")
	requireCommandSuccess(t, []string{"count", "0", "--largest", "1"},
		"n           0\nmax-part    any\ndistinct    no\nlargest     1\ncount       0\n")
}

func TestCountLargestLastValueWins(t *testing.T) {
	requireCommandSuccess(t, []string{"count", "6", "--largest=bad", "--largest", "2"},
		"n           6\nmax-part    any\ndistinct    no\nlargest     2\ncount       3\n")
}

func TestCountLargestErrors(t *testing.T) {
	requireDispatchError(t, []string{"count", "6", "--largest", "0"}, nil,
		"invalid --largest \"0\" (expected an integer from 1 to 5000)", 0)
	requireDispatchError(t, []string{"count", "6", "--largest"}, nil, "--largest needs a value", 0)
	requireDispatchError(t, []string{"count", "6", "--largest", "2", "--max-part", "3"}, nil,
		"--largest and --max-part exclude each other", 0)
	requireDispatchError(t, []string{"count", "6", "--max-part=3", "--largest=2"}, nil,
		"--largest and --max-part exclude each other", 0)
	requireDispatchError(t, []string{"list", "6", "--largest", "2"}, nil, "unknown option \"--largest\"", 0)
}
