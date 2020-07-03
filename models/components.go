package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type Component struct {
	Name  string
	Group string
}

func (c *Component) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Component) UnmarshalJSON(bytes []byte) error {
	var s string
	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}
	split := strings.SplitN(s, " - ", 2)
	if len(split) == 1 {
		c.Name = split[0]
		return nil
	}
	c.Group = split[0]
	c.Name = split[1]
	return nil
}

func (c Component) String() string {
	if c.Group == "" {
		return c.Name
	}
	return c.Group + " - " + c.Name
}

type Components []Component

func (c Components) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Components) Scan(src interface{}) error {
	if err := json.Unmarshal([]byte(src.(string)), c); err != nil {
		return err
	}
	return nil
}

func (c Components) String() string {
	components := make([]string, len(c))
	for i, co := range c {
		components[i] = co.String()
	}
	return strings.Join(components, ", ")
}

func (c *Components) Inline() []string {
	if c == nil {
		return []string{}
	}
	components := make([]string, len(*c))
	for i, co := range *c {
		components[i] = co.String()
	}
	return components
}
