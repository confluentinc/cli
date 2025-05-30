package types

import (
	"sync"

	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type MaterializedStatementResultsIterator struct {
	isTableMode bool
	iterator    *ListElement[StatementResultRow]
	lock        sync.RWMutex
}

func (i *MaterializedStatementResultsIterator) HasReachedEnd() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()

	return i.iterator == nil
}

func (i *MaterializedStatementResultsIterator) GetNext() *StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Next()
	return row
}

func (i *MaterializedStatementResultsIterator) GetPrev() *StatementResultRow {
	row := i.Value()

	i.lock.Lock()
	defer i.lock.Unlock()

	i.iterator = i.iterator.Prev()
	return row
}

func (i *MaterializedStatementResultsIterator) Value() *StatementResultRow {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.iterator == nil || i.iterator.Value() == nil {
		return nil
	}

	row := i.iterator.Value()
	if !i.isTableMode {
		operationField := AtomicStatementResultField{
			Type:  "VARCHAR",
			Value: row.Operation.String(),
		}
		row.Fields = append([]StatementResultField{operationField}, row.Fields...)
	}
	return row
}

func (i *MaterializedStatementResultsIterator) Move(stepsToMove int) *StatementResultRow {
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
	changelog   LinkedList[StatementResultRow]
	table       LinkedList[StatementResultRow]
	// Cache is a map of row keys to slices of pointers to the list elements in the table.
	// We have a slice of pointers to the list elements because there can be multiple rows with the same key since
	// primary keys are NOT ENFORCED in flink and clients can insert rows for the same key.
	cache         map[string][]*ListElement[StatementResultRow]
	lock          sync.RWMutex
	upsertColumns *[]int32
}

func NewMaterializedStatementResults(headers []string, maxCapacity int, upsertColumns *[]int32) MaterializedStatementResults {
	return MaterializedStatementResults{
		isTableMode:   true,
		maxCapacity:   maxCapacity,
		headers:       headers,
		changelog:     NewLinkedList[StatementResultRow](),
		table:         NewLinkedList[StatementResultRow](),
		cache:         map[string][]*ListElement[StatementResultRow]{},
		upsertColumns: upsertColumns,
	}
}

func (s *MaterializedStatementResults) GetTable() LinkedList[StatementResultRow] {
	return s.table
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
		s.changelog.RemoveFront()
	}

	if s.table.Len() > s.maxCapacity {
		removedElement := s.table.RemoveFront()
		if removedElement != nil {
			removedRowKey, _ := removedElement.GetRowKey(s.upsertColumns)

			if listPtrs, ok := s.cache[removedRowKey]; ok && len(listPtrs) > 0 {
				firstCachedElement := listPtrs[0]
				if firstCachedElement.Value() == removedElement {
					remainingListPtrs := listPtrs[1:]
					if len(remainingListPtrs) == 0 {
						delete(s.cache, removedRowKey)
					} else {
						s.cache[removedRowKey] = remainingListPtrs
					}
				} else {
					// The oldest removed the table should be the first element in the cached slice for that key.
					// This is a cache inconsistency and bug in the logic.
					log.CliLogger.Warnf("Materializing results error: Cache inconsistency during cleanup:"+
						" removed table element for key '%s' did not match the first cached element."+
						" Table element: %p, Cached element: %p", removedRowKey, removedElement, firstCachedElement)
				}
			}
		}
	}
}

func (s *MaterializedStatementResults) Append(rows ...StatementResultRow) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	allValuesInserted := true
	for _, row := range rows {
		if len(row.Fields) != len(s.headers) {
			allValuesInserted = false
			continue
		}
		s.changelog.PushBack(row)

		if rowKey, hasUpsertColumns := row.GetRowKey(s.upsertColumns); hasUpsertColumns {
			s.processRowWithUpsertColumns(row, rowKey)
		} else {
			s.processRowWithoutUpsertColumns(row, rowKey)
		}

		s.cleanup()
	}
	return allValuesInserted
}

func (s *MaterializedStatementResults) processRowWithUpsertColumns(row StatementResultRow, rowKey string) {
	switch row.Operation {
	case Insert:
		listPtr := s.table.PushBack(row)
		s.cache[rowKey] = append(s.cache[rowKey], listPtr)
	case UpdateAfter:
		listPtrs, ok := s.cache[rowKey]
		if ok && len(listPtrs) > 0 {
			lastPtr := listPtrs[len(listPtrs)-1]
			lastPtr.element.Value = row
		} else {
			listPtr := s.table.PushBack(row)
			s.cache[rowKey] = []*ListElement[StatementResultRow]{listPtr}
		}
	case UpdateBefore:
		return
	case Delete:
		listPtrs, ok := s.cache[rowKey]
		if ok && len(listPtrs) > 0 {
			lastPtr := listPtrs[len(listPtrs)-1]
			s.table.Remove(lastPtr)
			s.cache[rowKey] = listPtrs[:len(listPtrs)-1]
			if len(s.cache[rowKey]) == 0 {
				delete(s.cache, rowKey)
			}
		}
	default:
		log.CliLogger.Warnf("unknown operation received: %f\n", row.Operation)
	}
}

func (s *MaterializedStatementResults) processRowWithoutUpsertColumns(row StatementResultRow, rowKey string) {
	// insert row at the end
	if row.Operation.IsInsertOperation() {
		listPtr := s.table.PushBack(row)

		// We need to add the list pointer to the cache for this row key
		s.cache[rowKey] = append(s.cache[rowKey], listPtr)
		return
	}

	// delete row
	listPtrs, ok := s.cache[rowKey]
	if ok && len(listPtrs) > 0 {
		lastPtr := listPtrs[len(listPtrs)-1]
		s.table.Remove(lastPtr)
		s.cache[rowKey] = listPtrs[:len(listPtrs)-1]
		if len(s.cache[rowKey]) == 0 {
			delete(s.cache, rowKey)
		}
	}
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

func (s *MaterializedStatementResults) ForEach(f func(rowIdx int, row *StatementResultRow)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	iterator := s.Iterator(false)
	rowIdx := 0
	for !iterator.HasReachedEnd() {
		f(rowIdx, iterator.GetNext())
		rowIdx++
	}
}

func (s *MaterializedStatementResults) GetMaxWidthPerColumn() []int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	columnWidths := make([]int, len(s.GetHeaders()))
	for colIdx, column := range s.GetHeaders() {
		columnWidths[colIdx] = max(utils.GetMaxStrWidth(column), columnWidths[colIdx])
	}

	s.ForEach(func(rowIdx int, row *StatementResultRow) {
		for colIdx, field := range row.Fields {
			columnWidths[colIdx] = max(utils.GetMaxStrWidth(field.ToString()), columnWidths[colIdx])
		}
	})
	return columnWidths
}

func (s *MaterializedStatementResults) GetMaxResults() int {
	return s.maxCapacity
}

func (s *MaterializedStatementResults) GetTableSize() int {
	return s.table.Len()
}

func (s *MaterializedStatementResults) GetChangelogSize() int {
	return s.changelog.Len()
}
