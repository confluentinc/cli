package config

type ConnectLogsQueryState struct {
	PageToken   string `json:"page_token,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Level       string `json:"level,omitempty"`
	SearchText  string `json:"search_text,omitempty"`
	ConnectorId string `json:"connector_id,omitempty"`
}

func (c *ConnectLogsQueryState) SetPageToken(pageToken string) {
	c.PageToken = pageToken
}
