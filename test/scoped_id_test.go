package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd/cluster"
)

func (s *CLITestSuite) TestClusterScopedId() {
	// everything
	cpIdURL1 := serveClusterScopedId(&cluster.ScopedId{
		ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
		Scope: &cluster.Scope{
			Path:     []string{"This", "Is", "Ignored"},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1", "connect-cluster": "connectClusterA"},
		},
	}, s.T()).URL

	// no id
	cpIdURL2 := serveClusterScopedId(&cluster.ScopedId{
		ID: "",
		Scope: &cluster.Scope{
			Path:     []string{},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1", "connect-cluster": "connectClusterA"},
		},
	}, s.T()).URL

	// just kafka
	cpIdURL3 := serveClusterScopedId(&cluster.ScopedId{
		ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
		Scope: &cluster.Scope{
			Path:     []string{},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1"},
		},
	}, s.T()).URL

	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	caCertPath := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "localhost.pem")

	cpIdURL3TLS := serveTLSClusterScopedId(&cluster.ScopedId{
		ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
		Scope: &cluster.Scope{
			Path:     []string{},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1"},
		},
	}, s.T()).URL

	// old versions of CP without the cluster metadata id endpoint respond with 401
	cpIdURL4 := serveClusterScopedIdError().URL

	tests := []CLITest{
		{args: fmt.Sprintf("cluster describe --url %s", cpIdURL1), fixture: "scoped_id1.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", cpIdURL2), fixture: "scoped_id2.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", cpIdURL3), fixture: "scoped_id3.golden"},
		{args: fmt.Sprintf("cluster describe --url %s --ca-cert-path %s", cpIdURL3TLS, caCertPath), fixture: "scoped_id3.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", cpIdURL4), fixture: "scoped_id4.golden", wantErrCode: 1},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, "")
	}
}

func serveClusterScopedId(meta *cluster.ScopedId, t *testing.T) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/v1/metadata/id", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(meta)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	return httptest.NewServer(router)
}

func serveTLSClusterScopedId(meta *cluster.ScopedId, t *testing.T) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/v1/metadata/id", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(meta)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	})
	server := httptest.NewUnstartedServer(router)
	server.StartTLS()
	return server
}

func serveClusterScopedIdError() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["WWW-Authenticate"] = []string{"Bearer realm=\"\""}
		w.Header()["Cache-Control"] = []string{"must-revalidate,no-cache,no-store"}
		w.WriteHeader(http.StatusUnauthorized)
	})
	return httptest.NewServer(router)
}
