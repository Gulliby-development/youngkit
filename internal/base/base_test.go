package base

import (
	"errors"
	"reflect"
	"testing"
)

type fakeIO struct {
	stdin []byte
	reads int
	out   []string
	err   []string
}

func (f *fakeIO) ReadStdin() []byte {
	f.reads++
	return f.stdin
}

func (f *fakeIO) Out(text string) {
	f.out = append(f.out, text)
}

func (f *fakeIO) Err(text string) {
	f.err = append(f.err, text)
}

var _ IO = StdIO{}

func requireError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want %q", want)
	}
	if got := err.Error(); got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestErrorReturnsMessageUnchanged(t *testing.T) {
	err := (&Error{Msg: "exact\nmessage"}).Error()
	if err != "exact\nmessage" {
		t.Fatalf("Error() = %q", err)
	}
}

func TestFailReturnsConcreteError(t *testing.T) {
	got := Fail("")
	var baseErr *Error
	if !errors.As(got, &baseErr) {
		t.Fatalf("Fail returned %T, want *Error", got)
	}
	if baseErr.Msg != "" || got.Error() != "" {
		t.Fatalf("empty Fail result = %#v, %q", baseErr, got.Error())
	}
}

func TestFailPreservesNonemptyMessage(t *testing.T) {
	got := Fail("unchanged")
	if got.Error() != "unchanged" {
		t.Fatalf("Fail error = %q", got.Error())
	}
}

func TestShowEmpty(t *testing.T) {
	if got := Show(""); got != "\"\"" {
		t.Fatalf("Show empty = %q", got)
	}
}

func TestShowPrintableASCII(t *testing.T) {
	if got := Show("!Az~"); got != "\"!Az~\"" {
		t.Fatalf("Show printable = %q", got)
	}
}

func TestShowPreservesQuotesAndBackslashes(t *testing.T) {
	if got := Show("\"\\"); got != "\"\"\\\"" {
		t.Fatalf("Show quotes = %q", got)
	}
}

func TestShowReplacesSpaceUnicodeAndControl(t *testing.T) {
	if got := Show("A é\n"); got != "\"A???\"" {
		t.Fatalf("Show mixed = %q", got)
	}
}

func TestShowReplacesEachInvalidByte(t *testing.T) {
	if got := Show(string([]byte{0xff, 0xfe})); got != "\"??\"" {
		t.Fatalf("Show invalid bytes = %q", got)
	}
}

func TestLinePadsShortKey(t *testing.T) {
	if got := Line("n", "10"); got != "n           10" {
		t.Fatalf("Line = %q", got)
	}
}

func TestLineTenCharacterKey(t *testing.T) {
	if got := Line("abcdefghij", "x"); got != "abcdefghij  x" {
		t.Fatalf("Line = %q", got)
	}
}

func TestLineEmptyKeyHasNoNewline(t *testing.T) {
	if got := Line("", "value"); got != "            value" {
		t.Fatalf("Line = %q", got)
	}
}

func TestJoinNil(t *testing.T) {
	if got := Join(nil); got != "" {
		t.Fatalf("Join(nil) = %q", got)
	}
}

func TestJoinEmpty(t *testing.T) {
	if got := Join([]int{}); got != "" {
		t.Fatalf("Join(empty) = %q", got)
	}
}

func TestJoinSignedNumbers(t *testing.T) {
	if got := Join([]int{3, 0, -2}); got != "3 0 -2" {
		t.Fatalf("Join = %q", got)
	}
}

func TestParseIntAcceptsZero(t *testing.T) {
	got, err := ParseInt("n", "0", 0, 0)
	if err != nil || got != 0 {
		t.Fatalf("ParseInt = %d, %v", got, err)
	}
}

func TestParseIntAcceptsInclusiveBounds(t *testing.T) {
	for _, test := range []struct {
		text string
		lo   int64
		hi   int64
		want int64
	}{
		{text: "3", lo: 3, hi: 9, want: 3},
		{text: "9", lo: 3, hi: 9, want: 9},
	} {
		got, err := ParseInt("value", test.text, test.lo, test.hi)
		if err != nil || got != test.want {
			t.Errorf("ParseInt(%q) = %d, %v; want %d, nil", test.text, got, err, test.want)
		}
	}
}

func TestParseIntAcceptsTenDigitMaximum(t *testing.T) {
	got, err := ParseInt("count", "9999999999", 0, 9999999999)
	if err != nil || got != 9999999999 {
		t.Fatalf("ParseInt = %d, %v", got, err)
	}
}

