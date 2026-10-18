package cli_test

import (
	"testing"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/cli"
)

type commandAcceptanceIO struct {
	stdin    []byte
	reads    int
	outCalls []string
	errCalls []string
}

func (io *commandAcceptanceIO) ReadStdin() []byte {
	io.reads++
	return io.stdin
}

func (io *commandAcceptanceIO) Out(text string) {
	io.outCalls = append(io.outCalls, text)
}

func (io *commandAcceptanceIO) Err(text string) {
	io.errCalls = append(io.errCalls, text)
}

func requireCommandSuccess(t *testing.T, args []string, want string) {
	t.Helper()

	io := &commandAcceptanceIO{}
	if got := cli.Run(args, io); got != 0 {
		t.Fatalf("cli.Run(%q) status = %d, want 0 (stderr calls: %#v)", args, got, io.errCalls)
	}
	if len(io.outCalls) != 1 || io.outCalls[0] != want {
		t.Fatalf("cli.Run(%q) stdout calls = %#v, want one exact call %q", args, io.outCalls, want)
	}
	if len(io.errCalls) != 0 {
		t.Fatalf("cli.Run(%q) stderr calls = %#v, want none", args, io.errCalls)
	}
	if io.reads != 0 {
		t.Fatalf("cli.Run(%q) read stdin %d times, want none", args, io.reads)
	}
}

func TestCommandCountLiteralExamples(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"count", "10"}, "n           10\nmax-part    any\ndistinct    no\ncount       42\n"},
		{[]string{"count", "50"}, "n           50\nmax-part    any\ndistinct    no\ncount       204226\n"},
		{[]string{"count", "100"}, "n           100\nmax-part    any\ndistinct    no\ncount       190569292\n"},
		{[]string{"count", "100", "--distinct"}, "n           100\nmax-part    any\ndistinct    yes\ncount       444793\n"},
		{[]string{"count", "1000"}, "n           1000\nmax-part    any\ndistinct    no\ncount       24061467864032622473692149727991\n"},
		{[]string{"count", "6", "--max-part=3"}, "n           6\nmax-part    3\ndistinct    no\ncount       7\n"},
		{[]string{"count", "0", "--distinct"}, "n           0\nmax-part    any\ndistinct    yes\ncount       1\n"},
	}
	for _, test := range tests {
		requireCommandSuccess(t, test.args, test.want)
	}
}

func TestCommandListReverseLexicographicExample(t *testing.T) {
	requireCommandSuccess(t, []string{"list", "6", "--limit", "5"},
		"n           6\nmax-part    any\ncount       11\nshown       5\npart        6\npart        5,1\npart        4,2\npart        4,1,1\npart        3,3\n")
}

func TestCommandListRestrictedTwoHundredExample(t *testing.T) {
	requireCommandSuccess(t, []string{"list", "200", "--max-part", "3", "--limit", "3"},
		"n           200\nmax-part    3\ncount       3434\nshown       3\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,1,1\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,1\n")
}

func TestCommandListEmptyShapeAndDefaultLimit(t *testing.T) {
	requireCommandSuccess(t, []string{"list", "0"},
		"n           0\nmax-part    any\ncount       1\nshown       1\npart        0\n")
}

func TestCommandRankLiteralExampleAndEmptyPartition(t *testing.T) {
	requireCommandSuccess(t, []string{"rank", "4,2,1"},
		"shape       4,2,1\nsize        7\nrank        5\nof          15\n")
	requireCommandSuccess(t, []string{"rank", "0"},
		"shape       0\nsize        0\nrank        0\nof          1\n")
}

func TestCommandUnrankLiteralExamplesAndEmptyPartition(t *testing.T) {
	requireCommandSuccess(t, []string{"unrank", "6", "5"},
		"n           6\nrank        5\nshape       3,2,1\n")
	requireCommandSuccess(t, []string{"unrank", "400", "6123456789012345678"},
		"n           400\nrank        6123456789012345678\nshape       34,24,22,20,18,18,18,17,16,14,12,9,9,9,9,8,5,5,5,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1\n")
	requireCommandSuccess(t, []string{"unrank", "0", "0"},
		"n           0\nrank        0\nshape       0\n")
}

func TestCommandConjugateLiteralExample(t *testing.T) {
	requireCommandSuccess(t, []string{"conjugate", "5,3,3,1"},
		"shape       5,3,3,1\nsize        12\nconjugate   4,3,3,1,1\nself        no\nrow         #####\nrow         ###\nrow         ###\nrow         #\n")
}

func TestCommandConjugateSelfAndEmptyShapes(t *testing.T) {
	requireCommandSuccess(t, []string{"conjugate", "3,2,1"},
		"shape       3,2,1\nsize        6\nconjugate   3,2,1\nself        yes\nrow         ###\nrow         ##\nrow         #\n")
	requireCommandSuccess(t, []string{"conjugate", "0"},
		"shape       0\nsize        0\nconjugate   0\nself        yes\n")
}

func TestCommandHooksLiteralExampleAndOmittedEntries(t *testing.T) {
	requireCommandSuccess(t, []string{"hooks", "4,3,1", "--entries", "3"},
		"shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\nentries     3\nssyt        15\n")
	requireCommandSuccess(t, []string{"hooks", "4,3,1"},
		"shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\n")
}

