package results

func distributeLeftoverCharacters(columnWidths []int, truncatedColumns map[int]int, leftoverCharacters int) {
	// distribute extra characters greedy-style (each column takes as much as it needs if possible)
	for leftoverCharacters > 0 {
		for truncatedColIdx, missingChars := range truncatedColumns {
			if missingChars > leftoverCharacters {
				columnWidths[truncatedColIdx] += leftoverCharacters
				truncatedColumns[truncatedColIdx] -= leftoverCharacters
				return
			}

			columnWidths[truncatedColIdx] += missingChars
			truncatedColumns[truncatedColIdx] = 0
			leftoverCharacters -= missingChars
		}
	}
}

func sum(values []int) int {
	valuesSum := 0
	for _, col := range values {
		valuesSum += col
	}
	return valuesSum
}

func GetTruncatedColumnWidths(columnWidths []int, maxCharacters int) []int {
	numColumns := len(columnWidths)
	if numColumns == 0 || sum(columnWidths) <= maxCharacters {
		return columnWidths
	}

	charsPerColumn := maxCharacters / numColumns
	leftoverChars := maxCharacters % numColumns

	truncatedCols := map[int]int{}
	newColumns := make([]int, numColumns)
	for i, col := range columnWidths {
		if col > charsPerColumn {
			newColumns[i] = charsPerColumn
			truncatedCols[i] = col - charsPerColumn
			continue
		}

		newColumns[i] = col
		leftoverChars += charsPerColumn - col
	}

	distributeLeftoverCharacters(newColumns, truncatedCols, leftoverChars)

	return newColumns
}
