package config

import (
	"github.com/orange-cloudfoundry/statusetat/models"
)

type ForComponent struct {
	RequireAll bool     `yaml:"require_all"`
	GroupMatch []string `yaml:"groups"`
	NameMatch  []string `yaml:"names"`
}

func (lm ForComponent) MatchComponents(components models.Components) bool {
	if len(lm.GroupMatch) == 0 && len(lm.NameMatch) == 0 {
		return true
	}
	match := false
	for _, component := range components {
		match = lm.MatchComponent(component)
		if !match && lm.RequireAll {
			return false
		}
		if match && !lm.RequireAll {
			return true
		}
	}
	return match
}

func (lm ForComponent) MatchComponent(component models.Component) bool {
	if len(lm.GroupMatch) == 0 && len(lm.NameMatch) == 0 {
		return true
	}
	match := false
	for _, groupToMatch := range lm.GroupMatch {
		match = groupToMatch == component.Group
		if !match && lm.RequireAll {
			return false
		}
		if match && !lm.RequireAll {
			return true
		}
	}
	for _, nameToMatch := range lm.NameMatch {
		match = nameToMatch == component.Name
		if !match && lm.RequireAll {
			return false
		}
		if match && !lm.RequireAll {
			return true
		}
	}
	return match
}
