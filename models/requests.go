package models

type NotifyRequest struct {
	Incident      Incident
	TriggerByUser bool
	Subscribers   []string
}

func NewNotifyRequest(incident Incident, triggerByUser bool) *NotifyRequest {
	return &NotifyRequest{Incident: incident, TriggerByUser: triggerByUser}
}
