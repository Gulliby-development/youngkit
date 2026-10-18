package count_test

import (
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/cli"
	"github.com/Gulliby-development/youngkit/internal/count"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

type partitionTestIO struct {
	stdin    []byte
	reads    int
	outCalls int
	errCalls int
	outText  string
	errText  string
}

func (io *partitionTestIO) ReadStdin() []byte {
	io.reads++
	return io.stdin
}

func (io *partitionTestIO) Out(text string) {
	io.outCalls++
	io.outText += text
}

func (io *partitionTestIO) Err(text string) {
	io.errCalls++
	io.errText += text
}

func checkPartitionRun(t *testing.T, run func([]string, base.IO) (string, error), args []string, wantOutput, wantError string) {
	t.Helper()
	io := &partitionTestIO{}
	got, err := run(args, io)
	if wantError == "" {
		if err != nil {
			t.Fatalf("run(%q) error = %v, want nil", args, err)
		}
		if got != wantOutput {
			t.Fatalf("run(%q) output = %q, want %q", args, got, wantOutput)
		}
	} else {
		if got != "" {
			t.Fatalf("run(%q) output = %q, want empty on error", args, got)
		}
		if err == nil {
			t.Fatalf("run(%q) error = nil, want %q", args, wantError)
		}
		var failure *base.Error
		if !errors.As(err, &failure) {
			t.Fatalf("run(%q) error type = %T, want *base.Error", args, err)
		}
		if failure.Msg != wantError || err.Error() != wantError {
			t.Fatalf("run(%q) error = %q, want %q", args, err.Error(), wantError)
		}
	}
	if io.reads != 0 || io.outCalls != 0 || io.errCalls != 0 || io.outText != "" || io.errText != "" {
		t.Fatalf("adapter used IO: reads=%d out=%d/%q err=%d/%q", io.reads, io.outCalls, io.outText, io.errCalls, io.errText)
	}
}

func partitionShape(t *testing.T, text string) shape.Shape {
	t.Helper()
	got, err := shape.Parse(text)
	if err != nil {
		t.Fatalf("shape.Parse(%q): %v", text, err)
	}
	return got
}

func TestPartitionsLiteralKnownCounts(t *testing.T) {
	tests := []struct {
		n, k     int
		distinct bool
		want     string
	}{
		{0, 0, false, "1"},
		{0, 5, true, "1"},
		{10, 10, false, "42"},
		{50, 50, false, "204226"},
		{100, 100, false, "190569292"},
		{100, 100, true, "444793"},
		{1000, 1000, false, "24061467864032622473692149727991"},
		{6, 3, false, "7"},
		{6, 3, true, "1"},
		{6, 20, false, "11"},
		{200, 3, false, "3434"},
	}
	for _, test := range tests {
		got := count.Partitions(test.n, test.k, test.distinct)
		if got.String() != test.want {
			t.Errorf("Partitions(%d, %d, %t) = %s, want %s", test.n, test.k, test.distinct, got, test.want)
		}
	}
}

func TestPartitionsClampBoundsAndEmpty(t *testing.T) {
	if got := count.Partitions(0, 0, false).String(); got != "1" {
		t.Fatalf("Partitions(0, 0, false) = %s, want 1", got)
	}
	if got := count.Partitions(6, 100, false).String(); got != "11" {
		t.Fatalf("Partitions(6, 100, false) = %s, want 11", got)
	}
	if got := count.Partitions(6, 1, false).String(); got != "1" {
		t.Fatalf("Partitions(6, 1, false) = %s, want 1", got)
	}
	if got := count.Partitions(6, 0, true).String(); got != "0" {
		t.Fatalf("Partitions(6, 0, true) = %s, want 0", got)
	}
}

func TestPartitionsDistinctAndRestrictedBoundaries(t *testing.T) {
	if got := count.Partitions(6, 6, true).String(); got != "4" {
		t.Fatalf("Partitions(6, 6, true) = %s, want 4", got)
	}
	if got := count.Partitions(6, 3, true).String(); got != "1" {
		t.Fatalf("Partitions(6, 3, true) = %s, want 1", got)
	}
	if got := count.Partitions(1, 1, true).String(); got != "1" {
		t.Fatalf("Partitions(1, 1, true) = %s, want 1", got)
	}
}

func TestGenerateReverseLexicographicLiteralPrefix(t *testing.T) {
	got := count.Generate(6, 6, 5)
	want := []shape.Shape{{6}, {5, 1}, {4, 2}, {4, 1, 1}, {3, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Generate(6, 6, 5) = %#v, want %#v", got, want)
	}
}

func TestGenerateRestrictedLiteralPrefix(t *testing.T) {
	got := count.Generate(6, 3, 3)
	want := []shape.Shape{{3, 3}, {3, 2, 1}, {3, 1, 1, 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Generate(6, 3, 3) = %#v, want %#v", got, want)
	}
}

func TestGenerateEmptyAndLimitBoundaries(t *testing.T) {
	if got := count.Generate(0, 0, 20); !reflect.DeepEqual(got, []shape.Shape{{}}) {
		t.Fatalf("Generate(0, 0, 20) = %#v, want one empty shape", got)
	}
	if got := count.Generate(0, 0, 0); got != nil {
		t.Fatalf("Generate(0, 0, 0) = %#v, want nil", got)
	}
	if got := count.Generate(3, 3, 20); !reflect.DeepEqual(got, []shape.Shape{{3}, {2, 1}, {1, 1, 1}}) {
		t.Fatalf("Generate(3, 3, 20) = %#v", got)
	}
	if got := count.Generate(3, 3, 1); !reflect.DeepEqual(got, []shape.Shape{{3}}) {
		t.Fatalf("Generate(3, 3, 1) = %#v, want only [3]", got)
	}
	if got := count.Generate(3, 0, 20); got != nil {
		t.Fatalf("Generate(3, 0, 20) = %#v, want nil", got)
	}
}

func TestGenerateSize200StopsAtRequestedPrefix(t *testing.T) {
	got := count.Generate(200, 200, 3)
	want := []shape.Shape{{200}, {199, 1}, {198, 2}}
	if !reflect.DeepEqual(got, want) || len(got) != 3 {
		t.Fatalf("Generate(200, 200, 3) = %#v, want exactly %#v", got, want)
	}
}

func TestGenerateResultsHaveIndependentStorage(t *testing.T) {
	got := count.Generate(6, 6, 3)
	got[0][0] = 99
	if !reflect.DeepEqual(got[1], shape.Shape{5, 1}) || !reflect.DeepEqual(got[2], shape.Shape{4, 2}) {
		t.Fatalf("mutating one generated shape changed siblings: %#v", got)
	}
}

func TestRankFixedLiteralShapes(t *testing.T) {
	tests := []struct {
		text, rank, total string
	}{
		{"0", "0", "1"},
		{"6", "0", "11"},
		{"3,2,1", "5", "11"},
		{"1,1,1,1,1,1", "10", "11"},
		{"4,2,1", "5", "15"},
	}
	for _, test := range tests {
		rank, total := count.Rank(partitionShape(t, test.text))
		if rank.String() != test.rank || total.String() != test.total {
			t.Errorf("Rank(%s) = %s of %s, want %s of %s", test.text, rank, total, test.rank, test.total)
		}
	}
}

func TestUnrankFixedLiteralPositions(t *testing.T) {
	tests := []struct {
		n    int
		rank string
		want string
	}{
		{0, "0", "0"},
		{6, "0", "6"},
		{6, "5", "3,2,1"},
		{6, "10", "1,1,1,1,1,1"},
		{7, "5", "4,2,1"},
	}
	for _, test := range tests {
		rank, ok := new(big.Int).SetString(test.rank, 10)
		if !ok {
			t.Fatalf("bad literal rank %q", test.rank)
		}
		got := count.Unrank(test.n, rank)
		if got.String() != test.want {
			t.Errorf("Unrank(%d, %s) = %s, want %s", test.n, test.rank, got, test.want)
		}
	}
}

func TestRankAndUnrankSize400Literal(t *testing.T) {
	const requestedShape400 = "34,24,22,20,18,18,18,17,16,14,12,9,9,9,9,8,5,5,5,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1"
	rank, total := count.Rank(partitionShape(t, requestedShape400))
	if rank.String() != "6123456789012345678" || total.String() != "6727090051741041926" {
		t.Fatalf("Rank(size-400 literal) = %s of %s, want 6123456789012345678 of 6727090051741041926", rank, total)
	}
	rankInput, ok := new(big.Int).SetString("6123456789012345678", 10)
	if !ok {
		t.Fatal("could not parse literal rank")
	}
	got := count.Unrank(400, rankInput)
	if got.String() != requestedShape400 {
		t.Fatalf("Unrank(400, 6123456789012345678) = %s, want %s", got, requestedShape400)
	}
}

func TestUnrankDoesNotModifyCallerRank(t *testing.T) {
	rank := big.NewInt(5)
	if got := count.Unrank(6, rank).String(); got != "3,2,1" {
		t.Fatalf("Unrank(6, 5) = %s, want 3,2,1", got)
	}
	if rank.String() != "5" {
		t.Fatalf("Unrank changed caller rank to %s, want 5", rank)
	}
}

func TestRunCountRequestExamples(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"10"}, "n           10\nmax-part    any\ndistinct    no\ncount       42\n"},
		{[]string{"100", "--distinct"}, "n           100\nmax-part    any\ndistinct    yes\ncount       444793\n"},
		{[]string{"1000"}, "n           1000\nmax-part    any\ndistinct    no\ncount       24061467864032622473692149727991\n"},
	}
	for _, test := range tests {
		checkPartitionRun(t, cli.RunCount, test.args, test.want, "")
	}
}

