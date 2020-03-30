package output

import (
	"io"
	"sort"

	"github.com/confluentinc/go-printer"
)

type HumanListWriter struct {
	outputFormat Format
	data         [][]string
	listFields   []string
	listLabels   []string
	writer       io.Writer
}

func (o *HumanListWriter) AddElement(e interface{}) {
	o.data = append(o.data, printer.ToRow(e, o.listFields))
}

func (o *HumanListWriter) Out() error {
	printer.RenderCollectionTableOut(o.data, o.listLabels, o.writer)
	return nil
}

func (o *HumanListWriter) GetOutputFormat() Format {
	return o.outputFormat
}

func (o *HumanListWriter) StableSort() {
	sort.Slice(o.data, func(i, j int) bool {
		for x := range o.data[i] {
			if o.data[i][x] != o.data[j][x] {
				return o.data[i][x] < o.data[j][x]
			}
		}
		return false
	})
}
