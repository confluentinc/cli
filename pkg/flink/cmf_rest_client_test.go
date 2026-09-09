package flink

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func TestGetApplicationName(t *testing.T) {
	tests := []struct {
		name        string
		application cmfsdk.FlinkApplication
		wantName    string
		wantErr     string
	}{
		{
			name:        "valid name",
			application: cmfsdk.FlinkApplication{Metadata: map[string]interface{}{"name": "my-application"}},
			wantName:    "my-application",
		},
		{
			name:        "missing name",
			application: cmfsdk.FlinkApplication{Metadata: map[string]interface{}{}},
			wantErr:     `application name is required: ensure the resource file contains a non-empty "metadata.name" field`,
		},
		{
			name:        "empty name",
			application: cmfsdk.FlinkApplication{Metadata: map[string]interface{}{"name": ""}},
			wantErr:     `application name is required: ensure the resource file contains a non-empty "metadata.name" field`,
		},
		{
			name:        "non-string name",
			application: cmfsdk.FlinkApplication{Metadata: map[string]interface{}{"name": 123}},
			wantErr:     `application name is required: ensure the resource file contains a non-empty "metadata.name" field`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := getApplicationName(tt.application)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Empty(t, gotName)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantName, gotName)
		})
	}
}

func TestApplicationMethodsRejectInvalidNames(t *testing.T) {
	tests := []struct {
		name   string
		invoke func(*CmfRestClient) error
	}{
		{
			name: "create",
			invoke: func(client *CmfRestClient) error {
				_, err := client.CreateApplication(context.Background(), "environment", cmfsdk.FlinkApplication{})
				return err
			},
		},
		{
			name: "update",
			invoke: func(client *CmfRestClient) error {
				_, err := client.UpdateApplication(context.Background(), "environment", cmfsdk.FlinkApplication{})
				return err
			},
		},
	}

	wantErr := `application name is required: ensure the resource file contains a non-empty "metadata.name" field`
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invoke(&CmfRestClient{})
			require.EqualError(t, err, wantErr)
		})
	}
}

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

	t.Run("page size 0 defaults to 100 and fetches all pages", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(0, pagedFetcher(250, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 250)
		// Requests all use the default size of 100: three carry items (100 + 100 + 50) and a
		// final empty page terminates the loop.
		require.Equal(t, []int32{100, 100, 100, 100}, sizes)
	})

	t.Run("custom page size controls request size and round-trip count", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(50, pagedFetcher(100, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 100)
		// size=50 over 100 items → two full pages then a terminating empty page.
		require.Equal(t, []int32{50, 50, 50}, sizes)
	})

	t.Run("larger page size means fewer round trips for the same data", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(1000, pagedFetcher(250, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 250)
		// One data page of up to 1000 covers all 250, then a terminating empty page.
		require.Equal(t, []int32{1000, 1000}, sizes)
	})

	t.Run("page size larger than total returns all items in one data page", func(t *testing.T) {
		var sizes []int32
		items, err := listAllPages(100, pagedFetcher(30, &sizes))
		require.NoError(t, err)
		require.Len(t, items, 30)
		require.Equal(t, []int32{100, 100}, sizes)
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