func TestParseIntRejectsNoncanonicalForms(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "", want: "invalid N \"\" (expected an integer from 0 to 10)"},
		{text: "-1", want: "invalid N \"-1\" (expected an integer from 0 to 10)"},
		{text: "+3", want: "invalid N \"+3\" (expected an integer from 0 to 10)"},
		{text: "007", want: "invalid N \"007\" (expected an integer from 0 to 10)"},
		{text: "1.5", want: "invalid N \"1.5\" (expected an integer from 0 to 10)"},
		{text: " 3", want: "invalid N \"?3\" (expected an integer from 0 to 10)"},
		{text: "٣", want: "invalid N \"?\" (expected an integer from 0 to 10)"},
	} {
		_, err := ParseInt("N", test.text, 0, 10)
		requireError(t, err, test.want)
	}
}

func TestParseIntRejectsElevenDigits(t *testing.T) {
	_, err := ParseInt("n", "10000000000", 0, 9999999999)
	requireError(t, err, "invalid n \"10000000000\" (expected an integer from 0 to 9999999999)")
}

func TestParseIntRejectsTwentyDigitsWithoutOverflow(t *testing.T) {
	_, err := ParseInt("n", "99999999999999999999", 0, 9999999999)
	requireError(t, err, "invalid n \"99999999999999999999\" (expected an integer from 0 to 9999999999)")
}

func TestParseIntRejectsBelowLowerBound(t *testing.T) {
	_, err := ParseInt("n", "2", 3, 9)
	requireError(t, err, "invalid n \"2\" (expected an integer from 3 to 9)")
}

func TestParseIntRejectsAboveUpperBound(t *testing.T) {
	_, err := ParseInt("n", "10", 3, 9)
	requireError(t, err, "invalid n \"10\" (expected an integer from 3 to 9)")
}

func TestParseIntSanitizesInvalidText(t *testing.T) {
	_, err := ParseInt("n", string([]byte{0xff}), 0, 9)
	requireError(t, err, "invalid n \"?\" (expected an integer from 0 to 9)")
}

func TestOptionsParseEqualsAndSeparateValues(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 1, Hi: 9, Fallback: 2}}
	got, err := Parse([]string{"first", "--max=5", "-", "--max", "6", "last"}, specs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Values["--max"] != 6 || !reflect.DeepEqual(got.Args, []string{"first", "-", "last"}) {
		t.Fatalf("Parse = %#v", got)
	}
}

func TestOptionsPopulateEveryFallback(t *testing.T) {
	specs := []Spec{
		{Name: "--max", Lo: 1, Hi: 9, Fallback: 4},
		{Name: "--flag", Fallback: 7, Flag: true},
	}
	got, err := Parse(nil, specs)
	if err != nil {
		t.Fatal(err)
	}
	if got.Values["--max"] != 4 || got.Values["--flag"] != 7 || len(got.Values) != 2 {
		t.Fatalf("Values = %#v", got.Values)
	}
}

func TestOptionsSetPresentFlagToOne(t *testing.T) {
	specs := []Spec{{Name: "--flag", Fallback: 8, Flag: true}}
	got, err := Parse([]string{"--flag"}, specs)
	if err != nil || got.Values["--flag"] != 1 {
		t.Fatalf("Parse = %#v, %v", got, err)
	}
}

func TestOptionsLastRepeatedValueWins(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9}}
	got, err := Parse([]string{"--max=2", "--max=8"}, specs)
	if err != nil || got.Values["--max"] != 8 {
		t.Fatalf("Parse = %#v, %v", got, err)
	}
}

func TestOptionsValidLastValueOverwritesInvalidEarlierValue(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9}}
	got, err := Parse([]string{"--max=bad", "--max=3"}, specs)
	if err != nil || got.Values["--max"] != 3 {
		t.Fatalf("Parse = %#v, %v", got, err)
	}
}

func TestOptionsInvalidLastValueIsValidated(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9}}
	_, err := Parse([]string{"--max=3", "--max=bad"}, specs)
	requireError(t, err, "invalid --max \"bad\" (expected an integer from 0 to 9)")
}

func TestOptionsRetainLargeInt64Value(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9999999999}}
	got, err := Parse([]string{"--max=9999999999"}, specs)
	if err != nil || got.Values["--max"] != 9999999999 {
		t.Fatalf("Parse = %#v, %v", got, err)
	}
}