func TestRunCountOptionValuesAndBounds(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"6", "--max-part=3"}, "n           6\nmax-part    3\ndistinct    no\ncount       7\n"},
		{[]string{"0", "--distinct"}, "n           0\nmax-part    any\ndistinct    yes\ncount       1\n"},
		{[]string{"0", "--max-part=5000"}, "n           0\nmax-part    5000\ndistinct    no\ncount       1\n"},
		{[]string{"6", "--max-part", "2", "--max-part=3"}, "n           6\nmax-part    3\ndistinct    no\ncount       7\n"},
		{[]string{"6", "--max-part=1"}, "n           6\nmax-part    1\ndistinct    no\ncount       1\n"},
	}
	for _, test := range tests {
		checkPartitionRun(t, cli.RunCount, test.args, test.want, "")
	}
}

func TestRunListRequestExample(t *testing.T) {
	checkPartitionRun(t, cli.RunList, []string{"6", "--limit", "5"},
		"n           6\nmax-part    any\ncount       11\nshown       5\npart        6\npart        5,1\npart        4,2\npart        4,1,1\npart        3,3\n", "")
}

func TestRunListRestricted200RequestExample(t *testing.T) {
	checkPartitionRun(t, cli.RunList, []string{"200", "--max-part", "3", "--limit", "3"},
		"n           200\nmax-part    3\ncount       3434\nshown       3\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,1,1\npart        3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,3,2,2,1\n", "")
}

