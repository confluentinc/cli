package results

import (
	"sync"

	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type MaterializedStatementResultsIterator struct {
	isTableMode bool
	iterator    *types.ListElement[types.StatementResultRow]
	lock        sync.RWMutex
}

func (i *MaterializedStatementResultsIterator) HasReachedEnd() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()

	return i.iterator == nil
}

func (i *MaterializedStatementResultsIterator) GetNext() *types.StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Next()
	return row
}

func (i *MaterializedStatementResultsIterator) GetPrev() *types.StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Prev()
	return row
}

func (i *MaterializedStatementResultsIterator) Value() *types.StatementResultRow {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.iterator == nil || i.iterator.Value() == nil {
		return nil
	}

	row := i.iterator.Value()
	if !i.isTableMode {
		operationField := types.AtomicStatementResultField{
			Type:  "VARCHAR",
			Value: row.Operation.String(),
		}
		row.Fields = append([]types.StatementResultField{operationField}, row.Fields...)
	}
	return row
}

func (i *MaterializedStatementResultsIterator) Move(stepsToMove int) *types.StatementResultRow {
	for !i.HasReachedEnd() && stepsToMove != 0 {
		if stepsToMove < 0 {
			i.GetPrev()
			stepsToMove++
		} else {
			i.GetNext()
			stepsToMove--
		}
	}
	return i.Value()
}

type MaterializedStatementResults struct {
	isTableMode bool
	maxCapacity int
	headers     []string
	changelog   types.LinkedList[types.StatementResultRow]
	table       types.LinkedList[types.StatementResultRow]
	cache       map[string]*types.ListElement[types.StatementResultRow]
	lock        sync.RWMutex
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

func (s *MaterializedStatementResults) Iterator(startFromBack bool) MaterializedStatementResultsIterator {
	s.lock.RLock()
	defer s.lock.RUnlock()

	list := s.table
	if !s.isTableMode {
		list = s.changelog
	}

	iterator := list.Front()
	if startFromBack {
		iterator = list.Back()
	}

	return MaterializedStatementResultsIterator{
		isTableMode: s.isTableMode,
		iterator:    iterator,
	}
}

func (s *MaterializedStatementResults) cleanup() {
	if s.changelog.Len() > s.maxCapacity {
		removedRow := s.changelog.RemoveFront()
		removedRowKey := removedRow.GetRowKey()

		listPtr, ok := s.cache[removedRowKey]
		if ok {
			s.table.Remove(listPtr)
			delete(s.cache, removedRowKey)
		}
	}
}

func (s *MaterializedStatementResults) Append(rows ...types.StatementResultRow) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	allValuesInserted := true
	for _, row := range rows {
		if len(row.Fields) != len(s.headers) {
			allValuesInserted = false
			continue
		}
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
		s.cleanup()
	}
	return allValuesInserted
}

func (s *MaterializedStatementResults) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.isTableMode {
		return s.table.Len()
	}
	return s.changelog.Len()
}

func (s *MaterializedStatementResults) SetTableMode(isTableMode bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.isTableMode = isTableMode
}

func (s *MaterializedStatementResults) IsTableMode() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.isTableMode
}

func (s *MaterializedStatementResults) GetHeaders() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.isTableMode {
		return s.headers
	}
	return append([]string{"Operation"}, s.headers...)
}

func (s *MaterializedStatementResults) ForEach(f func(rowIdx int, row *types.StatementResultRow)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	iterator := s.Iterator(false)
	rowIdx := 0
	for !iterator.HasReachedEnd() {
		f(rowIdx, iterator.GetNext())
		rowIdx++
	}
}

func (s *MaterializedStatementResults) GetMaxWidthPerColum() []int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	columnWidths := make([]int, len(s.GetHeaders()))
	for colIdx, column := range s.GetHeaders() {
		columnWidths[colIdx] = lo.Max([]int{len(column), columnWidths[colIdx]})
	}

	s.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		for colIdx, field := range row.Fields {
			columnWidths[colIdx] = lo.Max([]int{len(field.ToString()), columnWidths[colIdx]})
		}
	})
	return columnWidths
}
