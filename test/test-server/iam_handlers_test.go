package testserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
)

func TestIamApiKeysCreateIssuesDistinctKeysConcurrently(t *testing.T) {
	const creates = 20
	handler := handleIamApiKeysCreate(t)
	ids := make([]string, creates)

	var wg sync.WaitGroup
	for i := range creates {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := `{"spec":{"owner":{"id":"u-123"},"resource":{"id":"lkc-123"}}}`
			rec := httptest.NewRecorder()
			handler(rec, httptest.NewRequest(http.MethodPost, "/iam/v2/api-keys", strings.NewReader(body)))
			key := &apikeysv2.IamV2ApiKey{}
			if err := json.NewDecoder(rec.Body).Decode(key); err != nil {
				t.Errorf("decode create response: %v", err)
				return
			}
			ids[i] = key.GetId()
		}(i)
	}
	wg.Wait()

	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("key id %q issued twice across concurrent creates", id)
		}
		seen[id] = true
	}
}

func TestIamApiKeyGetAndUpdateAreSafeConcurrently(t *testing.T) {
	get := handleIamApiKeyGet(t, "MYKEY1")
	update := handleIamApiKeyUpdate(t, "MYKEY1")

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			get(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/iam/v2/api-keys/MYKEY1", nil))
		}()
		go func() {
			defer wg.Done()
			body := strings.NewReader(`{"spec":{"description":"updated"}}`)
			update(httptest.NewRecorder(), httptest.NewRequest(http.MethodPatch, "/iam/v2/api-keys/MYKEY1", body))
		}()
	}
	wg.Wait()
}
