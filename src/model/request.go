package model

import "encoding/json"

type TransformRequest struct {
	LabelsActions []LabelsAction `json:"labelsActions"`

	Source Cluster `json:"source"`
	Target Cluster `json:"target"`
}

func (tr *TransformRequest) String() string {
	if dataInBytes, err := json.Marshal(tr); err == nil {
		return string(dataInBytes)
	}

	return ""
}
