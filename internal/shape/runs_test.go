package shape

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	"github.com/Gulliby-development/youngkit/internal/base"
)

const runsSyntaxMessage = " (expected parts such as 5,3,1, or 0 for the empty shape)"

func TestParseRunsExpand(t *testing.T) {
	for _, test := range []struct {
		text string
		want Shape
	}{
		{text: "3x2", want: Shape{3, 3}},
		{text: "4,3x2,1", want: Shape{4, 3, 3, 1}},
		{text: "2x3", want: Shape{2, 2, 2}},
		{text: "3x1", want: Shape{3}},
		{text: "2,1x1", want: Shape{2, 1}},
		{text: "2x2,2,1x2", want: Shape{2, 2, 2, 1, 1}},
	} {
		got, err := Parse(test.text)
		if err != nil {
			t.Fatalf("Parse(%q): %v", test.text, err)
		}
		requireShapeEqual(t, got, test.want)
	}

	// Round trip: a run text and its spelled-out text parse to the same shape.
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))
		n := 1 + rng.Intn(10)
		parts := make([]int, n)
		for i := range parts {
			parts[i] = 1 + rng.Intn(20)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(parts)))
		var runText, spelled []string
		for _, part := range parts {
			count := 1 + rng.Intn(20)
			runText = append(runText, strconv.Itoa(part)+"x"+strconv.Itoa(count))
			for j := 0; j < count; j++ {
				spelled = append(spelled, strconv.Itoa(part))
			}
		}
		spelledText := strings.Join(spelled, ",")
		fromRuns, err := Parse(strings.Join(runText, ","))
		if err != nil {
			t.Logf("runs %v: %v", runText, err)
			return false
		}
		fromSpelled, err := Parse(spelledText)
		if err != nil {
			t.Logf("spelled %q: %v", spelledText, err)
			return false
		}
		if len(fromRuns) != len(fromSpelled) {
			return false
		}
		for i := range fromRuns {
			if fromRuns[i] != fromSpelled[i] {
				return false
			}
		}
		return fromRuns.String() == spelledText
	}
	config := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(20240607))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func TestParseRunsAtSizeLimit(t *testing.T) {
	got, err := Parse("2500x2")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, got, Shape{2500, 2500})
}

func TestParseRunsRejectsMalformedTokens(t *testing.T) {
	for _, text := range []string{"3x", "x2", "x", "3xx", "3x2x2", "3x02", "3X2", "3x+2", "3x1.5", "3x2,"} {
		_, err := Parse(text)
		// A malformed PARTxCOUNT token yields a *base.Error whose Msg is the syntax message.
		var baseErr *base.Error
		if !errors.As(err, &baseErr) {
			t.Fatalf("Parse(%q): want *base.Error, got %v", text, err)
		}
		if want := "invalid shape " + strconv.Quote(text) + runsSyntaxMessage; baseErr.Msg != want {
			t.Fatalf("Parse(%q) Msg = %q, want %q", text, baseErr.Msg, want)
		}
		requireShapeError(t, err, "invalid shape "+strconv.Quote(text)+runsSyntaxMessage)
	}
}

func TestParseRunsRejectsOutOfRange(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "0x2", want: `invalid part "0" in "0x2" (expected 1 to 5000)`},
		{text: "7000x2", want: `invalid part "7000" in "7000x2" (expected 1 to 5000)`},
		{text: "3x0", want: `invalid count "0" in "3x0" (expected 1 to 5000)`},
		{text: "2x5001", want: `invalid count "5001" in "2x5001" (expected 1 to 5000)`},
		{text: "0x0", want: `invalid part "0" in "0x0" (expected 1 to 5000)`},
		{text: "3x0,0x2", want: `invalid count "0" in "3x0,0x2" (expected 1 to 5000)`},
	} {
		_, err := Parse(test.text)
		requireShapeError(t, err, test.want)
	}
}

func TestParseRunsSyntaxPrecedesRange(t *testing.T) {
	_, err := Parse("3x0,y")
	requireShapeError(t, err, `invalid shape "3x0,y"`+runsSyntaxMessage)
}

func TestParseRunsEmptyShape(t *testing.T) {
	got, err := Parse("0")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, got, Shape{})
	_, err = Parse("0x1")
	requireShapeError(t, err, `invalid part "0" in "0x1" (expected 1 to 5000)`)
}

func TestParseRunsOrderingPosition(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "3x2,4", want: `parts of "3x2,4" must not increase (part 3 is 4, after 3)`},
		{text: "2,3x2", want: `parts of "2,3x2" must not increase (part 2 is 3, after 2)`},
		{text: "4x2,3x3,4", want: `parts of "4x2,3x3,4" must not increase (part 6 is 4, after 3)`},
		{text: "1x5000,2", want: `parts of "1x5000,2" must not increase (part 5001 is 2, after 1)`},
	} {
		_, err := Parse(test.text)
		requireShapeError(t, err, test.want)
	}
}

func TestParseRunsSizeBeforeExpanding(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "5000x2", want: `shape "5000x2" has size 10000 (above 5000)`},
		{text: "2500x2,1", want: `shape "2500x2,1" has size 5001 (above 5000)`},
		{text: "5000x5000", want: `shape "5000x5000" has size 25000000 (above 5000)`},
	} {
		_, err := Parse(test.text)
		requireShapeError(t, err, test.want)
	}
}
