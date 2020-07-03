package models

type Metadata struct {
	IncidentGUID string `json:"incident_guid"`
	Key          string `json:"key"`
	Value        string `json:"value"`
}
