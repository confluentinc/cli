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

func NewLinkedList[T any]() LinkedList[T] {
	return LinkedList[T]{list: list.New()}
}
