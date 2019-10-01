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
	metadataURL := serveMetadata(s.T()).URL

	tests := []CLITest{
		{args: fmt.Sprintf("cluster describe --url %s", metadataURL), fixture: "metadata1.golden"},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt)
	}
}

func serveMetadata(t *testing.T) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/v1/metadata/id", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&cluster.ClusterMetadata{
			ID: "crn://md01.example.com/kafka=kafkaCluster1/connect=connectClusterA",
			Scope: &cluster.Scope{
				Path: []string{},
				Clusters: map[string]string{"kafka-cluster": "kafkaCluster1", "connect-cluster": "connectClusterA"},
			},
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)

	})
	return httptest.NewServer(router)
}
