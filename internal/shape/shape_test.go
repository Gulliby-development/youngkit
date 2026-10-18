package shape

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Gulliby-development/youngkit/internal/base"
)

func requireShapeError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want %q", want)
	}
	var got *base.Error
	if !errors.As(err, &got) {
		t.Fatalf("error type = %T, want *base.Error", err)
	}
	if got.Msg != want || err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func requireShapeEqual(t *testing.T, got, want Shape) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("shape = %#v, want %#v", got, want)
	}
}

func TestParseEmptyShape(t *testing.T) {
	got, err := Parse("0")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, got, Shape{})
}

func TestParseTypicalShape(t *testing.T) {
	got, err := Parse("5,3,3,1")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, got, Shape{5, 3, 3, 1})
}

func TestParseMaximumPart(t *testing.T) {
	got, err := Parse("5000")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, got, Shape{5000})
}

// Required syntax examples: empty text, 1,,1, 1 followed by comma, 03, +3, -1, 1.5, whitespace, non-ASCII digits, overlong decimal tokens.
func TestParseRejectsMalformedTokens(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "", want: "invalid shape \"\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "1,,1", want: "invalid shape \"1,,1\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "1,", want: "invalid shape \"1,\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "03", want: "invalid shape \"03\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "+3", want: "invalid shape \"+3\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "-1", want: "invalid shape \"-1\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "1.5", want: "invalid shape \"1.5\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: " 1", want: "invalid shape \"?1\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "1 ", want: "invalid shape \"1?\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "٣", want: "invalid shape \"?\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
	} {
		t.Run(test.text, func(t *testing.T) {
			_, err := Parse(test.text)
			requireShapeError(t, err, test.want)
		})
	}
}

