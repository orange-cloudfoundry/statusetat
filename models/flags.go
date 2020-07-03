package models

const (
	Operational ComponentState = iota
	UnderMaintenance
	DegradedPerformance
	PartialOutage
	MajorOutage
)

const (
	Unresolved IncidentState = iota
	Resolved
	Monitoring
)

var AllIncidentState = []IncidentState{Unresolved, Monitoring, Resolved}
var AllComponentState = []ComponentState{MajorOutage, PartialOutage, DegradedPerformance, UnderMaintenance, Operational}

type ComponentState int
type IncidentState int
