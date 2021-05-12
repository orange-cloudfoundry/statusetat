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
	}
	return "unresolved"
}
