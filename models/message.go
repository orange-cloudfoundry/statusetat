package models

import (
	"time"
)

type Message struct {
	GUID         string    `json:"guid" gorm:"primary_key"`
	IncidentGUID string    `json:"incident_guid"`
	CreatedAt    time.Time `json:"created_at"`
	Title        string    `json:"title"`
	Content      string    `json:"content" gorm:"type:text"`
}

type Messages []Message

func (p Messages) Len() int {
	return len(p)
}

func (p Messages) Less(i, j int) bool {
	return p[i].CreatedAt.After(p[j].CreatedAt)
}

func (p Messages) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