func TestRunListDefaultAndMaximumLimit(t *testing.T) {
	checkPartitionRun(t, cli.RunList, []string{"0"},
		"n           0\nmax-part    any\ncount       1\nshown       1\npart        0\n", "")
	checkPartitionRun(t, cli.RunList, []string{"0", "--limit=1000"},
		"n           0\nmax-part    any\ncount       1\nshown       1\npart        0\n", "")
	checkPartitionRun(t, cli.RunList, []string{"6", "--max-part=200", "--limit=1"},
		"n           6\nmax-part    200\ncount       11\nshown       1\npart        6\n", "")
}

func TestRunListDefaultLimitShowsTwentyRows(t *testing.T) {
	io := &partitionTestIO{}
	got, err := cli.RunList([]string{"200"}, io)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(got, "\n")
	if strings.Count(got, "\n") != 24 || len(lines) != 25 {
		t.Fatalf("default list has %d lines, want 24 output rows", strings.Count(got, "\n"))
	}
	if lines[0] != "n           200" || lines[1] != "max-part    any" ||
		lines[2] != "count       3972999029388" || lines[3] != "shown       20" {
		t.Fatalf("default list header = %#v", lines[:4])
	}
	if lines[4] != "part        200" || lines[5] != "part        199,1" ||
		lines[23] != "part        194,6" || lines[24] != "" {
		t.Fatalf("default list selected rows = %#v / %#v / %#v", lines[4], lines[5], lines[23])
	}
	if io.reads != 0 || io.outCalls != 0 || io.errCalls != 0 || io.outText != "" || io.errText != "" {
		t.Fatalf("RunList used IO: reads=%d out=%d/%q err=%d/%q", io.reads, io.outCalls, io.outText, io.errCalls, io.errText)
	}
}

