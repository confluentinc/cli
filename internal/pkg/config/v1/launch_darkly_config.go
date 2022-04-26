package v1

type LDConfig struct {
	AnonFlagValues      map[string]interface{} `json:"anon_flag_values" hcl:"anon_flag_values"`
	AnonFlagsUpdateTime int64                  `json:"anon_flags_update_time hcl:anon_flags_update_time"`
	AuthFlagValues      map[string]interface{} `json:"auth_flag_values" hcl:"auth_flag_values"`
	AuthFlagsUpdateTime int64                  `json:"auth_flags_update_time hcl:auth_flags_update_time"`
}
