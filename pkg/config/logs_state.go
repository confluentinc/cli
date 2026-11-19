package config

// LogsQueryState remembers the last `confluent logs fetch` query, so that `--next` can continue it. Key identifies
// the query and its window, and StartTime and EndTime are the absolute bounds it used, so the next page is read from
// the same window even when the query was given as a relative --since.
type LogsQueryState struct {
	Key       string `json:"key,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}