func TestRunRankRequestAndEmptyExamples(t *testing.T) {
	checkPartitionRun(t, cli.RunRank, []string{"4,2,1"},
		"shape       4,2,1\nsize        7\nrank        5\nof          15\n", "")
	checkPartitionRun(t, cli.RunRank, []string{"0"},
		"shape       0\nsize        0\nrank        0\nof          1\n", "")
}

func TestRunUnrankRequestExamples(t *testing.T) {
	checkPartitionRun(t, cli.RunUnrank, []string{"6", "5"},
		"n           6\nrank        5\nshape       3,2,1\n", "")
	checkPartitionRun(t, cli.RunUnrank, []string{"400", "6123456789012345678"},
		"n           400\nrank        6123456789012345678\nshape       34,24,22,20,18,18,18,17,16,14,12,9,9,9,9,8,5,5,5,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1\n", "")
}

func TestRunUnrankEmptyAndLastValidIndex(t *testing.T) {
	checkPartitionRun(t, cli.RunUnrank, []string{"0", "0"},
		"n           0\nrank        0\nshape       0\n", "")
	checkPartitionRun(t, cli.RunUnrank, []string{"6", "10"},
		"n           6\nrank        10\nshape       1,1,1,1,1,1\n", "")
}

func TestRunUnrankRejectsMalformedRankTokens(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"", "invalid rank \"\" (expected an integer from 0 to 0)"},
		{"00", "invalid rank \"00\" (expected an integer from 0 to 0)"},
		{"01", "invalid rank \"01\" (expected an integer from 0 to 0)"},
		{"+3", "invalid rank \"+3\" (expected an integer from 0 to 0)"},
		{"-1", "invalid rank \"-1\" (expected an integer from 0 to 0)"},
		{"1.5", "invalid rank \"1.5\" (expected an integer from 0 to 0)"},
		{" 1", "invalid rank \"?1\" (expected an integer from 0 to 0)"},
		{"٣", "invalid rank \"?\" (expected an integer from 0 to 0)"},
		{"2", "invalid rank \"2\" (expected an integer from 0 to 0)"},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", "invalid rank \"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999\" (expected an integer from 0 to 0)"},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", "invalid rank \"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999\" (expected an integer from 0 to 0)"},
	}
	for _, test := range tests {
		checkPartitionRun(t, cli.RunUnrank, []string{"0", test.token}, "", test.want)
	}
}

func TestRunUnrankChecksNBeforeMalformedRank(t *testing.T) {
	checkPartitionRun(t, cli.RunUnrank, []string{"401", "not-a-rank"}, "",
		"invalid n \"401\" (expected an integer from 0 to 400)")
}

func TestAdaptersUsageAndOptionErrors(t *testing.T) {
	checkPartitionRun(t, cli.RunCount, nil, "", "usage: youngkit count N [--max-part K] [--distinct]")
	checkPartitionRun(t, cli.RunList, nil, "", "usage: youngkit list N [--max-part K] [--limit L]")
	checkPartitionRun(t, cli.RunRank, nil, "", "usage: youngkit rank SHAPE")
	checkPartitionRun(t, cli.RunUnrank, nil, "", "usage: youngkit unrank N RANK")
	checkPartitionRun(t, cli.RunList, []string{"3", "4"}, "", "usage: youngkit list N [--max-part K] [--limit L]")
	checkPartitionRun(t, cli.RunCount, []string{"--distinct=1"}, "", "option --distinct takes no value")
	checkPartitionRun(t, cli.RunCount, []string{"--max-part"}, "", "--max-part needs a value")
	checkPartitionRun(t, cli.RunList, []string{"--limit"}, "", "--limit needs a value")
	checkPartitionRun(t, cli.RunRank, []string{"2", "3"}, "", "usage: youngkit rank SHAPE")

}

