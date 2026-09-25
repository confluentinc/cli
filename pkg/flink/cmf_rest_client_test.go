package flink

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
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

func TestDoCmfRequest(t *testing.T) {
	type capturedRequest struct {
		header http.Header
		body   string
	}
	captured := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured <- capturedRequest{header: r.Header.Clone(), body: string(body)}
		_, _ = w.Write([]byte(`{"name":"my-artifact"}`))
	}))
	defer server.Close()

	// send dispatches a POST carrying an upload body through doCmfRequest and returns the response body.
	send := func(t *testing.T, debug bool) string {
		cfg := cmfsdk.NewConfiguration()
		cfg.HTTPClient = server.Client()
		cfg.UserAgent = "Confluent-CLI/test"
		cfg.DefaultHeader = map[string]string{"X-Default-Header": "applied"}
		cfg.Debug = debug
		client := &CmfRestClient{APIClient: cmfsdk.NewAPIClient(cfg)}

		ctx := context.WithValue(context.Background(), cmfsdk.ContextAccessToken, "my-token")
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, strings.NewReader("artifact bytes"))
		require.NoError(t, err)

		response, err := client.doCmfRequest(ctx, request)
		require.NoError(t, err)
		defer response.Body.Close()
		responseBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		return string(responseBody)
	}

	t.Run("applies the User-Agent, default headers, and bearer token", func(t *testing.T) {
		require.Equal(t, `{"name":"my-artifact"}`, send(t, false))
		got := <-captured
		require.Equal(t, "Confluent-CLI/test", got.header.Get("User-Agent"))
		require.Equal(t, "applied", got.header.Get("X-Default-Header"))
		require.Equal(t, "Bearer my-token", got.header.Get("Authorization"))
		require.Equal(t, "artifact bytes", got.body)
	})

	t.Run("tracing logs headers without the upload body and leaves both bodies intact", func(t *testing.T) {
		var logs bytes.Buffer
		defer log.SetOutput(log.Writer())
		log.SetOutput(&logs)

		require.Equal(t, `{"name":"my-artifact"}`, send(t, true))
		require.Equal(t, "artifact bytes", (<-captured).body)
		require.Contains(t, logs.String(), "User-Agent: Confluent-CLI/test")
		require.NotContains(t, logs.String(), "artifact bytes")
		require.Contains(t, logs.String(), `{"name":"my-artifact"}`)
	})
}
