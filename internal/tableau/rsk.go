package tableau

import "sort"

// RSK applies Robinson-Schensted row insertion to a permutation.
func RSK(perm []int) (Tableau, Tableau) {
	p := make(Tableau, 0)
	q := make(Tableau, 0)

	for time, entry := range perm {
		carried := entry
		for rowIndex := 0; ; rowIndex++ {
			if rowIndex == len(p) {
				p = append(p, []int{})
				q = append(q, []int{})
			}

			row := p[rowIndex]
			columnIndex := sort.Search(len(row), func(i int) bool {
				return row[i] > carried
			})
			if columnIndex == len(row) {
				p[rowIndex] = append(row, carried)
				q[rowIndex] = append(q[rowIndex], time+1)
				break
			}

			row[columnIndex], carried = carried, row[columnIndex]
		}
	}

	return p, q
}

// Inverse reverses Robinson-Schensted insertion without mutating its inputs.
func Inverse(p, q Tableau) []int {
	pWork := cloneTableau(p)
	qWork := cloneTableau(q)
	size := 0
	for _, row := range q {
		size += len(row)
	}

	labelRows := make([]int, size+1)
	labelColumns := make([]int, size+1)
	for rowIndex, row := range q {
		for columnIndex, label := range row {
			labelRows[label] = rowIndex
			labelColumns[label] = columnIndex
		}
	}

	perm := make([]int, size)
	for label := size; label > 0; label-- {
		rowIndex := labelRows[label]
		columnIndex := labelColumns[label]
		qWork[rowIndex] = qWork[rowIndex][:columnIndex]

		carried := pWork[rowIndex][columnIndex]
		pWork[rowIndex] = pWork[rowIndex][:columnIndex]
		for rowIndex--; rowIndex >= 0; rowIndex-- {
			row := pWork[rowIndex]
			columnIndex = sort.Search(len(row), func(i int) bool {
				return row[i] >= carried
			}) - 1
			row[columnIndex], carried = carried, row[columnIndex]
		}
		perm[label-1] = carried
	}

	return perm
}

func cloneTableau(t Tableau) Tableau {
	copyOf := make(Tableau, len(t))
	for rowIndex, row := range t {
		copyOf[rowIndex] = append([]int(nil), row...)
	}
	return copyOf
}