func TestParseRejectsOverlongDecimalTokens(t *testing.T) {
	for _, test := range []struct {
		text string
		want string
	}{
		{text: "10000000000", want: "invalid shape \"10000000000\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
		{text: "99999999999999999999", want: "invalid shape \"99999999999999999999\" (expected parts such as 5,3,1, or 0 for the empty shape)"},
	} {
		_, err := Parse(test.text)
		requireShapeError(t, err, test.want)
	}
}

func TestParseRejectsLiteralTextToken(t *testing.T) {
	_, err := Parse("TEXT")
	requireShapeError(t, err, "invalid shape \"TEXT\" (expected parts such as 5,3,1, or 0 for the empty shape)")
}

func TestParseRejectsZeroPart(t *testing.T) {
	_, err := Parse("3,0")
	requireShapeError(t, err, "invalid part \"0\" in \"3,0\" (expected 1 to 5000)")
}

func TestParseRejectsPartAboveMaximum(t *testing.T) {
	_, err := Parse("9999999999")
	requireShapeError(t, err, "invalid part \"9999999999\" in \"9999999999\" (expected 1 to 5000)")
}

func TestParseRejectsFirstAdjacentIncrease(t *testing.T) {
	_, err := Parse("2,3")
	requireShapeError(t, err, "parts of \"2,3\" must not increase (part 2 is 3, after 2)")
}

func TestParseSyntaxPrecedesPartRange(t *testing.T) {
	_, err := Parse("0,x")
	requireShapeError(t, err, "invalid shape \"0,x\" (expected parts such as 5,3,1, or 0 for the empty shape)")
}

func TestParsePartRangePrecedesOrdering(t *testing.T) {
	_, err := Parse("2,3,0")
	requireShapeError(t, err, "invalid part \"0\" in \"2,3,0\" (expected 1 to 5000)")
}

func TestParseOrderingPrecedesTotalSize(t *testing.T) {
	_, err := Parse("1,5000")
	requireShapeError(t, err, "parts of \"1,5000\" must not increase (part 2 is 5000, after 1)")
}

func TestParseReportsFirstIncrease(t *testing.T) {
	_, err := Parse("3,4,5")
	requireShapeError(t, err, "parts of \"3,4,5\" must not increase (part 2 is 4, after 3)")
}

func TestParseRejectsTotalAboveMaximum(t *testing.T) {
	_, err := Parse("5000,1")
	requireShapeError(t, err, "shape \"5000,1\" has size 5001 (above 5000)")
}

func TestParseSanitizesUnicodeInError(t *testing.T) {
	_, err := Parse("é")
	requireShapeError(t, err, "invalid shape \"?\" (expected parts such as 5,3,1, or 0 for the empty shape)")
}

func TestParseSanitizesInvalidBytesInError(t *testing.T) {
	text := string([]byte{'1', ',', 0xff})
	_, err := Parse(text)
	requireShapeError(t, err, "invalid shape \"1,?\" (expected parts such as 5,3,1, or 0 for the empty shape)")
}

func TestShapeStringEmpty(t *testing.T) {
	if got := (Shape{}).String(); got != "0" {
		t.Fatalf("String() = %q, want %q", got, "0")
	}
}

func TestShapeStringParts(t *testing.T) {
	if got := (Shape{5, 3, 3, 1}).String(); got != "5,3,3,1" {
		t.Fatalf("String() = %q, want %q", got, "5,3,3,1")
	}
}

func TestShapeSizeEmpty(t *testing.T) {
	if got := (Shape{}).Size(); got != 0 {
		t.Fatalf("Size() = %d, want 0", got)
	}
}

func TestShapeSizeSumsParts(t *testing.T) {
	if got := (Shape{5, 3, 3, 1}).Size(); got != 12 {
		t.Fatalf("Size() = %d, want 12", got)
	}
}

func TestShapeLimitSucceedsBelowBound(t *testing.T) {
	if err := (Shape{200}).Limit(201, "conjugate"); err != nil {
		t.Fatalf("Limit() = %v, want nil", err)
	}
}

func TestShapeLimitSucceedsAtBound(t *testing.T) {
	if err := (Shape{200}).Limit(200, "conjugate"); err != nil {
		t.Fatalf("Limit() = %v, want nil", err)
	}
}

func TestShapeLimitReportsExactExcess(t *testing.T) {
	err := (Shape{200, 1}).Limit(200, "conjugate")
	requireShapeError(t, err, "size 201 is above 200 (too large for conjugate)")
}

func TestShapeMethodsSupportSize5001(t *testing.T) {
	got := Shape{5000, 1}
	if got.Size() != 5001 || got.String() != "5000,1" {
		t.Fatalf("size/string = %d/%q, want 5001/\"5000,1\"", got.Size(), got.String())
	}
}

func TestConjugateEmptyShape(t *testing.T) {
	requireShapeEqual(t, Conjugate(Shape{}), Shape{})
}

func TestConjugateSymmetricShape(t *testing.T) {
	input, err := Parse("3,2,1")
	if err != nil {
		t.Fatal(err)
	}
	requireShapeEqual(t, Conjugate(input), Shape{3, 2, 1})
}

func TestConjugateAsymmetricShape(t *testing.T) {
	requireShapeEqual(t, Conjugate(Shape{5, 3, 3, 1}), Shape{4, 3, 3, 1, 1})
}

func TestConjugateMaximumWidthShape(t *testing.T) {
	got := Conjugate(Shape{5000})
	if len(got) != 5000 {
		t.Fatalf("conjugate length = %d, want 5000", len(got))
	}
	if got[0] != 1 || got[2499] != 1 || got[4999] != 1 {
		t.Fatalf("selected conjugate parts = %d, %d, %d; want 1, 1, 1", got[0], got[2499], got[4999])
	}
}

func TestDiagramEmptyShape(t *testing.T) {
	got := Diagram(Shape{})
	if !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("Diagram() = %#v, want no rows", got)
	}
}

func TestDiagramTypicalShape(t *testing.T) {
	got := Diagram(Shape{5, 3, 3, 1})
	want := []string{"#####", "###", "###", "#"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Diagram() = %#v, want %#v", got, want)
	}
}

func TestDiagramMaximumWidthRow(t *testing.T) {
	got := Diagram(Shape{5000})
	if len(got) != 1 || len(got[0]) != 5000 {
		t.Fatalf("diagram dimensions = %d rows, row length %d; want 1, 5000", len(got), len(got[0]))
	}
	if got[0][0] != '#' || got[0][2499] != '#' || got[0][4999] != '#' || strings.Trim(got[0], "#") != "" {
		t.Fatal("diagram row does not contain 5000 # characters")
	}
}

func TestCompareEmptyShapes(t *testing.T) {
	got := Compare(Shape{}, Shape{})
	want := Dominance{Relation: "equal", SumsA: []int{}, SumsB: []int{}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %#v, want %#v", got, want)
	}
}