func TestAdaptersOptionValidationPrecedence(t *testing.T) {
	checkPartitionRun(t, cli.RunCount, []string{"--max-part=bad"}, "",
		"invalid --max-part \"bad\" (expected an integer from 1 to 5000)")
	checkPartitionRun(t, cli.RunCount, []string{"5", "--max-part=bad", "--distinct=1"}, "",
		"option --distinct takes no value")
	checkPartitionRun(t, cli.RunCount, []string{"5", "--distinct", "--max-part=bad"}, "",
		"invalid --max-part \"bad\" (expected an integer from 1 to 5000)")
	checkPartitionRun(t, cli.RunList, []string{"5", "--limit=0"}, "",
		"invalid --limit \"0\" (expected an integer from 1 to 1000)")
	checkPartitionRun(t, cli.RunList, []string{"5", "--max-part=201"}, "",
		"invalid --max-part \"201\" (expected an integer from 1 to 200)")
	checkPartitionRun(t, cli.RunRank, []string{"4,2,1", "--bad"}, "", "unknown option \"--bad\"")
}

func TestAdaptersNumericAndShapeValidation(t *testing.T) {
	checkPartitionRun(t, cli.RunCount, []string{"5001"}, "",
		"invalid n \"5001\" (expected an integer from 0 to 5000)")
	checkPartitionRun(t, cli.RunCount, []string{"007"}, "",
		"invalid n \"007\" (expected an integer from 0 to 5000)")
	checkPartitionRun(t, cli.RunRank, []string{"2,3"}, "",
		"parts of \"2,3\" must not increase (part 2 is 3, after 2)")
	checkPartitionRun(t, cli.RunRank, []string{"3,0"}, "",
		"invalid part \"0\" in \"3,0\" (expected 1 to 5000)")
	checkPartitionRun(t, cli.RunRank, []string{"400,1"}, "",
		"size 401 is above 400 (too large for rank)")
	checkPartitionRun(t, cli.RunUnrank, []string{"6", "11"}, "",
		"invalid rank \"11\" (expected an integer from 0 to 10)")
}

func TestSpecifiedStressCountsAndPositionalAccess(t *testing.T) {
	if got := count.Partitions(5000, 5000, false).String(); got != "169820168825442121851975101689306431361757683049829233322203824652329144349" {
		t.Fatalf("Partitions(5000, 5000, false) = %s", got)
	}
	if got := count.Partitions(5000, 5000, true).String(); got != "15988884521431077020247618131907553242282546626679512" {
		t.Fatalf("Partitions(5000, 5000, true) = %s", got)
	}
	if got := count.Generate(200, 200, 3); len(got) != 3 || got[0].String() != "200" || got[1].String() != "199,1" || got[2].String() != "198,2" {
		t.Fatalf("Generate(200, 200, 3) = %#v", got)
	}
	if got := count.Unrank(400, big.NewInt(6123456789012345678)).String(); got != "34,24,22,20,18,18,18,17,16,14,12,9,9,9,9,8,5,5,5,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1" {
		t.Fatalf("Unrank(400, request index) = %s", got)
	}
	if rank, _ := count.Rank(partitionShape(t, "34,24,22,20,18,18,18,17,16,14,12,9,9,9,9,8,5,5,5,4,4,4,4,4,4,4,4,4,4,4,3,3,3,3,3,3,3,3,3,3,3,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1")); rank.String() != "6123456789012345678" {
		t.Fatalf("Rank(size-400 literal) = %s", rank)
	}
}

func TestHooksLiteralExamples(t *testing.T) {
	tests := []struct {
		text string
		want [][]int
	}{
		{"0", [][]int{}},
		{"1", [][]int{{1}}},
		{"4", [][]int{{4, 3, 2, 1}}},
		{"1,1,1", [][]int{{3}, {2}, {1}}},
		{"3,2", [][]int{{4, 3, 1}, {2, 1}}},
		{"4,3,1", [][]int{{6, 4, 3, 1}, {4, 2, 1}, {1}}},
	}
	for _, test := range tests {
		got := count.Hooks(partitionShape(t, test.text))
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Hooks(%s) = %#v, want %#v", test.text, got, test.want)
		}
	}
}

