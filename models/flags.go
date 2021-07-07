package models

type ComponentState int

const (
	Operational ComponentState = iota
	UnderMaintenance
	DegradedPerformance
	PartialOutage
	MajorOutage
)

type IncidentState int

const (
	Unresolved IncidentState = iota
	Resolved
	Monitoring
	Idle
)

var AllIncidentState = []IncidentState{Unresolved, Monitoring, Resolved}
var AllComponentState = []ComponentState{MajorOutage, PartialOutage, DegradedPerformance, UnderMaintenance, Operational}
