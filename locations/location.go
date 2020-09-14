package locations

import (
	"time"
)

var currentLocation = time.Local

func LoadByTimezone(timeZone string) error {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return err
	}
	currentLocation = loc
	return nil
}

func DefaultLocation() *time.Location {
	return currentLocation
}
