package local

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListPlugins(t *testing.T) {
	req := require.New(t)

	server, err := mockConnectService()
	req.NoError(err)
	defer server.Close()

	out, err := mockLocalCommand("plugins", "list")
	req.NoError(err)
	req.Contains(out, "{\n  \"key\": \"val\"\n}\n")
}

func mockConnectService() (*httptest.Server, error) {
	router := http.NewServeMux()
	router.HandleFunc("/connector-plugins", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"key\":\"val\"}")
	})
	server := httptest.NewUnstartedServer(router)

	lis, err := net.Listen("tcp", "localhost:8083")
	if err != nil {
		return nil, err
	}
	server.Listener = lis
	server.Start()

	return server, nil
}