func TestHooksEmptyShapeHasAllocatedZeroRows(t *testing.T) {
	got := count.Hooks(partitionShape(t, "0"))
	if got == nil || len(got) != 0 {
		t.Fatalf("Hooks(0) = %#v, want an allocated slice with zero rows", got)
	}
}

func TestHooksPreservesInputAndReturnsIndependentRows(t *testing.T) {
	input := shape.Shape{4, 3, 1}
	wantInput := shape.Shape{4, 3, 1}
	first := count.Hooks(input)
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("Hooks changed its input: got %v, want %v", input, wantInput)
	}
	first[0][0] = 99
	if first[1][0] != 4 || first[2][0] != 1 {
		t.Fatalf("mutating one hook row changed another: %#v", first)
	}
	second := count.Hooks(input)
	if !reflect.DeepEqual(second, [][]int{{6, 4, 3, 1}, {4, 2, 1}, {1}}) {
		t.Fatalf("mutating returned hooks changed later results: %#v", second)
	}
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("mutating returned hooks changed the input: %v", input)
	}
}

func TestSYTLiteralExactCounts(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"0", "1"},
		{"4", "1"},
		{"1,1,1", "1"},
		{"3,2", "5"},
		{"4,3,1", "70"},
		{"8,8,8,8,8", "231471904322784840"},
	}
	for _, test := range tests {
		got := count.SYT(partitionShape(t, test.text))
		if got.String() != test.want {
			t.Errorf("SYT(%s) = %s, want %s", test.text, got, test.want)
		}
	}
}

func TestSYTPreservesShapeAndReturnsFreshCounts(t *testing.T) {
	input := shape.Shape{4, 3, 1}
	first := count.SYT(input)
	second := count.SYT(input)
	if !reflect.DeepEqual(input, shape.Shape{4, 3, 1}) {
		t.Fatalf("SYT changed its input: %v", input)
	}
	if first == second {
		t.Fatal("SYT returned the same big.Int pointer twice")
	}
	first.SetInt64(0)
	if second.String() != "70" || count.SYT(input).String() != "70" {
		t.Fatalf("mutating one SYT result changed another result: second=%s", second)
	}
}

func TestSSYTLiteralExactCounts(t *testing.T) {
	tests := []struct {
		text    string
		entries int
		want    string
	}{
		{"0", 1, "1"},
		{"1", 3, "3"},
		{"3", 2, "4"},
		{"1,1,1", 2, "0"},
		{"1,1,1", 3, "1"},
		{"2,1", 2, "2"},
		{"4,3,1", 3, "15"},
	}
	for _, test := range tests {
		got := count.SSYT(partitionShape(t, test.text), test.entries)
		if got.String() != test.want {
			t.Errorf("SSYT(%s, %d) = %s, want %s", test.text, test.entries, got, test.want)
		}
	}
}

func TestSSYTPreservesShapeAndReturnsFreshCounts(t *testing.T) {
	input := shape.Shape{4, 3, 1}
	first := count.SSYT(input, 3)
	second := count.SSYT(input, 3)
	if !reflect.DeepEqual(input, shape.Shape{4, 3, 1}) {
		t.Fatalf("SSYT changed its input: %v", input)
	}
	if first == second {
		t.Fatal("SSYT returned the same big.Int pointer twice")
	}
	first.SetInt64(0)
	if second.String() != "15" || count.SSYT(input, 3).String() != "15" {
		t.Fatalf("mutating one SSYT result changed another result: second=%s", second)
	}
}

func TestRunHooksRequestExampleAndOmittedEntries(t *testing.T) {
	checkPartitionRun(t, cli.RunHooks, []string{"4,3,1", "--entries", "3"},
		"shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\nentries     3\nssyt        15\n", "")
	checkPartitionRun(t, cli.RunHooks, []string{"4,3,1"},
		"shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\n", "")
}

