package cli_test

import (
	"testing"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/cli"
)

type dispatchTestIO struct {
	stdin     []byte
	readCalls int
	outCalls  []string
	errCalls  []string
}

func (io *dispatchTestIO) ReadStdin() []byte {
	io.readCalls++
	return io.stdin
}

func (io *dispatchTestIO) Out(text string) {
	io.outCalls = append(io.outCalls, text)
}

func (io *dispatchTestIO) Err(text string) {
	io.errCalls = append(io.errCalls, text)
}

func requireDispatchError(t *testing.T, args []string, stdin []byte, want string, wantReads int) {
	t.Helper()
	fake := &dispatchTestIO{stdin: stdin}
	if got := cli.Run(args, fake); got != 2 {
		t.Fatalf("Run(%q) status = %d, want 2; out=%q err=%q", args, got, fake.outCalls, fake.errCalls)
	}
	if len(fake.outCalls) != 0 {
		t.Fatalf("Run(%q) stdout calls = %#v, want none", args, fake.outCalls)
	}
	if len(fake.errCalls) != 1 || fake.errCalls[0] != "youngkit: "+want+"\n" {
		t.Fatalf("Run(%q) stderr calls = %#v, want one exact call %q", args, fake.errCalls, "youngkit: "+want+"\n")
	}
	if fake.readCalls != wantReads {
		t.Fatalf("Run(%q) stdin reads = %d, want %d", args, fake.readCalls, wantReads)
	}
}

func requireDispatchSuccess(t *testing.T, args []string, stdin []byte, want string, wantReads int) {
	t.Helper()
	fake := &dispatchTestIO{stdin: stdin}
	if got := cli.Run(args, fake); got != 0 {
		t.Fatalf("Run(%q) status = %d, want 0; out=%q err=%q", args, got, fake.outCalls, fake.errCalls)
	}
	if len(fake.outCalls) != 1 || fake.outCalls[0] != want {
		t.Fatalf("Run(%q) stdout calls = %#v, want one exact call %q", args, fake.outCalls, want)
	}
	if len(fake.errCalls) != 0 {
		t.Fatalf("Run(%q) stderr calls = %#v, want none", args, fake.errCalls)
	}
	if fake.readCalls != wantReads {
		t.Fatalf("Run(%q) stdin reads = %d, want %d", args, fake.readCalls, wantReads)
	}
}

var _ base.IO = (*dispatchTestIO)(nil)

// AC-002: no arguments and an unknown command report their exact diagnostics.
func TestDispatchMissingAndUnknownCommandDiagnostics(t *testing.T) {
	requireDispatchError(t, nil, nil,
		"missing command (count, list, rank, unrank, conjugate, hooks, compare, neighbors, rsk, unrsk, syt or core)", 0)
	requireDispatchError(t, []string{"frob"}, nil, "unknown command \"frob\"", 0)
}

// AC-002: unknown text with Unicode, controls or invalid bytes is sanitized.
func TestDispatchSanitizesUnknownCommandText(t *testing.T) {
	requireDispatchError(t, []string{"é"}, nil, "unknown command \"?\"", 0)
	requireDispatchError(t, []string{"bad\ncommand"}, nil, "unknown command \"bad?command\"", 0)
	requireDispatchError(t, []string{string([]byte{0xff, 0xfe})}, nil, "unknown command \"??\"", 0)
}

// AC-002: success returns status zero and makes one complete output call.
func TestDispatchSuccessWritesOneCompleteOutputCall(t *testing.T) {
	requireDispatchSuccess(t, []string{"count", "1"}, nil,
		"n           1\nmax-part    any\ndistinct    no\ncount       1\n", 0)
}

// AC-002: failure, including after earlier valid input, has no partial output.
func TestDispatchFailureAfterValidInputWritesNoPartialOutput(t *testing.T) {
	requireDispatchError(t, []string{"rsk", "1", "invalid"}, nil,
		"invalid entry \"invalid\" (expected an integer from 1 to 2)", 0)
}

