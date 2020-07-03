package serves

import (
	"net/http"
	"time"

	"github.com/ArthurHlt/statusetat/models"
	ics "github.com/arran4/golang-ical"
)

func (a Serve) Ical(w http.ResponseWriter, req *http.Request) {

	scheduled, err := a.scheduled(req)
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
		event.SetURL(a.baseInfo.BaseURL)
	}

	w.Header().Add("Content-Type", "text/calendar")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(cal.Serialize()))
}