func TestOptionsRejectBareDoubleDash(t *testing.T) {
	_, err := Parse([]string{"--"}, nil)
	requireError(t, err, "unknown option \"--\"")
}

func TestOptionsRejectHelpWhenUnspecified(t *testing.T) {
	_, err := Parse([]string{"--help"}, nil)
	requireError(t, err, "unknown option \"--help\"")
}

func TestOptionsSanitizeWholeUnknownToken(t *testing.T) {
	_, err := Parse([]string{"--mystery\n"}, nil)
	requireError(t, err, "unknown option \"--mystery?\"")
}

func TestOptionsRejectFlagAssignmentIncludingEmptyValue(t *testing.T) {
	specs := []Spec{{Name: "--flag", Flag: true}}
	for _, arg := range []string{"--flag=1", "--flag="} {
		_, err := Parse([]string{arg}, specs)
		requireError(t, err, "option --flag takes no value")
	}
}

func TestOptionsRejectMissingFinalValue(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9}}
	_, err := Parse([]string{"--max"}, specs)
	requireError(t, err, "--max needs a value")
}

func TestOptionsSplitAtFirstEquals(t *testing.T) {
	specs := []Spec{{Name: "--text", Lo: 0, Hi: 9}}
	_, err := Parse([]string{"--text=3=4"}, specs)
	requireError(t, err, "invalid --text \"3=4\" (expected an integer from 0 to 9)")
}

func TestOptionsConsumeOptionLikeValue(t *testing.T) {
	specs := []Spec{
		{Name: "--first", Lo: 0, Hi: 9},
		{Name: "--second", Lo: 0, Hi: 9},
	}
	_, err := Parse([]string{"--first", "--second"}, specs)
	requireError(t, err, "invalid --first \"--second\" (expected an integer from 0 to 9)")
}

func TestOptionsReturnFirstStructuralError(t *testing.T) {
	specs := []Spec{{Name: "--flag", Flag: true}}
	_, err := Parse([]string{"--flag=1", "--unknown"}, specs)
	requireError(t, err, "option --flag takes no value")
}

func TestOptionsUnknownStructuralErrorPrecedesLaterFlagError(t *testing.T) {
	specs := []Spec{{Name: "--flag", Flag: true}}
	_, err := Parse([]string{"--unknown", "--flag=1"}, specs)
	requireError(t, err, "unknown option \"--unknown\"")
}

func TestOptionsStructuralErrorPrecedesNumericValidation(t *testing.T) {
	specs := []Spec{
		{Name: "--max", Lo: 0, Hi: 9},
		{Name: "--flag", Flag: true},
	}
	_, err := Parse([]string{"--max=bad", "--flag=1"}, specs)
	requireError(t, err, "option --flag takes no value")
}

func TestOptionsNumericErrorsFollowSpecOrder(t *testing.T) {
	specs := []Spec{
		{Name: "--first", Lo: 0, Hi: 9},
		{Name: "--second", Lo: 0, Hi: 9},
	}
	_, err := Parse([]string{"--second=bad", "--first=also-bad"}, specs)
	requireError(t, err, "invalid --first \"also-bad\" (expected an integer from 0 to 9)")
}

func TestOptionsEmptyEqualsValueUsesIntegerError(t *testing.T) {
	specs := []Spec{{Name: "--max", Lo: 0, Hi: 9}}
	_, err := Parse([]string{"--max="}, specs)
	requireError(t, err, "invalid --max \"\" (expected an integer from 0 to 9)")
}

func TestWordsPassThroughEmptyArgsWithoutReading(t *testing.T) {
	input := &fakeIO{stdin: []byte("unused")}
	got, err := Words(input, nil)
	if err != nil || len(got) != 0 || input.reads != 0 {
		t.Fatalf("Words = %#v, %v; reads = %d", got, err, input.reads)
	}
}

func TestWordsPassThroughDashAmongOtherArgs(t *testing.T) {
	args := []string{"-", "other"}
	input := &fakeIO{stdin: []byte("ignored")}
	got, err := Words(input, args)
	if err != nil || !reflect.DeepEqual(got, args) || input.reads != 0 {
		t.Fatalf("Words = %#v, %v; reads = %d", got, err, input.reads)
	}
}