func TestRunHooksEmptyAndInsufficientEntries(t *testing.T) {
	checkPartitionRun(t, cli.RunHooks, []string{"0"},
		"shape       0\nsize        0\nsyt         1\n", "")
	checkPartitionRun(t, cli.RunHooks, []string{"0", "--entries=1"},
		"shape       0\nsize        0\nsyt         1\nentries     1\nssyt        1\n", "")
	checkPartitionRun(t, cli.RunHooks, []string{"1,1,1", "--entries=2"},
		"shape       1,1,1\nsize        3\nhook        3\nhook        2\nhook        1\nsyt         1\nentries     2\nssyt        0\n", "")
}

func TestRunHooksEntryRangeBoundaries(t *testing.T) {
	checkPartitionRun(t, cli.RunHooks, []string{"1", "--entries=1"},
		"shape       1\nsize        1\nhook        1\nsyt         1\nentries     1\nssyt        1\n", "")
	checkPartitionRun(t, cli.RunHooks, []string{"1", "--entries=1000"},
		"shape       1\nsize        1\nhook        1\nsyt         1\nentries     1000\nssyt        1000\n", "")
}

func TestRunHooksOptionSpellingsInterspersionAndLastValue(t *testing.T) {
	want := "shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\nentries     3\nssyt        15\n"
	tests := [][]string{
		{"--entries", "3", "4,3,1"},
		{"4,3,1", "--entries=3"},
		{"--entries=1", "4,3,1", "--entries", "3"},
		{"--entries=not-a-number", "--entries=3", "4,3,1"},
	}
	for _, args := range tests {
		checkPartitionRun(t, cli.RunHooks, args, want, "")
	}
	checkPartitionRun(t, cli.RunHooks, []string{"--entries=3", "4,3,1", "--entries=0"}, "",
		"invalid --entries \"0\" (expected an integer from 1 to 1000)")
}

func TestRunHooksOptionsAreValidatedBeforeArityAndShape(t *testing.T) {
	checkPartitionRun(t, cli.RunHooks, []string{"--entries=0"}, "",
		"invalid --entries \"0\" (expected an integer from 1 to 1000)")
	checkPartitionRun(t, cli.RunHooks, []string{"2,3", "--entries=1001"}, "",
		"invalid --entries \"1001\" (expected an integer from 1 to 1000)")
	checkPartitionRun(t, cli.RunHooks, []string{"--entries=0", "--unknown"}, "",
		"unknown option \"--unknown\"")
	checkPartitionRun(t, cli.RunHooks, []string{"2,3", "extra"}, "",
		"usage: youngkit hooks SHAPE [--entries K]")
}

func TestRunHooksCanonicalOptionDecimalErrors(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"+3", "invalid --entries \"+3\" (expected an integer from 1 to 1000)"},
		{"-1", "invalid --entries \"-1\" (expected an integer from 1 to 1000)"},
		{"03", "invalid --entries \"03\" (expected an integer from 1 to 1000)"},
		{"1.5", "invalid --entries \"1.5\" (expected an integer from 1 to 1000)"},
		{"", "invalid --entries \"\" (expected an integer from 1 to 1000)"},
		{"٣", "invalid --entries \"?\" (expected an integer from 1 to 1000)"},
		{"12345678901", "invalid --entries \"12345678901\" (expected an integer from 1 to 1000)"},
	}
	for _, test := range tests {
		checkPartitionRun(t, cli.RunHooks, []string{"1", "--entries=" + test.value}, "", test.want)
	}
}

func TestRunHooksValidationErrorsAndShapeStages(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{nil, "usage: youngkit hooks SHAPE [--entries K]"},
		{[]string{"0", "1"}, "usage: youngkit hooks SHAPE [--entries K]"},
		{[]string{"--entries"}, "--entries needs a value"},
		{[]string{"--bad"}, "unknown option \"--bad\""},
		{[]string{"03"}, "invalid shape \"03\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{[]string{"4,"}, "invalid shape \"4,\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{[]string{"3,0"}, "invalid part \"0\" in \"3,0\" (expected 1 to 5000)"},
		{[]string{"2,3"}, "parts of \"2,3\" must not increase (part 2 is 3, after 2)"},
		{[]string{"5001"}, "invalid part \"5001\" in \"5001\" (expected 1 to 5000)"},
		{[]string{"2001"}, "size 2001 is above 2000 (too large for hooks)"},
	}
	for _, test := range tests {
		checkPartitionRun(t, cli.RunHooks, test.args, "", test.want)
	}
}

