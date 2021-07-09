package serves

import (
	"net/http"
	"sort"
	"time"

	"github.com/orange-cloudfoundry/statusetat/models"
)

func (a Serve) scheduled(from, to time.Time) ([]models.Incident, error) {
	scheduled := make([]models.Incident, 0)

	incidents, err := a.store.ByDate(from, to)
	if err != nil {
		return scheduled, err
	}
	for _, incident := range incidents {
		if (incident.IsScheduled && incident.ScheduledEnd.After(time.Now())) ||
			(incident.IsScheduled && (incident.State == models.Resolved || incident.State == models.Idle)) {
			scheduled = append(scheduled, incident)
		}
	}
	sort.Sort(models.Incidents(scheduled))
	return scheduled, nil
}

func (a Serve) periodFromReq(req *http.Request, nbDaysFrom, nbDaysTo int) (from, to time.Time, err error) {
	defaultNbDaysFrom := time.Duration(nbDaysFrom) * 24 * time.Hour
	defaultNbDaysTo := time.Duration(nbDaysTo) * 24 * time.Hour

	y, m, d := time.Now().Date()
	from, err = a.parseDate(req, "from", time.Date(y, m, d, 0, 0, 0, 0, a.Location(req)).Add(defaultNbDaysFrom))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to, err = a.parseDate(req, "to", from.Add(-defaultNbDaysFrom+defaultNbDaysTo+23*time.Hour+59*time.Minute+59*time.Second))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil

}

func (a Serve) isAllType(req *http.Request) bool {
	return req.URL.Query().Get("all_types") != ""
}

func (a Serve) incidentsByParamsDate(from, to time.Time, allType bool) ([]models.Incident, error) {
	incidents, err := a.store.ByDate(from, to)
	if err != nil {
		return []models.Incident{}, err
	}
	finalIncidents := make([]models.Incident, 0)
	for _, incident := range incidents {
		if allType {
			finalIncidents = append(finalIncidents, incident)
			continue
		}
		if (incident.IsScheduled && incident.ScheduledEnd.After(time.Now())) ||
			(incident.IsScheduled && (incident.State == models.Resolved || incident.State == models.Idle)) {
			continue
		}
		finalIncidents = append(finalIncidents, incident)
	}
	sort.Sort(sort.Reverse(models.Incidents(finalIncidents)))
	return finalIncidents, nil
}

func (a Serve) timelineFormat(t time.Time) string {
	return t.Format("Jan 02, 2006")
}
