package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testConfig(r float64) (Config, *[]string) {
	var notices []string
	cfg := Default
	cfg.random = func() float64 { return r }
	cfg.printNotice = func(s string) { notices = append(notices, s) }
	return cfg, &notices
}

func TestBackoff_Computed(t *testing.T) {
	tests := []struct {
		r    float64   // the random param, 0 to 1, with .5 being no jitter
		want []float64 // in ms, each entry is an attempt
	}{
		{0, []float64{800, 1280, 2048, 3276.8}},
		{0.5, []float64{1000, 1600, 2560, 4096}},
		{1, []float64{1200, 1920, 3072, 4915.2}},
	}

	for _, test := range tests {
		cfg, _ := testConfig(test.r)
		for attempt, ms := range test.want {
			got := cfg.backoff(Default.RetryWaitMin, Default.RetryWaitMax, attempt, nil)
			require.InDelta(t, ms*float64(time.Millisecond), got, float64(time.Microsecond), "r=%v attempt=%d", test.r, attempt)
		}
	}
}

func TestBackoff_JitterAtCap(t *testing.T) {
	cfg, _ := testConfig(0)
	require.Equal(t, 24*time.Second, cfg.backoff(time.Second, 30*time.Second, 8, nil))
}

func TestBackoff_RetryAfter(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		r          float64
		want       time.Duration
	}{
		{"no jitter", "10", 0, 10 * time.Second},
		{"at the maximum, full jitter", "30", 1, 36 * time.Second},
		{"zero uses the base wait for jitter", "0", 1, 200 * time.Millisecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, _ := testConfig(test.r)
			got := cfg.backoff(time.Second, 30*time.Second, 3, response(http.StatusTooManyRequests, test.retryAfter))
			require.Equal(t, test.want, got)
		})
	}
}

func TestNotice(t *testing.T) {
	tests := []struct {
		name string
		wait time.Duration
		resp *http.Response
		want []string
	}{
		{"short wait", 2900 * time.Millisecond, nil, nil},
		{"connection failure", 4 * time.Second, nil, []string{"Connection failed; retrying in 4s."}},
		{"status", 3600 * time.Millisecond, response(http.StatusTooManyRequests, ""), []string{"Received 429 Too Many Requests; retrying in 4s."}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, notices := testConfig(0)
			cfg.notice(test.wait, test.resp)
			require.Equal(t, test.want, *notices)
		})
	}
}

func TestSilenceRetryNotices(t *testing.T) {
	t.Cleanup(func() { noticesSilenced.Store(false) })
	cfg, notices := testConfig(0)

	cfg.notice(10*time.Second, nil)
	require.Len(t, *notices, 1)

	SilenceRetryNotices()
	cfg.notice(10*time.Second, nil)
	require.Len(t, *notices, 1) // still only the line from before silencing
}
