package errors

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

var (
	paymentRequiredErrorMsg  = "402 Payment Required"
	clusterExceedMsg         = "Your environment is currently limited to 50 kafka clusters"
	serviceAccountsExceedMsg = "Your environment is currently limited to 1000 service accounts"
)

type test struct {
	name     string
	err      error
	response *http.Response
	want     string
	wantErr  bool
}

func TestCatchClustersExceedError(t *testing.T) {
	t.Parallel()

	tt := test{
		name:     "cluster exceed",
		err:      New(paymentRequiredErrorMsg),
		response: &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 50 kafka clusters"}]}`))},
		want:     clusterExceedMsg + ": " + paymentRequiredErrorMsg,
		wantErr:  true,
	}
	var err error
	if err = CatchClusterConfigurationNotValidError(tt.err, tt.response); (err != nil) != tt.wantErr {
		t.Errorf("CatchClusterConfigurationNotValidError()\nerror: %v\nwantErr: %v", err, tt.wantErr)
	}
	if err.Error() != tt.want {
		t.Errorf("CatchClusterConfigurationNotValidError()\ngot: %s\nwant: %s", err, tt.want)
	}
}

func TestCatchServiceAccountExceedError(t *testing.T) {
	t.Parallel()

	tt := test{
		name:     "service accounts exceed",
		err:      New(paymentRequiredErrorMsg),
		response: &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 1000 service accounts"}]}`))},
		want:     serviceAccountsExceedMsg + ": " + paymentRequiredErrorMsg,
		wantErr:  true,
	}
	var err error
	if err = CatchServiceNameInUseError(tt.err, tt.response, ""); (err != nil) != tt.wantErr {
		t.Errorf("CatchServiceNameInUseError()\nerror: %v\nwantErr: %v", err, tt.wantErr)
	}
	if err.Error() != tt.want {
		t.Errorf("CatchServiceNameInUseError()\ngot: %s\nwant: %s", err, tt.want)
	}
}
