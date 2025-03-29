package dto

type OrchestrationEntity struct {
	UUID           string `json:"uuid"`
	Status         string `json:"status"`
	ExpirationTime int64  `json:"expirationTime"`
}
