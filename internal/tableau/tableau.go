package tableau

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/shape"
)

// Tableau is a standard Young tableau represented by its rows.
type Tableau [][]int

// Parse parses a standard Young tableau in row notation.
func Parse(text string) (Tableau, error) {
	if text == "0" {
		return Tableau{}, nil
	}

	rowTexts := strings.Split(text, "/")
	result := make(Tableau, len(rowTexts))
	total := 0

	// Syntax and entry ranges are checked over the whole tableau first.
	for rowIndex, rowText := range rowTexts {
		entries := strings.Split(rowText, ",")
		row := make([]int, len(entries))
		for columnIndex, entryText := range entries {
			entry, err := base.ParseInt("entry", entryText, 1, 5000)
			if err != nil {
				return nil, invalidTableau(text)
			}
			row[columnIndex] = int(entry)
		}
		result[rowIndex] = row
		total += len(row)
	}

	// Row lengths precede all standardness checks on entries.
	for rowIndex := 1; rowIndex < len(result); rowIndex++ {
		if len(result[rowIndex]) > len(result[rowIndex-1]) {
			return nil, nonstandardTableau(text,
				"row "+strconv.Itoa(rowIndex+1)+" is longer than row "+strconv.Itoa(rowIndex))
		}
	}

	// Entries are individually bounded by 5000, so a larger tableau cannot
	// contain exactly the integers 1 through its size.
	if total > 5000 {
		return nil, nonstandardTableau(text, "the entries are not 1 to N each once")
	}
	seen := make([]bool, total+1)
	for _, row := range result {
		for _, entry := range row {
			if entry > total || seen[entry] {
				return nil, nonstandardTableau(text, "the entries are not 1 to N each once")
			}
			seen[entry] = true
		}
	}
	for entry := 1; entry <= total; entry++ {
		if !seen[entry] {
			return nil, nonstandardTableau(text, "the entries are not 1 to N each once")
		}
	}

	for rowIndex, row := range result {
		for columnIndex := 1; columnIndex < len(row); columnIndex++ {
			if row[columnIndex] <= row[columnIndex-1] {
				return nil, nonstandardTableau(text,
					"row "+strconv.Itoa(rowIndex+1)+" does not increase")
			}
		}
	}

	maxColumns := 0
	if len(result) > 0 {
		maxColumns = len(result[0])
	}
	for columnIndex := 0; columnIndex < maxColumns; columnIndex++ {
		for rowIndex := 1; rowIndex < len(result); rowIndex++ {
			if columnIndex >= len(result[rowIndex]) {
				break
			}
			if result[rowIndex][columnIndex] <= result[rowIndex-1][columnIndex] {
				return nil, nonstandardTableau(text,
					"column "+strconv.Itoa(columnIndex+1)+" does not increase")
			}
		}
	}

	return result, nil
}

// String returns the canonical row notation, or 0 for the empty tableau.
func (t Tableau) String() string {
	if len(t) == 0 {
		return "0"
	}

	var text strings.Builder
	for rowIndex, row := range t {
		if rowIndex > 0 {
			text.WriteByte('/')
		}
		for columnIndex, entry := range row {
			if columnIndex > 0 {
				text.WriteByte(',')
			}
			text.WriteString(strconv.Itoa(entry))
		}
	}
	return text.String()
}

// Shape returns a fresh shape containing the tableau's row lengths.
func (t Tableau) Shape() shape.Shape {
	result := make(shape.Shape, len(t))
	for rowIndex, row := range t {
		result[rowIndex] = len(row)
	}
	return result
}

func invalidTableau(text string) error {
	return base.Fail("invalid tableau " + base.Show(text) +
		" (expected rows of numbers such as 1,3,4/2,5, or 0 for the empty tableau)")
}

func nonstandardTableau(text, reason string) error {
	return base.Fail("tableau " + base.Show(text) + " is not standard: " + reason)
}
