package output

import "github.com/confluentinc/go-printer"

type HumanListWriter struct {
	outputFormat Format
	data         [][]string
	listFields   []string
	listLabels   []string
}

func (o *HumanListWriter) AddElement(e interface{}) {
	o.data = append(o.data, printer.ToRow(e, o.listFields))
}


func (o *HumanListWriter) Out() error {
	printer.RenderCollectionTable(o.data, o.listLabels)
	return nil
}

func (o *HumanListWriter) GetOutputFormat() Format {
	return o.outputFormat
}
