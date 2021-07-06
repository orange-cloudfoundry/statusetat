package models

import (
	"fmt"
	"strings"
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
	Select
	Checkbox
	Radio
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

func (m MetadataField) ToHTML() string {

	if m.InputType == Text {
		return fmt.Sprintf(`<input id="%s" type="text" name="%s" class="metadata-field">
<label for="%s" class="tooltip" data-position="bottom" data-tooltip="%s">%s</label>`,
			m.Id, m.Id,
			m.Id, m.Info, strings.Title(m.Name))
	}
	if m.InputType == Password {
		return fmt.Sprintf(`<input id="%s" type="password" name="%s" class="metadata-field">
<label for="%s" class="tooltip" data-position="bottom" data-tooltip="%s">%s</label>`,
			m.Id, m.Id,
			m.Id, m.Info, strings.Title(m.Name))
	}

	if m.InputType == Checkbox {
		return fmt.Sprintf(`<label for="%s" title="%s">
        <input type="checkbox" id="%s" name="%s" class="metadata-field"/>
        <span>%s</span>
      </label>`,
			m.Id, m.Info, m.Id, m.Id, m.Name,
		)
	}

	if m.InputType == Radio {
		text := ""
		for i, v := range m.Opts {
			checked := ""
			if i == 0 {
				checked = "checked"
			}
			text += fmt.Sprintf(`<p>
      <label for="%s" class="tooltip" data-position="bottom" data-tooltip="%s">
        <input name="%s" value="%s" class="metadata-field" type="radio" %s />
        <span>%s</span>
      </label>
    </p>`,
				m.Id, m.Info, m.Id, v, checked, strings.Title(v))
		}
		return text
	}
	selectOpts := ""
	for _, v := range m.Opts {
		selectOpts += fmt.Sprintf(`<option value="%s">%s</option>`,
			v, strings.Title(v))
	}
	text := fmt.Sprintf(`<select id="%s" name="%s" class="metadata-field">
      <option value="" disabled selected>Choose your option</option>
      %s
    </select>
    <label for="%s" class="tooltip" data-position="bottom" data-tooltip="%s">%s</label>`, m.Id, m.Id, selectOpts, m.Id, m.Info, strings.Title(m.Name))

	return text

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