func TestEveryCommandReportsItsExactUsageLine(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"count"}, "usage: youngkit count N [--max-part K] [--distinct]"},
		{[]string{"list"}, "usage: youngkit list N [--max-part K] [--limit L]"},
		{[]string{"rank"}, "usage: youngkit rank SHAPE"},
		{[]string{"unrank"}, "usage: youngkit unrank N RANK"},
		{[]string{"conjugate"}, "usage: youngkit conjugate SHAPE"},
		{[]string{"hooks"}, "usage: youngkit hooks SHAPE [--entries K]"},
		{[]string{"compare"}, "usage: youngkit compare A B"},
		{[]string{"neighbors"}, "usage: youngkit neighbors SHAPE"},
		{[]string{"rsk"}, "usage: youngkit rsk ENTRY... | youngkit rsk -"},
		{[]string{"unrsk"}, "usage: youngkit unrsk P Q"},
		{[]string{"syt"}, "usage: youngkit syt SHAPE [--index K] [--limit L]"},
		{[]string{"core"}, "usage: youngkit core K SHAPE"},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

// AC-004: --max-part=5 and --max-part 5; repeated flag; options interspersed with arguments.
func TestOptionsAcceptEqualsSeparateInterspersedAndRepeatedFlags(t *testing.T) {
	want := "n           5\nmax-part    5\ndistinct    yes\ncount       3\n"
	requireDispatchSuccess(t, []string{"count", "--max-part=5", "5", "--distinct"}, nil, want, 0)
	requireDispatchSuccess(t, []string{"count", "5", "--max-part", "5", "--distinct", "--distinct"}, nil, want, 0)
}

// AC-004: repeated option, including invalid earlier numeric value, uses the last value.
func TestOptionsLastNumericValueWinsAndDiscardsEarlierInvalidValue(t *testing.T) {
	requireDispatchSuccess(t, []string{"count", "6", "--max-part=bad", "--max-part", "3"}, nil,
		"n           6\nmax-part    3\ndistinct    no\ncount       7\n", 0)
	requireDispatchSuccess(t, []string{"count", "6", "--max-part=4", "--max-part=3"}, nil,
		"n           6\nmax-part    3\ndistinct    no\ncount       7\n", 0)
}

// AC-004: unknown option, --, --help, and a flag assignment are rejected.
func TestOptionsRejectUnknownTokensAndFlagAssignments(t *testing.T) {
	requireDispatchError(t, []string{"count", "5", "--"}, nil, "unknown option \"--\"", 0)
	requireDispatchError(t, []string{"count", "5", "--help"}, nil, "unknown option \"--help\"", 0)
	requireDispatchError(t, []string{"count", "5", "--odd\nname"}, nil, "unknown option \"--odd?name\"", 0)
	requireDispatchError(t, []string{"count", "5", "--distinct=1"}, nil, "option --distinct takes no value", 0)
	requireDispatchError(t, []string{"count", "5", "--max-part=bad", "--distinct=1"}, nil,
		"option --distinct takes no value", 0)
	requireDispatchError(t, []string{"count", "5", "--max-part=bad", "--unknown"}, nil,
		"unknown option \"--unknown\"", 0)
}

// AC-004: missing values and dash-prefixed values are consumed by numeric options.
func TestOptionsRejectMissingAndConsumeDashPrefixedValues(t *testing.T) {
	requireDispatchError(t, []string{"count", "5", "--max-part"}, nil, "--max-part needs a value", 0)
	requireDispatchError(t, []string{"count", "5", "--max-part", "-"}, nil,
		"invalid --max-part \"-\" (expected an integer from 1 to 5000)", 0)
	requireDispatchError(t, []string{"count", "5", "--max-part=-"}, nil,
		"invalid --max-part \"-\" (expected an integer from 1 to 5000)", 0)
}

