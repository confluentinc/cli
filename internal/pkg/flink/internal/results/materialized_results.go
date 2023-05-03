package results

import (
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"sync"
)

type materializedStatementResultsIterator struct {
	isTableMode bool
	iterator    *types.ListElement[types.StatementResultRow]
	lock        sync.RWMutex
}

func (i *materializedStatementResultsIterator) HasNext() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()

	return i.iterator != nil
}

func (i *materializedStatementResultsIterator) GetNext() *types.StatementResultRow {
	i.lock.Lock()
	defer i.lock.Unlock()

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

func (s *MaterializedStatementResults) Iterator() materializedStatementResultsIterator {
	s.lock.RLock()
	defer s.lock.RUnlock()

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
	s.lock.Lock()
	defer s.lock.Unlock()

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
