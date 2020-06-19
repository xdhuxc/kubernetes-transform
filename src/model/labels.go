package model

type LabelsAction struct {
	Action string            `json:"action"`
	Labels map[string]string `json:"labels"`
}
