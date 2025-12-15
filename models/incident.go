package models

import (
	"time"
)

type Incident struct {
	GUID           string         `json:"guid" gorm:"primary_key"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	State          IncidentState  `json:"state"`
	ComponentState ComponentState `json:"component_state"`
	Components     *Components    `json:"components" gorm:"type:varchar(300)"`
	Messages       []Message      `json:"messages" gorm:"ForeignKey:IncidentGUID;"`
	Metadata       []Metadata     `json:"metadata" gorm:"ForeignKey:IncidentGUID;"`
	IsScheduled    bool           `json:"is_scheduled"`
	ScheduledEnd   time.Time      `json:"scheduled_end"`
	Origin         string         `json:"origin"`
	Persistent     bool           `json:"persistent"`
}

type IncidentUpdateRequest struct {
	GUID           *string         `json:"guid"`
	CreatedAt      time.Time       `json:"created_at"`
	State          *IncidentState  `json:"state"`
	ComponentState *ComponentState `json:"component_state"`
	Components     *Components     `json:"components"`
	Messages       *[]Message      `json:"messages"`
	Metadata       *[]Metadata     `json:"metadata"`
	IsScheduled    *bool           `json:"is_scheduled"`
	ScheduledEnd   time.Time       `json:"scheduled_end"`
	Origin         *string         `json:"origin"`
	NoNotify       bool            `json:"no_notify"`
	Persistent     *bool           `json:"persistent"`
}

func (i Incident) MainMessage() Message {
	if len(i.Messages) == 0 {
		return Message{}
	}
	return i.Messages[len(i.Messages)-1]
}

func (i Incident) UpdateMessages() []Message {
	return i.Messages[:len(i.Messages)-1]
}

func (i Incident) UpdateMessagesReverse() []Message {
	messages := i.UpdateMessages()
	newOrder := make([]Message, len(messages))
	p := len(messages) - 1
	for _, message := range messages {
		newOrder[p] = message
		p--
	}
	return newOrder
}

func (i Incident) LastMessage() Message {
	if len(i.Messages) == 0 {
		return Message{}
	}
	return i.Messages[0]
}

func (i Incident) IsNew() bool {
	return i.CreatedAt.Equal(i.UpdatedAt)
}

type Incidents []Incident

func (p Incidents) Len() int {
	return len(p)
}

func (p Incidents) Less(i, j int) bool {
	return p[i].CreatedAt.Before(p[j].CreatedAt)
}

func (p Incidents) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Incidents) Find(guid string) Incident {
	for _, incident := range p {
		if incident.GUID == guid {
			return incident
		}
	}
	return Incident{}
}

func (p Incidents) Filter(guid string) Incidents {
	incidents := make(Incidents, 0)
	for _, incident := range p {
		if incident.GUID == guid {
			continue
		}
		incidents = append(incidents, incident)
	}
	return incidents
}

func TextState(state ComponentState) string {
	switch state {
	case DegradedPerformance:
		return "Degraded Performance"
	case PartialOutage:
		return "Partial Outage"
	case UnderMaintenance:
		return "Under Maintenance"
	case MajorOutage:
		return "Major Outage"
	}
	return "Operational"
}

func TextIncidentState(state IncidentState) string {
	switch state {
	case Resolved:
		return "resolved"
	case Monitoring:
		return "monitoring"
	case Idle:
		return "idle"
	}
	return "unresolved"
}

func TextScheduledState(state IncidentState) string {
	switch state {
	case Resolved:
		return "finished"
	case Monitoring:
		return "monitoring"
	case Idle:
		return "idle"
	}
	return "started"
}

func (i *Incident) HasRealScheduledEndDate() bool {
	return !i.ScheduledEnd.IsZero() && i.ScheduledEnd.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
}

func (i *Incident) ShouldBeIncident(disableMaintenanceToIncident bool) bool {
	if !disableMaintenanceToIncident {
		if (i.IsScheduled && i.ScheduledEnd.After(time.Now())) ||
			(i.IsScheduled && (i.State == Resolved || i.State == Idle)) {
			return false
		}
		return true
	}
	return !i.HasRealScheduledEndDate()
}
