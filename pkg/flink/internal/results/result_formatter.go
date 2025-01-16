package results

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/samber/lo"
)

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

func GetTruncatedColumnWidths(columnWidths []int, maxCharacters int) []int {
	numColumns := len(columnWidths)
	if numColumns == 0 || lo.Sum(columnWidths) <= maxCharacters {
		return columnWidths
	}

	charsPerColumn := maxCharacters / numColumns
	leftoverChars := maxCharacters % numColumns

	var truncatedColumns []truncatedColumn // slice of struct instead of map because we need to preserve the order
	truncatedColumnWidths := make([]int, numColumns)
	for i, col := range columnWidths {
		// We should not need to check only col > charsPerColumn but rather first search for line breaks
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
	lines := strings.Split(str, "\n")
	truncatedLines := make([]string, 0)
	// We need to manually format each line here so that we append empty spaces to lines that doesn't fill the whole row, for it to look like "error box component" and not always have a different format depending on the error message.
	for _, line := range lines {
		// Using runewidth.Truncate instead of substring is important because it handles multi-byte characters
		// (e.g. chinese characters, emojis etc.). It uses runewidth.StringWidth internally to calculate the width
		// of the string which is important in a terminal environment where any miscalculation causes poor formatting
		truncatedLines = append(truncatedLines, runewidth.Truncate(line, maxCharCountToDisplay, "..."))
	}
	return strings.Join(truncatedLines, "\n")
}
