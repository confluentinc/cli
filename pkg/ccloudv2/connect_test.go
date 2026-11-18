package ccloudv2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
)

const (
	testConnectEnvironmentId = "env-123456"
	testConnectClusterId     = "lkc-123456"
	// maxConnectListRequests stops a client that never ends the listing, e.g. one that drops the page
	// token and keeps asking for the first page, so the test fails instead of hanging.
	maxConnectListRequests = 20
)

// connectListPage is one page served by connectListServer: the connectors it returns and the
// X-Next-Page-Token it sends, if any.
type connectListPage struct {
	connectors    []string
	nextPageToken string
	status        int
}

// connectListServer serves the connector list API from fixed pages keyed by page_token, where the
// first page has the empty token, and records the query of every request.
type connectListServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []url.Values
}

func newConnectListServer(t *testing.T, pages map[string]connectListPage) *connectListServer {
	t.Helper()
	server := &connectListServer{}
	path := fmt.Sprintf("/connect/v1/environments/%s/clusters/%s/connectors", testConnectEnvironmentId, testConnectClusterId)
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.mu.Lock()
		server.requests = append(server.requests, r.URL.Query())
		count := len(server.requests)
		server.mu.Unlock()

		if count > maxConnectListRequests {
			t.Errorf("client sent more than %d list requests; the paging loop did not end", maxConnectListRequests)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page, ok := pages[r.URL.Query().Get("page_token")]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error_code": 400, "message": "Invalid page_token"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if page.nextPageToken != "" {
			w.Header().Set(connectNextPageTokenHeader, page.nextPageToken)
		}
		if page.status != 0 && page.status != http.StatusOK {
			w.WriteHeader(page.status)
			_, _ = w.Write([]byte(`{"error_code": 500, "message": "internal error"}`))
			return
		}
		body := make(map[string]any, len(page.connectors))
		for _, name := range page.connectors {
			body[name] = map[string]any{"id": map[string]string{"id": "lcc-" + name, "id_type": "ID"}}
		}
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func (s *connectListServer) recordedRequests() []url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]url.Values(nil), s.requests...)
}

func newTestConnectClient(server *connectListServer) *Client {
	return &Client{
		cfg:           config.AuthenticatedCloudConfigMock(),
		ConnectClient: newConnectClient(server.Client(), server.URL, "test-agent", false),
	}
}

func connectorNames(connectors map[string]connectv1.ConnectV1ConnectorExpansion) []string {
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	return names
}

func TestListConnectorsWithExpansions_MultiplePages(t *testing.T) {
	server := newConnectListServer(t, map[string]connectListPage{
		"":   {connectors: []string{"a", "b"}, nextPageToken: "t1"},
		"t1": {connectors: []string{"c", "d"}, nextPageToken: "t2"},
		"t2": {connectors: []string{"e"}},
	})

	connectors, err := newTestConnectClient(server).ListConnectorsWithExpansions(testConnectEnvironmentId, testConnectClusterId, "id,status")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a", "b", "c", "d", "e"}, connectorNames(connectors))
	require.Equal(t, "lcc-c", connectors["c"].Id.GetId())

	requests := server.recordedRequests()
	require.Len(t, requests, 3)
	for i, query := range requests {
		require.Equal(t, strconv.Itoa(int(ccloudV2ListPageSize)), query.Get("page_size"), "request %d", i)
		// The SDK always adds its own default expand=info,status,id, so look for the exact value the
		// caller asked for among the expand params of every page.
		require.Contains(t, query["expand"], "id,status", "request %d", i)
	}
	require.False(t, requests[0].Has("page_token"))
	require.Equal(t, "t1", requests[1].Get("page_token"))
	require.Equal(t, "t2", requests[2].Get("page_token"))
}

func TestListConnectorsWithExpansions_EmptyPageWithToken(t *testing.T) {
	// The server may return a short or empty page while more remain, e.g. between managed and custom connectors.
	server := newConnectListServer(t, map[string]connectListPage{
		"":   {connectors: []string{"a", "b"}, nextPageToken: "t1"},
		"t1": {nextPageToken: "t2"},
		"t2": {connectors: []string{"c"}},
	})

	connectors, err := newTestConnectClient(server).ListConnectorsWithExpansions(testConnectEnvironmentId, testConnectClusterId, "id")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a", "b", "c"}, connectorNames(connectors))
	require.Len(t, server.recordedRequests(), 3)
}

func TestListConnectorsWithExpansions_PaginationDisabled(t *testing.T) {
	// With pagination off for the organization, the server ignores page_size and returns every
	// connector with no X-Next-Page-Token.
	server := newConnectListServer(t, map[string]connectListPage{
		"": {connectors: []string{"a", "b", "c"}},
	})

	connectors, err := newTestConnectClient(server).ListConnectorsWithExpansions(testConnectEnvironmentId, testConnectClusterId, "id")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a", "b", "c"}, connectorNames(connectors))
	require.Len(t, server.recordedRequests(), 1)
}

func TestListConnectorsWithExpansions_EmptyCluster(t *testing.T) {
	server := newConnectListServer(t, map[string]connectListPage{
		"": {},
	})

	connectors, err := newTestConnectClient(server).ListConnectorsWithExpansions(testConnectEnvironmentId, testConnectClusterId, "id")
	require.NoError(t, err)
	require.Empty(t, connectors)
	require.Len(t, server.recordedRequests(), 1)
}

func TestListConnectorsWithExpansions_ErrorOnLaterPage(t *testing.T) {
	server := newConnectListServer(t, map[string]connectListPage{
		"":   {connectors: []string{"a", "b"}, nextPageToken: "t1"},
		"t1": {status: http.StatusInternalServerError},
	})

	connectors, err := newTestConnectClient(server).ListConnectorsWithExpansions(testConnectEnvironmentId, testConnectClusterId, "id")
	require.Error(t, err)
	require.Nil(t, connectors, "a failed page must not return the connectors from earlier pages")
	require.Len(t, server.recordedRequests(), 2)
}
