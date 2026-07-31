package flink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListAllPages(t *testing.T) {
	// pagedFetcher emulates a CMF endpoint that pages by zero-based index (offset = page * size).
	// It records the sizes it was asked for so tests can assert on the requested page size.
	pagedFetcher := func(total int, requestedSizes *[]int32) func(page, size int32) ([]int, error) {
		return func(page, size int32) ([]int, error) {
			*requestedSizes = append(*requestedSizes, size)
			start := int(page * size)
			if start >= total {
				return []int{}, nil
			}
			end := start + int(size)
			if end > total {
				end = total
			}
			items := make([]int, 0, end-start)
			for i := start; i < end; i++ {
				items = append(items, i)
			}
			return items, nil
		}
	}

	t.Run("limit 0 returns all items across pages", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(0, pagedFetcher(250, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 250)
		// limit == 0 keeps the requested page size constant at 100. Four pages are
		// requested: three carrying items (100 + 100 + 50) and a final empty page
		// that terminates the loop.
		require.Equal(t, []int32{100, 100, 100, 100}, sizes)
	})

	t.Run("limit smaller than page size fetches a single small page", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(3, pagedFetcher(250, &sizes))
		require.NoError(t, err)
		require.Equal(t, []int{0, 1, 2}, items)
		// Only one page of exactly the limit size is requested.
		require.Equal(t, []int32{3}, sizes)
	})

	t.Run("limit spanning multiple pages truncates to exactly limit", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(150, pagedFetcher(250, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 150)
		require.Equal(t, 149, items[149])
		// Page size stays constant at 100 across pages.
		require.Equal(t, []int32{100, 100}, sizes)
	})

	t.Run("limit larger than total returns all items", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(1000, pagedFetcher(30, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 30)
	})

	t.Run("empty result set", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(0, pagedFetcher(0, &sizes))
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("propagates fetch error", func(t *testing.T) {
		wantErr := fmt.Errorf("boom")
		_, err := listAllPages(0, func(page, size int32) ([]int, error) {
			return nil, wantErr
		})
		require.ErrorIs(t, err, wantErr)
	})
}