func TestCompareEqualShapes(t *testing.T) {
	input, err := Parse("2,1")
	if err != nil {
		t.Fatal(err)
	}
	got := Compare(input, input)
	want := Dominance{Relation: "equal", SumsA: []int{2, 3}, SumsB: []int{2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %#v, want %#v", got, want)
	}
}

func TestCompareADominatesB(t *testing.T) {
	a, err := Parse("3,2,1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse("2,2,2")
	if err != nil {
		t.Fatal(err)
	}
	got := Compare(a, b)
	want := Dominance{Relation: "a dominates b", SumsA: []int{3, 5, 6}, SumsB: []int{2, 4, 6}, AAhead: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %#v, want %#v", got, want)
	}
}

func TestCompareBDominatesA(t *testing.T) {
	got := Compare(Shape{2, 2, 2}, Shape{3, 2, 1})
	want := Dominance{Relation: "b dominates a", SumsA: []int{2, 4, 6}, SumsB: []int{3, 5, 6}, BAhead: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %#v, want %#v", got, want)
	}
}

func TestCompareIncomparableAndFirstLeads(t *testing.T) {
	got := Compare(Shape{3, 3}, Shape{4, 1, 1})
	want := Dominance{Relation: "incomparable", SumsA: []int{3, 6, 6}, SumsB: []int{4, 5, 6}, AAhead: 2, BAhead: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %#v, want %#v", got, want)
	}
}

func TestComparePadsPrefixSumsToLongerShape(t *testing.T) {
	got := Compare(Shape{3, 3}, Shape{4, 1, 1})
	if len(got.SumsA) != 3 || got.SumsA[2] != 6 || got.SumsB[2] != 6 {
		t.Fatalf("padded sums = %#v and %#v, want three entries ending in 6", got.SumsA, got.SumsB)
	}
}

func TestAddsAndRemovesKeepRowOrder(t *testing.T) {
	input, err := Parse("4,2,2")
	if err != nil {
		t.Fatal(err)
	}
	adds := Adds(input)
	wantAdds := []Shape{{5, 2, 2}, {4, 3, 2}, {4, 2, 2, 1}}
	if !reflect.DeepEqual(adds, wantAdds) {
		t.Fatalf("Adds() = %#v, want %#v", adds, wantAdds)
	}
	removes := Removes(input)
	wantRemoves := []Shape{{3, 2, 2}, {4, 2, 1}}
	if !reflect.DeepEqual(removes, wantRemoves) {
		t.Fatalf("Removes() = %#v, want %#v", removes, wantRemoves)
	}
}

func TestAddsAndRemovesEmptyShape(t *testing.T) {
	adds := Adds(Shape{})
	if !reflect.DeepEqual(adds, []Shape{{1}}) {
		t.Fatalf("Adds(empty) = %#v, want []Shape{{1}}", adds)
	}
	if got := Removes(Shape{}); !reflect.DeepEqual(got, []Shape{}) {
		t.Fatalf("Removes(empty) = %#v, want no removals", got)
	}
}

func TestAddsAndRemovesSingleCell(t *testing.T) {
	if got := Adds(Shape{1}); !reflect.DeepEqual(got, []Shape{{2}, {1, 1}}) {
		t.Fatalf("Adds(1) = %#v, want [2; 1,1]", got)
	}
	if got := Removes(Shape{1}); !reflect.DeepEqual(got, []Shape{{}}) {
		t.Fatalf("Removes(1) = %#v, want the empty shape", got)
	}
}

func TestNeighborsPreserveSize5001Additions(t *testing.T) {
	adds := Adds(Shape{5000})
	wantAdds := []Shape{{5001}, {5000, 1}}
	if !reflect.DeepEqual(adds, wantAdds) {
		t.Fatalf("Adds(5000) = %#v, want %#v", adds, wantAdds)
	}
	removes := Removes(Shape{5000})
	if !reflect.DeepEqual(removes, []Shape{{4999}}) {
		t.Fatalf("Removes(5000) = %#v, want []Shape{{4999}}", removes)
	}
	if adds[0].Size() != 5001 || adds[1].Size() != 5001 {
		t.Fatalf("addition sizes = %d and %d, want 5001", adds[0].Size(), adds[1].Size())
	}
}

func TestNeighborSlicesAreIndependent(t *testing.T) {
	input := Shape{4, 2, 2}
	adds := Adds(input)
	adds[0][0] = 99
	adds[1][1] = 99
	if !reflect.DeepEqual(input, Shape{4, 2, 2}) {
		t.Fatalf("input changed through neighbor: %#v", input)
	}
	if adds[0][1] != 2 || adds[1][0] != 4 {
		t.Fatalf("neighbor slices share storage: %#v", adds)
	}
}
