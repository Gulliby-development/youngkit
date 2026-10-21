package cli_test

import "testing"

func TestRunNotationSevenCommands(t *testing.T) {
	requireCommandSuccess(t, []string{"rank", "4,3x2,1"},
		"shape       4,3,3,1\nsize        11\nrank        32\nof          56\n")
	requireCommandSuccess(t, []string{"conjugate", "3x2"},
		"shape       3,3\nsize        6\nconjugate   2,2,2\nself        no\nrow         ###\nrow         ###\n")
	requireCommandSuccess(t, []string{"core", "3", "4x2,1"},
		"shape       4,4,1\nk           3\ncore        0\nweight      3\nquotient    0=2\nquotient    1=0\nquotient    2=1\n")
	requireCommandSuccess(t, []string{"compare", "3x2", "4,1,1"},
		"a           3,3\nb           4,1,1\nsize        6\nsums-a      3 6 6\nsums-b      4 5 6\nrelation    incomparable\na-ahead     2\nb-ahead     1\n")
	requireCommandSuccess(t, []string{"hooks", "2x2,1"},
		"shape       2,2,1\nsize        5\nhook        4 2\nhook        3 1\nhook        1\nsyt         5\n")
	requireCommandSuccess(t, []string{"neighbors", "2x2"},
		"shape       2,2\nsize        4\nadds        2\nadd         3,2\nadd         2,2,1\nremoves     1\nremove      2,1\n")
	requireCommandSuccess(t, []string{"syt", "2x2"},
		"shape       2,2\nsize        4\ncount       2\nindex       0\nshown       2\ntableau     1,3/2,4\ntableau     1,2/3,4\n")
}

func TestRunNotationRunOfOne(t *testing.T) {
	requireCommandSuccess(t, []string{"rank", "3x1"},
		"shape       3\nsize        3\nrank        0\nof          3\n")
	requireCommandSuccess(t, []string{"rank", "2,1x1"},
		"shape       2,1\nsize        3\nrank        1\nof          3\n")
}

func TestRunNotationErrors(t *testing.T) {
	requireDispatchError(t, []string{"rank", "3x"}, nil,
		"invalid shape \"3x\" (expected parts such as 5,3,1, or 0 for the empty shape)", 0)
	requireDispatchError(t, []string{"rank", "3x0"}, nil,
		"invalid count \"0\" in \"3x0\" (expected 1 to 5000)", 0)
	requireDispatchError(t, []string{"hooks", "3x2,4"}, nil,
		"parts of \"3x2,4\" must not increase (part 3 is 4, after 3)", 0)
	requireDispatchError(t, []string{"conjugate", "2x5001"}, nil,
		"invalid count \"5001\" in \"2x5001\" (expected 1 to 5000)", 0)
	requireDispatchError(t, []string{"core", "3", "5000x2"}, nil,
		"shape \"5000x2\" has size 10000 (above 5000)", 0)
	requireDispatchError(t, []string{"rank", "1x5000"}, nil,
		"size 5000 is above 400 (too large for rank)", 0)
}