func TestNumericOptionConsumesDashPrefixedOptionToken(t *testing.T) {
	fake := &dispatchTestIO{}
	args := []string{"count", "5", "--max-part", "--distinct"}
	if got := cli.Run(args, fake); got != 2 {
		t.Fatalf("Run(%q) status = %d, want 2", args, got)
	}
	if len(fake.outCalls) != 0 {
		t.Fatalf("Run(%q) stdout calls = %#v, want none", args, fake.outCalls)
	}
	if len(fake.errCalls) != 1 || fake.errCalls[0] != "youngkit: invalid --max-part \"--distinct\" (expected an integer from 1 to 5000)\n" {
		t.Fatalf("Run(%q) stderr calls = %#v, want the exact prefixed diagnostic", args, fake.errCalls)
	}
	if fake.readCalls != 0 {
		t.Fatalf("Run(%q) stdin reads = %d, want 0", args, fake.readCalls)
	}
}

func TestNumericOptionFollowedByDashPrefixedTokenHasExactStreams(t *testing.T) {
	requireDispatchError(t, []string{"count", "5", "--max-part", "--distinct"}, nil,
		`invalid --max-part "--distinct" (expected an integer from 1 to 5000)`, 0)
}

// AC-004: structural error plus invalid numeric option; option error plus bad arity; bad arity plus malformed positional text.
func TestOptionsAreCheckedBeforeArityAndPositionalParsing(t *testing.T) {
	requireDispatchError(t, []string{"count", "--max-part=0"}, nil,
		"invalid --max-part \"0\" (expected an integer from 1 to 5000)", 0)
	requireDispatchError(t, []string{"count", "--unknown"}, nil, "unknown option \"--unknown\"", 0)
	requireDispatchError(t, []string{"count", "bad", "extra"}, nil,
		"usage: youngkit count N [--max-part K] [--distinct]", 0)
	requireDispatchError(t, []string{"rank", "2,3", "extra"}, nil, "usage: youngkit rank SHAPE", 0)
}

// AC-004: multiple numeric errors are reported in specification order.
func TestOptionsNumericErrorsFollowSpecificationOrder(t *testing.T) {
	requireDispatchError(t, []string{"list", "5", "--limit=bad", "--max-part=bad"}, nil,
		"invalid --max-part \"bad\" (expected an integer from 1 to 200)", 0)
	requireDispatchError(t, []string{"list", "5", "--limit=0", "--max-part=0"}, nil,
		"invalid --max-part \"0\" (expected an integer from 1 to 200)", 0)
}

func TestOptionRangeBoundariesAndFallbacks(t *testing.T) {
	requireDispatchSuccess(t, []string{"count", "0", "--max-part=1"}, nil,
		"n           0\nmax-part    1\ndistinct    no\ncount       1\n", 0)
	requireDispatchSuccess(t, []string{"count", "1", "--max-part", "5000"}, nil,
		"n           1\nmax-part    5000\ndistinct    no\ncount       1\n", 0)
	requireDispatchSuccess(t, []string{"list", "0", "--max-part=1", "--limit=1"}, nil,
		"n           0\nmax-part    1\ncount       1\nshown       1\npart        0\n", 0)
	requireDispatchSuccess(t, []string{"list", "0", "--max-part", "200", "--limit", "1000"}, nil,
		"n           0\nmax-part    200\ncount       1\nshown       1\npart        0\n", 0)
	requireDispatchSuccess(t, []string{"hooks", "0", "--entries=1"}, nil,
		"shape       0\nsize        0\nsyt         1\nentries     1\nssyt        1\n", 0)
	requireDispatchSuccess(t, []string{"hooks", "0", "--entries", "1000"}, nil,
		"shape       0\nsize        0\nsyt         1\nentries     1000\nssyt        1\n", 0)
	requireDispatchSuccess(t, []string{"syt", "0", "--index=0", "--limit=50"}, nil,
		"shape       0\nsize        0\ncount       1\nindex       0\nshown       1\ntableau     0\n", 0)
}

