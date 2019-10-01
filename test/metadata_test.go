package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/cmd/cluster"
)

func (s *CLITestSuite) TestClusterMetadata() {
	// everything
	metadataURL1 := serveMetadata(&cluster.ClusterMetadata{
		ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
		Scope: &cluster.Scope{
			Path: []string{"This", "Is", "Ignored"},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1", "connect-cluster": "connectClusterA"},
		},
	}, s.T()).URL

	// no id
	metadataURL2 := serveMetadata(&cluster.ClusterMetadata{
		ID: "",
		Scope: &cluster.Scope{
			Path: []string{},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1", "connect-cluster": "connectClusterA"},
		},
	}, s.T()).URL

	// just kafka
	metadataURL3 := serveMetadata(&cluster.ClusterMetadata{
		ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
		Scope: &cluster.Scope{
			Path: []string{},
			Clusters: map[string]string{"kafka-cluster": "kafkaCluster1"},
		},
	}, s.T()).URL

	// old versions of CP without the metadata endpoint respond with 401
	metadataURL4 := serveMetadataError().URL

	tests := []CLITest{
		{args: fmt.Sprintf("cluster describe --url %s", metadataURL1), fixture: "metadata1.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", metadataURL2), fixture: "metadata2.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", metadataURL3), fixture: "metadata3.golden"},
		{args: fmt.Sprintf("cluster describe --url %s", metadataURL4), fixture: "metadata4.golden", wantErrCode: 1},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func serveMetadata(meta *cluster.ClusterMetadata, t *testing.T) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/v1/metadata/id", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(meta)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)

	})
	return httptest.NewServer(router)
}

func serveMetadataError() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header()["WWW-Authenticate"] = []string{"Bearer realm=\"\""}
		w.Header()["Cache-Control"] = []string{"must-revalidate,no-cache,no-store"}
		w.WriteHeader(http.StatusUnauthorized)
	})
	return httptest.NewServer(router)
}
