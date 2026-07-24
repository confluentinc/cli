package flink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

// newTestCmfClient builds a CmfRestClient that talks to srv.
func newTestCmfClient(srv *httptest.Server) *CmfRestClient {
	cfg := cmfsdk.NewConfiguration()
	cfg.Servers = cmfsdk.ServerConfigurations{{URL: srv.URL}}
	cfg.HTTPClient = srv.Client()
	return &CmfRestClient{APIClient: cmfsdk.NewAPIClient(cfg)}
}

// sortRecorder captures the sort query params of the most recent request.
type sortRecorder struct {
	mu   sync.Mutex
	last []string
}

func (s *sortRecorder) set(v []string) {
	s.mu.Lock()
	s.last = v
	s.mu.Unlock()
}

func (s *sortRecorder) get() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

// TestListStatements_ReturnsAllItemsAcrossPages is the CF-4076 regression: a
// paginated list of >100 items must return every item exactly once. The server
// pages a stable, name-sorted dataset and rejects any request that omits sort.
func TestListStatements_ReturnsAllItemsAcrossPages(t *testing.T) {
	const total = 250 // spans page boundaries at 100 and 200

	names := make([]string, total)
	for i := range names {
		names[i] = fmt.Sprintf("stmt-%04d", i)
	}

	rec := &sortRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		rec.set(q["sort"])
		if len(q["sort"]) == 0 {
			http.Error(w, "no sort: paging would be non-deterministic", http.StatusInternalServerError)
			return
		}
		page, _ := strconv.Atoi(q.Get("page"))
		size, _ := strconv.Atoi(q.Get("size"))

		var items []cmfsdk.Statement
		if start := page * size; start < total {
			for _, n := range names[start:min(start+size, total)] {
				items = append(items, cmfsdk.Statement{Metadata: cmfsdk.StatementMetadata{Name: n}})
			}
		}
		out := cmfsdk.StatementsPage{}
		out.SetItems(items)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(out))
	}))
	defer srv.Close()

	got, err := newTestCmfClient(srv).ListStatements(context.Background(), "env", "", "")
	require.NoError(t, err)
	require.Equal(t, []string{"name"}, rec.get())

	require.Len(t, got, total)
	seen := make(map[string]int, total)
	for _, s := range got {
		seen[s.Metadata.Name]++
	}
	require.Len(t, seen, total, "every statement returned exactly once")
	for name, count := range seen {
		require.Equal(t, 1, count, "statement %q returned %d times", name, count)
	}
}

// TestListCommands_SendUniqueSortKey locks in the per-resource sort key sent by
// every paginated list call. An empty page ends the loop after one request, so
// the recorder holds exactly that call's sort params.
func TestListCommands_SendUniqueSortKey(t *testing.T) {
	rec := &sortRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.set(r.URL.Query()["sort"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	c := newTestCmfClient(srv)
	ctx := context.Background()

	cases := []struct {
		name   string
		invoke func() error
		want   []string
	}{
		{"statements", func() error { _, e := c.ListStatements(ctx, "env", "", ""); return e }, []string{"name"}},
		{"applications", func() error { _, e := c.ListApplications(ctx, "env"); return e }, []string{"name"}},
		{"application-instances", func() error { _, e := c.ListApplicationInstances(ctx, "env", "app"); return e }, []string{"name"}},
		{"application-events", func() error { _, e := c.ListApplicationEvents(ctx, "env", "app"); return e }, []string{"creationTimestamp,desc", "name"}},
		{"savepoints-statement", func() error { _, e := c.ListSavepoint(ctx, "env", "stmt", "", true); return e }, []string{"name"}},
		{"savepoints-application", func() error { _, e := c.ListSavepoint(ctx, "env", "", "app", false); return e }, []string{"name"}},
		{"detached-savepoints", func() error { _, e := c.ListDetachedSavepoint(ctx, ""); return e }, []string{"name"}},
		{"compute-pools", func() error { _, e := c.ListComputePools(ctx, "env"); return e }, []string{"name"}},
		{"environments", func() error { _, e := c.ListEnvironments(ctx); return e }, []string{"name"}},
		{"catalogs", func() error { _, e := c.ListCatalog(ctx); return e }, []string{"name"}},
		{"databases", func() error { _, e := c.ListDatabases(ctx, "cat"); return e }, []string{"name"}},
		{"secrets", func() error { _, e := c.ListSecrets(ctx); return e }, []string{"name"}},
		{"secret-mappings", func() error { _, e := c.ListSecretMappings(ctx, "env"); return e }, []string{"name", "uid"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.invoke())
			require.Equal(t, tc.want, rec.get())
		})
	}
}