func TestWordsReadExactlyOneDash(t *testing.T) {
	input := &fakeIO{stdin: []byte("one two")}
	got, err := Words(input, []string{"-"})
	if err != nil || !reflect.DeepEqual(got, []string{"one", "two"}) || input.reads != 1 {
		t.Fatalf("Words = %#v, %v; reads = %d", got, err, input.reads)
	}
	if len(input.out) != 0 || len(input.err) != 0 {
		t.Fatalf("Words wrote streams: out=%#v err=%#v", input.out, input.err)
	}
}

func TestWordsSplitOnlySupportedSeparatorsAndCRLF(t *testing.T) {
	input := &fakeIO{stdin: []byte(" \tone\r\ntwo\rthree\nfour\t")}
	got, err := Words(input, []string{"-"})
	if err != nil || !reflect.DeepEqual(got, []string{"one", "two", "three", "four"}) {
		t.Fatalf("Words = %#v, %v", got, err)
	}
}

func TestWordsRemoveExactlyOneLeadingBOM(t *testing.T) {
	input := &fakeIO{stdin: []byte("\xef\xbb\xbfword")}
	got, err := Words(input, []string{"-"})
	if err != nil || !reflect.DeepEqual(got, []string{"word"}) {
		t.Fatalf("Words = %#v, %v", got, err)
	}
}

func TestWordsKeepSecondAndInternalBOMAndOtherWhitespace(t *testing.T) {
	input := &fakeIO{stdin: []byte("\xef\xbb\xbf\xef\xbb\xbfword\xc2\xa0next mid\xef\xbb\xbf")}
	got, err := Words(input, []string{"-"})
	want := []string{"\xef\xbb\xbfword\xc2\xa0next", "mid\xef\xbb\xbf"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Words = %#v, %v; want %#v", got, err, want)
	}
}

func TestWordsKeepNULInsideWord(t *testing.T) {
	input := &fakeIO{stdin: []byte("one\x00two")}
	got, err := Words(input, []string{"-"})
	if err != nil || !reflect.DeepEqual(got, []string{"one\x00two"}) {
		t.Fatalf("Words = %#v, %v", got, err)
	}
}

func TestWordsRejectEmptyAndSeparatorOnlyInput(t *testing.T) {
	for _, data := range [][]byte{nil, {}, []byte(" \t\r\n"), []byte("\xef\xbb\xbf")} {
		input := &fakeIO{stdin: data}
		_, err := Words(input, []string{"-"})
		requireError(t, err, "<stdin>: no input")
	}
}

func TestWordsAcceptASCIIAndValidUnicode(t *testing.T) {
	input := &fakeIO{stdin: []byte("ascii \xc3\xa9 \xf0\x9f\x98\x80")}
	got, err := Words(input, []string{"-"})
	want := []string{"ascii", "é", "😀"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Words = %#v, %v; want %#v", got, err, want)
	}
}

func TestDecodeStdinRejectsMalformedUTF8Bytes(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
	}{
		{name: "worked bytes", data: []byte{0x31, 0x20, 0xff, 0x0a}},
		{name: "isolated continuation", data: []byte{0x80}},
		{name: "truncated sequence", data: []byte{0xc3}},
		{name: "overlong encoding", data: []byte{0xc0, 0xaf}},
		{name: "surrogate encoding", data: []byte{0xed, 0xa0, 0x80}},
		{name: "above Unicode maximum", data: []byte{0xf4, 0x90, 0x80, 0x80}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeStdin(test.data)
			requireError(t, err, "<stdin>: input is not valid UTF-8")
			input := &fakeIO{stdin: test.data}
			_, err = Words(input, []string{"-"})
			requireError(t, err, "<stdin>: input is not valid UTF-8")
			if input.reads != 1 {
				t.Fatalf("stdin reads = %d, want 1", input.reads)
			}
		})
	}
}

func TestWordsRejectInvalidUTF8BeforeEmptyInput(t *testing.T) {
	input := &fakeIO{stdin: []byte{0xff, ' ', '\n'}}
	_, err := Words(input, []string{"-"})
	requireError(t, err, "<stdin>: input is not valid UTF-8")
}

func TestWordsPreserveRealOutputChannelState(t *testing.T) {
	input := &fakeIO{stdin: []byte("word")}
	_, err := Words(input, []string{"-"})
	if err != nil {
		t.Fatal(err)
	}
	if input.out != nil || input.err != nil {
		t.Fatalf("Words output = %#v, %#v", input.out, input.err)
	}
}
