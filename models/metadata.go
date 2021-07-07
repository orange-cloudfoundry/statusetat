package models

import (
	"fmt"
)

type Metadata struct {
	IncidentGUID string `json:"incident_guid"`
	Key          string `json:"key"`
	Value        string `json:"value"`
}

type InputTypeMedata uint

const (
	Text InputTypeMedata = iota
	Password
	Checkbox
	Radio
	Select
)

type MetadataField struct {
	Name         string
	Id           string
	Info         string
	InputType    InputTypeMedata
	ForScheduled bool
	Opts         []string
}

func (m MetadataField) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("Name must be defined")
	}
	if m.Id == "" {
		return fmt.Errorf("Id must be defined")
	}
	switch m.InputType {
	case Radio, Select:
		if len(m.Opts) == 0 {
			return fmt.Errorf("Opts must be set for a radio or select metadata field")
		}
	}
	return nil
}

type MetadataFields []MetadataField

func (mf MetadataFields) LenIncident() int {
	i := 0
	for _, field := range mf {
		if !field.ForScheduled {
			i++
		}
	}
	return i
}

func (mf MetadataFields) LenScheduled() int {
	i := 0
	for _, field := range mf {
		if field.ForScheduled {
			i++
		}
	}
	return i
}