func TestEveryNumericOptionRejectsValuesOutsideItsRange(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"count", "5", "--max-part=0"}, "invalid --max-part \"0\" (expected an integer from 1 to 5000)"},
		{[]string{"count", "5", "--max-part=5001"}, "invalid --max-part \"5001\" (expected an integer from 1 to 5000)"},
		{[]string{"list", "5", "--max-part=0"}, "invalid --max-part \"0\" (expected an integer from 1 to 200)"},
		{[]string{"list", "5", "--max-part=201"}, "invalid --max-part \"201\" (expected an integer from 1 to 200)"},
		{[]string{"list", "201"}, `invalid n "201" (expected an integer from 0 to 200)`},
		{[]string{"list", "5", "--limit=0"}, "invalid --limit \"0\" (expected an integer from 1 to 1000)"},
		{[]string{"list", "5", "--limit=1001"}, "invalid --limit \"1001\" (expected an integer from 1 to 1000)"},
		{[]string{"hooks", "0", "--entries=0"}, "invalid --entries \"0\" (expected an integer from 1 to 1000)"},
		{[]string{"hooks", "0", "--entries=1001"}, "invalid --entries \"1001\" (expected an integer from 1 to 1000)"},
		{[]string{"syt", "0", "--index=-1"}, "invalid --index \"-1\" (expected an integer from 0 to 9999999999)"},
		{[]string{"syt", "0", "--index=10000000000"}, "invalid --index \"10000000000\" (expected an integer from 0 to 9999999999)"},
		{[]string{"syt", "0", "--limit=0"}, "invalid --limit \"0\" (expected an integer from 1 to 50)"},
		{[]string{"syt", "0", "--limit=51"}, "invalid --limit \"51\" (expected an integer from 1 to 50)"},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

func TestIntegerGrammarAndPositionalRangesAreExact(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"count", "01"}, "invalid n \"01\" (expected an integer from 0 to 5000)"},
		{[]string{"count", "+1"}, "invalid n \"+1\" (expected an integer from 0 to 5000)"},
		{[]string{"count", "5001"}, "invalid n \"5001\" (expected an integer from 0 to 5000)"},
		{[]string{"count", "٣"}, "invalid n \"?\" (expected an integer from 0 to 5000)"},
		{[]string{"core", "1", "0"}, "invalid K \"1\" (expected an integer from 2 to 50)"},
		{[]string{"core", "51", "0"}, "invalid K \"51\" (expected an integer from 2 to 50)"},
		{[]string{"unrank", "401", "bad"}, "invalid n \"401\" (expected an integer from 0 to 400)"},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

func TestListAndSytUseTheirLiteralDefaultLimits(t *testing.T) {
	requireDispatchSuccess(t, []string{"list", "8"}, nil,
		"n           8\nmax-part    any\ncount       22\nshown       20\npart        8\npart        7,1\npart        6,2\npart        6,1,1\npart        5,3\npart        5,2,1\npart        5,1,1,1\npart        4,4\npart        4,3,1\npart        4,2,2\npart        4,2,1,1\npart        4,1,1,1,1\npart        3,3,2\npart        3,3,1,1\npart        3,2,2,1\npart        3,2,1,1,1\npart        3,1,1,1,1,1\npart        2,2,2,2\npart        2,2,2,1,1\npart        2,2,1,1,1,1\n", 0)
	requireDispatchSuccess(t, []string{"syt", "4,3"}, nil,
		"shape       4,3\nsize        7\ncount       14\nindex       0\nshown       10\ntableau     1,3,5,7/2,4,6\ntableau     1,2,5,7/3,4,6\ntableau     1,3,4,7/2,5,6\ntableau     1,2,4,7/3,5,6\ntableau     1,2,3,7/4,5,6\ntableau     1,3,5,6/2,4,7\ntableau     1,2,5,6/3,4,7\ntableau     1,3,4,6/2,5,7\ntableau     1,2,4,6/3,5,7\ntableau     1,2,3,6/4,5,7\n", 0)
}

