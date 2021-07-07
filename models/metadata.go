package models

import (
	"fmt"
	"reflect"
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
	Opts         interface{}
	DefaultOpt   interface{}
}

func (m MetadataField) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("Name must be defined")
	}
	if m.Id == "" {
		return fmt.Errorf("Id must be defined")
	}
	switch m.InputType {
	case Radio:
		valOf := reflect.ValueOf(m.Opts)
		if valOf.IsNil() {
			return fmt.Errorf("Opts must be set for a radio as a slice")
		}
		if valOf.Kind() != reflect.Slice && valOf.Kind() != reflect.Array {
			return fmt.Errorf("Opts must be set for a radio as a slice, type %s given", valOf.Type())
		}
		if valOf.Len() == 0 {
			return fmt.Errorf("Opts must be set for a radio as a slice, empty one given")
		}

	case Select:
		valOf := reflect.ValueOf(m.Opts)
		if valOf.IsNil() {
			return fmt.Errorf("Opts must be set for a select as a map")
		}
		if valOf.Kind() != reflect.Map {
			return fmt.Errorf("Opts must be set for a select as a map, type %s given", valOf.Type())
		}
		if valOf.Len() == 0 {
			return fmt.Errorf("Opts must be be set for a select as a map, empty one given")
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
