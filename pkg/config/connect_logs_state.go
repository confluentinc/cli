package config

type ConnectLogsQueryState struct {
	PageToken   string `json:"page_token,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Level       string `json:"level,omitempty"`
	TaskId      string `json:"task_id,omitempty"`
	SearchText  string `json:"search_text,omitempty"`
	ConnectorId string `json:"connector_id,omitempty"`
}
