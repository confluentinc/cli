package results

import "github.com/confluentinc/flink-sql-client/pkg/types"

type materializedStatementResultsIterator struct {
	isTableMode bool
	iterator    *types.ListElement[types.StatementResultRow]
}

func (i *materializedStatementResultsIterator) HasNext() bool {
	return i.iterator != nil
}

func (i *materializedStatementResultsIterator) GetNext() *types.StatementResultRow {
	row := i.iterator.Value()
	if !i.isTableMode {
		operationField := types.AtomicStatementResultField{
			Type:  "VARCHAR",
			Value: row.Operation.String(),
		}
		row.Fields = append([]types.StatementResultField{operationField}, row.Fields...)
	}
	i.iterator = i.iterator.Next()
	return row
}

type MaterializedStatementResults struct {
	isTableMode bool
	maxCapacity int
	headers     []string
	changelog   types.LinkedList[types.StatementResultRow]
	table       types.LinkedList[types.StatementResultRow]
	cache       map[string]*types.ListElement[types.StatementResultRow]
}

func NewMaterializedStatementResults(headers []string, maxCapacity int) MaterializedStatementResults {
	return MaterializedStatementResults{
		isTableMode: true,
		maxCapacity: maxCapacity,
		headers:     headers,
		changelog:   types.NewLinkedList[types.StatementResultRow](),
		table:       types.NewLinkedList[types.StatementResultRow](),
		cache:       map[string]*types.ListElement[types.StatementResultRow]{},
	}
}

func (s *MaterializedStatementResults) Iterator() materializedStatementResultsIterator {
	iterator := s.table.Front()
	if !s.isTableMode {
		iterator = s.changelog.Front()
	}
	return materializedStatementResultsIterator{
		isTableMode: s.isTableMode,
		iterator:    iterator,
	}
}

func (s *MaterializedStatementResults) Append(row types.StatementResultRow) {
	s.changelog.PushBack(row)

	rowKey := row.GetRowKey()
	if row.Operation.IsInsertOperation() {
		listPtr := s.table.PushBack(row)
		s.cache[rowKey] = listPtr
	} else {
		listPtr, ok := s.cache[rowKey]
		if ok {
			s.table.Remove(listPtr)
			delete(s.cache, rowKey)
		}
	}

	// if we are now over the capacity we need to remove some records
	if s.changelog.Len() > s.maxCapacity {
		removedRow := s.changelog.RemoveFront()

		removedRowKey := removedRow.GetRowKey()
		// we only care about insert events, delete events have already been handled on arrival
		if row.Operation.IsInsertOperation() {
			listPtr, ok := s.cache[removedRowKey]
			if ok {
				s.table.Remove(listPtr)
				delete(s.cache, removedRowKey)
			}
		}
	}
}

func (s *MaterializedStatementResults) AppendAll(rows []types.StatementResultRow) {
	for _, row := range rows {
		s.Append(row)
	}
}

func (s *MaterializedStatementResults) Size() int {
	if s.isTableMode {
		return s.table.Len()
	}
	return s.changelog.Len()
}

func (s *MaterializedStatementResults) SetTableMode(isTableMode bool) {
	s.isTableMode = isTableMode
}

func (s *MaterializedStatementResults) IsTableMode() bool {
	return s.isTableMode
}

func (s *MaterializedStatementResults) GetHeaders() []string {
	if s.isTableMode {
		return s.headers
	}
	return append([]string{"Operation"}, s.headers...)
}
