package types

import (
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
	"testing"
)

func TestDequeModel(t *testing.T) {

	rapid.Check(t, func(t *rapid.T) {
		entries := rapid.SliceOfN(rapid.StringN(5, 10, 20), 10, 100).Draw(t, "entries")

		list := NewLinkedList[string]()
		for _, entry := range entries {
			list.PushBack(entry)
		}

		list.RemoveFront()

		require.Equal(t, entries[1:], list.ToSlice())
		list.RemoveBack()
		require.Equal(t, entries[1:len(entries)-1], list.ToSlice())

		list.PushBack(entries[len(entries)-1])
		require.Equal(t, entries[1:], list.ToSlice())

		list.PushFront(entries[0])
		require.Equal(t, entries, list.ToSlice())

		for idx, val := range entries {
			require.Equal(t, val, *list.ElementAtIndex(idx).Value())
		}

		list.Remove(list.ElementAtIndex(5))
		require.Equal(t, append(entries[:5], entries[6:]...), list.ToSlice())
	})

}