func TestShapeParsingReportsSyntaxPartOrderAndSizeStages(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"rank", ""}, "invalid shape \"\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{[]string{"rank", "2,"}, "invalid shape \"2,\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{[]string{"rank", "5001,03"}, "invalid shape \"5001,03\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{[]string{"rank", "0,3"}, "invalid part \"0\" in \"0,3\" (expected 1 to 5000)"},
		{[]string{"rank", "3,0"}, `invalid part "0" in "3,0" (expected 1 to 5000)`},
		{[]string{"rank", "5001"}, "invalid part \"5001\" in \"5001\" (expected 1 to 5000)"},
		{[]string{"rank", "2,3"}, "parts of \"2,3\" must not increase (part 2 is 3, after 2)"},
		{[]string{"rank", "5000,1"}, "shape \"5000,1\" has size 5001 (above 5000)"},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

func TestCommandShapeLimitsApplyAfterShapeParsing(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"rank", "401"}, "size 401 is above 400 (too large for rank)"},
		{[]string{"rank", "400,1"}, "size 401 is above 400 (too large for rank)"},
		{[]string{"conjugate", "201"}, "size 201 is above 200 (too large for conjugate)"},
		{[]string{"hooks", "2001"}, "size 2001 is above 2000 (too large for hooks)"},
		{[]string{"syt", "201"}, "size 201 is above 200 (too large for syt)"},
		{[]string{"rank", "401,402"}, "parts of \"401,402\" must not increase (part 2 is 402, after 401)"},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

// AC-005: rank syntax, the 200-digit length boundary, and out-of-range values.
func TestRankChecksCanonicalSyntaxLengthAndRange(t *testing.T) {
	const rank200 = "9999999999" + "9999999999" + "9999999999" + "9999999999" + "9999999999" +
		"9999999999" + "9999999999" + "9999999999" + "9999999999" + "9999999999" +
		"9999999999" + "9999999999" + "9999999999" + "9999999999" + "9999999999" +
		"9999999999" + "9999999999" + "9999999999" + "9999999999" + "9999999999"
	if len(rank200) != 200 {
		t.Fatalf("200-digit input length = %d, want 200", len(rank200))
	}
	rank201 := rank200 + "9"
	// The dispatcher reports the same bounded maximum for canonical out-of-range and overlong ranks.
	requireDispatchError(t, []string{"unrank", "6", rank200}, nil,
		"invalid rank \""+rank200+"\" (expected an integer from 0 to 10)", 0)
	requireDispatchError(t, []string{"unrank", "6", rank201}, nil,
		"invalid rank \""+rank201+"\" (expected an integer from 0 to 10)", 0)
	if len(rank201) != 201 {
		t.Fatalf("201-digit input length = %d, want 201", len(rank201))
	}
	requireDispatchError(t, []string{"unrank", "6", "11"}, nil,
		"invalid rank \"11\" (expected an integer from 0 to 10)", 0)
	requireDispatchError(t, []string{"unrank", "6", "01"}, nil,
		"invalid rank \"01\" (expected an integer from 0 to 10)", 0)
}

func TestCompareRejectsDifferentSizesAndValidatesShapesLeftToRight(t *testing.T) {
	requireDispatchError(t, []string{"compare", "3", "2"}, nil, "the sizes differ (3 and 2)", 0)
	requireDispatchError(t, []string{"compare", "2,3", "0"}, nil,
		"parts of \"2,3\" must not increase (part 2 is 3, after 2)", 0)
}