func TestRunHooksRepeatedCallsAreDeterministic(t *testing.T) {
	args := []string{"4,3,1", "--entries=3"}
	want := "shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\nentries     3\nssyt        15\n"
	checkPartitionRun(t, cli.RunHooks, args, want, "")
	checkPartitionRun(t, cli.RunHooks, args, want, "")
}

const hooksRectangle40x50 = "50,50,50,50,50,50,50,50,50,50," +
	"50,50,50,50,50,50,50,50,50,50," +
	"50,50,50,50,50,50,50,50,50,50," +
	"50,50,50,50,50,50,50,50,50,50"

func TestRunHooksLargeRectangleExactCountsUnderBudget(t *testing.T) {
	args := []string{hooksRectangle40x50, "--entries", "1000"}
	io := &partitionTestIO{}
	started := time.Now()
	got, err := cli.RunHooks(args, io)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("RunHooks(40-by-50 rectangle) error = %v", err)
	}
	if elapsed >= 10*time.Second {
		t.Fatalf("RunHooks(40-by-50 rectangle) took %s, want under 10 seconds", elapsed)
	}
	if io.reads != 0 || io.outCalls != 0 || io.errCalls != 0 || io.outText != "" || io.errText != "" {
		t.Fatalf("RunHooks used IO: reads=%d out=%d/%q err=%d/%q", io.reads, io.outCalls, io.outText, io.errCalls, io.errText)
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 46 || lines[45] != "" {
		t.Fatalf("large rectangle output has %d lines, want 45 LF-terminated rows", len(lines)-1)
	}
	if lines[0] != "shape       "+hooksRectangle40x50 || lines[1] != "size        2000" {
		t.Fatalf("large rectangle header = %q / %q", lines[0], lines[1])
	}
	for row := 0; row < 40; row++ {
		fields := strings.Fields(strings.TrimPrefix(lines[row+2], "hook        "))
		if len(fields) != 50 {
			t.Fatalf("large rectangle hook row %d has %d cells, want 50", row+1, len(fields))
		}
	}
	firstRow := strings.Fields(strings.TrimPrefix(lines[2], "hook        "))
	middleRow := strings.Fields(strings.TrimPrefix(lines[22], "hook        "))
	lastRow := strings.Fields(strings.TrimPrefix(lines[41], "hook        "))
	if firstRow[0] != "89" || firstRow[49] != "40" ||
		middleRow[25] != "44" || lastRow[0] != "50" || lastRow[49] != "1" {
		t.Fatalf("large rectangle selected hooks = %s/%s, %s, %s/%s",
			firstRow[0], firstRow[49], middleRow[25], lastRow[0], lastRow[49])
	}

	if !strings.HasPrefix(lines[42], "syt         ") {
		t.Fatalf("large rectangle syt row = %q", lines[42])
	}
	syt := strings.TrimPrefix(lines[42], "syt         ")
	if len(syt) != 2529 {
		t.Fatalf("large rectangle syt has %d digits, want 2529", len(syt))
	}
	if !strings.HasPrefix(syt, "79132721192156282272839009225071") ||
		!strings.HasSuffix(syt, "55680924221440000000000000000000") {
		t.Fatalf("large rectangle syt samples do not match the literal expectations")
	}
	if lines[43] != "entries     1000" {
		t.Fatalf("large rectangle entries row = %q", lines[43])
	}
	if !strings.HasPrefix(lines[44], "ssyt        ") {
		t.Fatalf("large rectangle ssyt row = %q", lines[44])
	}
	ssyt := strings.TrimPrefix(lines[44], "ssyt        ")
	if len(ssyt) != 2798 {
		t.Fatalf("large rectangle ssyt has %d digits, want 2798", len(ssyt))
	}
	if !strings.HasPrefix(ssyt, "36553644007669828472796194589850") ||
		!strings.HasSuffix(ssyt, "00000000000000000000000000000000") {
		t.Fatalf("large rectangle ssyt samples do not match the literal expectations")
	}
}