func TestCommandHooksEmptyShape(t *testing.T) {
	requireCommandSuccess(t, []string{"hooks", "0", "--entries=1"},
		"shape       0\nsize        0\nsyt         1\nentries     1\nssyt        1\n")
}

func TestCommandCompareIncomparableLiteralExample(t *testing.T) {
	requireCommandSuccess(t, []string{"compare", "3,3", "4,1,1"},
		"a           3,3\nb           4,1,1\nsize        6\nsums-a      3 6 6\nsums-b      4 5 6\nrelation    incomparable\na-ahead     2\nb-ahead     1\n")
}

func TestCommandCompareOtherRelationsAndEmptyShapes(t *testing.T) {
	requireCommandSuccess(t, []string{"compare", "3,2,1", "2,2,2"},
		"a           3,2,1\nb           2,2,2\nsize        6\nsums-a      3 5 6\nsums-b      2 4 6\nrelation    a dominates b\n")
	requireCommandSuccess(t, []string{"compare", "2,2,2", "3,2,1"},
		"a           2,2,2\nb           3,2,1\nsize        6\nsums-a      2 4 6\nsums-b      3 5 6\nrelation    b dominates a\n")
	requireCommandSuccess(t, []string{"compare", "2,1", "2,1"},
		"a           2,1\nb           2,1\nsize        3\nsums-a      2 3\nsums-b      2 3\nrelation    equal\n")
	requireCommandSuccess(t, []string{"compare", "0", "0"},
		"a           0\nb           0\nsize        0\nsums-a      0\nsums-b      0\nrelation    equal\n")
}

func TestCommandNeighborsLiteralExample(t *testing.T) {
	requireCommandSuccess(t, []string{"neighbors", "4,2,2"},
		"shape       4,2,2\nsize        8\nadds        3\nadd         5,2,2\nadd         4,3,2\nadd         4,2,2,1\nremoves     2\nremove      3,2,2\nremove      4,2,1\n")
}

func TestCommandNeighborsEmptyShape(t *testing.T) {
	requireCommandSuccess(t, []string{"neighbors", "0"},
		"shape       0\nsize        0\nadds        1\nadd         1\nremoves     0\n")
}

func TestCommandRSKLiteralExample(t *testing.T) {
	requireCommandSuccess(t, []string{"rsk", "2", "5", "1", "4", "3"},
		"perm        2 5 1 4 3\nshape       2,2,1\nlis         2\nlds         3\np           1 3\np           2 4\np           5\nq           1 2\nq           3 4\nq           5\n")
}

func TestCommandUnrskLiteralExampleAndEmptyTableaux(t *testing.T) {
	requireCommandSuccess(t, []string{"unrsk", "1,3/2,4/5", "1,2/3,5/4"},
		"shape       2,2,1\nsize        5\nperm        2 5 4 1 3\n")
	requireCommandSuccess(t, []string{"unrsk", "0", "0"},
		"shape       0\nsize        0\nperm        0\n")
}

func TestCommandSytFiveTableauxOfThreeTwo(t *testing.T) {
	requireCommandSuccess(t, []string{"syt", "3,2", "--limit", "5"},
		"shape       3,2\nsize        5\ncount       5\nindex       0\nshown       5\ntableau     1,3,5/2,4\ntableau     1,2,5/3,4\ntableau     1,3,4/2,5\ntableau     1,2,4/3,5\ntableau     1,2,3/4,5\n")
}

func TestCommandSytSuppliedLargeIndexAndLimit(t *testing.T) {
	requireCommandSuccess(t, []string{"syt", "8,8,8,8,8", "--index", "9999999999", "--limit", "2"},
		"shape       8,8,8,8,8\nsize        40\ncount       231471904322784840\nindex       9999999999\nshown       2\ntableau     1,2,3,9,12,24,31,36/4,5,8,11,13,26,32,37/6,10,15,18,22,28,33,38/7,14,16,20,25,29,34,39/17,19,21,23,27,30,35,40\ntableau     1,4,6,9,12,24,31,36/2,5,7,11,13,26,32,37/3,10,15,18,22,28,33,38/8,14,16,20,25,29,34,39/17,19,21,23,27,30,35,40\n")
}

func TestCommandSytEmptyShapeAndDefaultOptions(t *testing.T) {
	requireCommandSuccess(t, []string{"syt", "0"},
		"shape       0\nsize        0\ncount       1\nindex       0\nshown       1\ntableau     0\n")
}

func TestCommandCoreLiteralExamplesAndEmptyShape(t *testing.T) {
	requireCommandSuccess(t, []string{"core", "3", "5,4,1"},
		"shape       5,4,1\nk           3\ncore        2,1,1\nweight      2\nquotient    0=0\nquotient    1=1\nquotient    2=1\n")
	requireCommandSuccess(t, []string{"core", "2", "4,3,2,1"},
		"shape       4,3,2,1\nk           2\ncore        4,3,2,1\nweight      0\nquotient    0=0\nquotient    1=0\n")
	requireCommandSuccess(t, []string{"core", "2", "0"},
		"shape       0\nk           2\ncore        0\nweight      0\nquotient    0=0\nquotient    1=0\n")
}

// Keep this compile-time use near the local fake so the fixture stays tied to the
// exact IO contract exercised by cli.Run.
var _ base.IO = (*commandAcceptanceIO)(nil)