func TestTableauSyntaxAndStandardnessDiagnosticsFollowStages(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"unrsk", "", "0"}, "invalid tableau \"\" (expected rows of numbers such as 1,3,4/2,5, or 0 for the empty tableau)"},
		{[]string{"unrsk", "1,2/3,4,5001", "0"}, "invalid tableau \"1,2/3,4,5001\" (expected rows of numbers such as 1,3,4/2,5, or 0 for the empty tableau)"},
		{[]string{"unrsk", "1,2/3,4,5", "0"}, "tableau \"1,2/3,4,5\" is not standard: row 2 is longer than row 1"},
		{[]string{"unrsk", "1,1", "0"}, "tableau \"1,1\" is not standard: the entries are not 1 to N each once"},
		{[]string{"unrsk", "2,1/3", "0"}, "tableau \"2,1/3\" is not standard: row 1 does not increase"},
		{[]string{"unrsk", "1,4/2,3", "0"}, `tableau "1,4/2,3" is not standard: column 2 does not increase`},
		{[]string{"unrsk", "1,2/3,4,4", "0"}, `tableau "1,2/3,4,4" is not standard: row 2 is longer than row 1`},
		{[]string{"unrsk", "2,2/3", "0"}, `tableau "2,2/3" is not standard: the entries are not 1 to N each once`},
		{[]string{"unrsk", "3,1/2,4", "0"}, `tableau "3,1/2,4" is not standard: row 1 does not increase`},
	}
	for _, test := range tests {
		requireDispatchError(t, test.args, nil, test.want, 0)
	}
}

func TestUnrskParsesPThenQAndRequiresEqualShapes(t *testing.T) {
	requireDispatchError(t, []string{"unrsk", "1,4/2,3", "1,2/3,4"}, nil,
		"tableau \"1,4/2,3\" is not standard: column 2 does not increase", 0)
	requireDispatchError(t, []string{"unrsk", "1,2,3/4", "1,2/3"}, nil,
		"the shapes differ (3,1 and 2,1)", 0)
}

func TestRSKValidatesEntryCountAndWordsInOrder(t *testing.T) {
	requireDispatchError(t, []string{"rsk", "2", "2"}, nil, "entry 2 is repeated", 0)
	requireDispatchError(t, []string{"rsk", "0", "2"}, nil,
		"invalid entry \"0\" (expected an integer from 1 to 2)", 0)
	requireDispatchError(t, []string{"rsk", "2", "2", "bad"}, nil, "entry 2 is repeated", 0)

	excessiveInput := []byte(`1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1
1`)
	requireDispatchError(t, []string{"rsk", "-"}, excessiveInput,
		"too many entries (5001 is above 5000)", 1)
}

func TestRSKReadsOnlySoleDashAndHandlesLiteralStdinSeparators(t *testing.T) {
	requireDispatchSuccess(t, []string{"rsk", "-"}, []byte("\xef\xbb\xbf  1\t2\r\n"),
		"perm        1 2\nshape       2\nlis         2\nlds         1\np           1 2\nq           1 2\n", 1)
	requireDispatchSuccess(t, []string{"rsk", "1", "2"}, []byte("ignored"),
		"perm        1 2\nshape       2\nlis         2\nlds         1\np           1 2\nq           1 2\n", 0)
	requireDispatchError(t, []string{"rsk", "-"}, []byte(" \t\r\n"), "<stdin>: no input", 1)
	requireDispatchError(t, []string{"rsk", "-"}, []byte{}, "<stdin>: no input", 1)
	requireDispatchError(t, []string{"rsk", "-"}, []byte{0x31, 0x20, 0xff, 0x0a},
		"<stdin>: input is not valid UTF-8", 1)
	requireDispatchError(t, []string{"rsk", "-", "1"}, []byte("ignored"),
		"invalid entry \"-\" (expected an integer from 1 to 2)", 0)
}

func TestSytRejectsOutOfRangeIndicesAfterShapeValidation(t *testing.T) {
	requireDispatchError(t, []string{"syt", "3,2", "--index", "5"}, nil,
		"index 5 is not below 5 (the number of tableaux of \"3,2\")", 0)
	requireDispatchError(t, []string{"syt", "0", "--index=1"}, nil,
		"index 1 is not below 1 (the number of tableaux of \"0\")", 0)
	requireDispatchError(t, []string{"syt", "201,0"}, nil,
		"invalid part \"0\" in \"201,0\" (expected 1 to 5000)", 0)
}
