package dto

type ServerInfo struct {
	ServiceId string   `json:"service_id"`
	Methods   []string `json:"methods"`
}
