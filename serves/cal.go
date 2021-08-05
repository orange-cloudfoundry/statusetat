package serves

import (
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"

	"github.com/orange-cloudfoundry/statusetat/models"
)

func (a Serve) Ical(w http.ResponseWriter, req *http.Request) {
	from, to, err := a.periodFromReq(req, -26, 26)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	scheduled, err := a.scheduled(from, to)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	for _, sch := range scheduled {
		if sch.State == models.Resolved {
			continue
		}
		mainMsg := sch.MainMessage()
		event := cal.AddEvent(sch.GUID)
		event.SetCreatedTime(time.Now().In(a.Location(req)))
		event.SetStartAt(sch.CreatedAt)
		event.SetEndAt(sch.ScheduledEnd)
		event.SetSummary(mainMsg.Title)
		event.SetLocation(sch.Components.String())
		event.SetDescription(mainMsg.Content)
		event.SetURL(a.BaseURL())
	}

	w.Header().Add("Content-Type", "text/calendar")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(cal.Serialize()))
}
