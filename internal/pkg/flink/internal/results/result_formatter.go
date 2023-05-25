package results

type truncatedColumn struct {
	idx            int
	truncatedChars int
}

func distributeLeftoverCharacters(columnWidths []int, truncatedColumns []truncatedColumn, leftoverCharacters int) {
	// distribute extra characters greedily (each column takes as much as it needs if possible)
	for leftoverCharacters > 0 {
		for _, col := range truncatedColumns {
			if col.truncatedChars > leftoverCharacters {
				columnWidths[col.idx] += leftoverCharacters
				return
			}

			columnWidths[col.idx] += col.truncatedChars
			leftoverCharacters -= col.truncatedChars
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

	var truncatedColumns []truncatedColumn // slice of struct instead of map because we need to preserve the order
	truncatedColumnWidths := make([]int, numColumns)
	for i, col := range columnWidths {
		if col > charsPerColumn {
			truncatedColumnWidths[i] = charsPerColumn
			truncatedColumns = append(truncatedColumns, truncatedColumn{
				idx:            i,
				truncatedChars: col - charsPerColumn,
			})
			continue
		}

		truncatedColumnWidths[i] = col
		leftoverChars += charsPerColumn - col
	}

	distributeLeftoverCharacters(truncatedColumnWidths, truncatedColumns, leftoverChars)

	return truncatedColumnWidths
}

func TruncateString(str string, maxCharCountToDisplay int) string {
	if len(str) > maxCharCountToDisplay {
		if maxCharCountToDisplay <= 3 {
			return "..."
		}
		return str[:maxCharCountToDisplay-3] + "..."
	}
	return str
}
