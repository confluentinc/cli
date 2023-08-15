package types

import "container/list"

type ListElement[T any] struct {
	element *list.Element
}

func (e ListElement[T]) Value() *T {
	if e.element == nil {
		return nil
	}
	val := e.element.Value.(T)
	return &val
}

func (e ListElement[T]) Prev() *ListElement[T] {
	if e.element == nil || e.element.Prev() == nil {
		return nil
	}
	return &ListElement[T]{element: e.element.Prev()}
}

func (e ListElement[T]) Next() *ListElement[T] {
	if e.element == nil || e.element.Next() == nil {
		return nil
	}
	return &ListElement[T]{element: e.element.Next()}
}

type LinkedList[T any] struct {
	list *list.List
}

func (l LinkedList[T]) PushBack(row T) *ListElement[T] {
	if l.list == nil {
		return nil
	}
	return &ListElement[T]{element: l.list.PushBack(row)}
}

func (l LinkedList[T]) PushFront(row T) *ListElement[T] {
	if l.list == nil {
		return nil
	}
	return &ListElement[T]{element: l.list.PushFront(row)}
}

func (l LinkedList[T]) Remove(listElement *ListElement[T]) *T {
	if l.list == nil || listElement == nil {
		return nil
	}
	val := l.list.Remove(listElement.element).(T)
	return &val
}

func (l LinkedList[T]) RemoveFront() *T {
	return l.Remove(l.Front())
}

func (l LinkedList[T]) RemoveBack() *T {
	return l.Remove(l.Back())
}

func (l LinkedList[T]) ElementAtIndex(idx int) *ListElement[T] {
	head := l.Front()
	for id := 0; id != idx; id++ {
		head = head.Next()
	}
	return head
}

func (l LinkedList[T]) Front() *ListElement[T] {
	if l.list == nil || l.list.Front() == nil {
		return nil
	}
	return &ListElement[T]{element: l.list.Front()}
}

func (l LinkedList[T]) Back() *ListElement[T] {
	if l.list == nil || l.list.Back() == nil {
		return nil
	}
	return &ListElement[T]{element: l.list.Back()}
}

func (l LinkedList[T]) Len() int {
	if l.list == nil {
		return 0
	}
	return l.list.Len()
}

func (l LinkedList[T]) Clear() {
	if l.list == nil {
		return
	}
	l.list.Init()
}

func (l LinkedList[T]) ToSlice() []T {
	if l.list == nil {
		return nil
	}

	slice := make([]T, 0, l.list.Len())
	for e := l.list.Front(); e != nil; e = e.Next() {
		slice = append(slice, e.Value.(T))
	}
	return slice
}

func NewLinkedList[T any]() LinkedList[T] {
	return LinkedList[T]{list: list.New()}
}
