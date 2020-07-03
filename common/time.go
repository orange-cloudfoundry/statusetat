package common

import (
	"time"

	"github.com/hako/durafmt"
)

func HumanDuration(from, to time.Time) string {
	diff := to.Sub(from)
	if diff.Seconds() < 60 {
		return "few seconds"
	}
	dur := durafmt.Parse(diff)
	return dur.LimitFirstN(1).String()
}
