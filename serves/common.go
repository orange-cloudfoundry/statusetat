package serves

import (
	"net/http"
	"sort"
	"time"

	"github.com/orange-cloudfoundry/statusetat/models"
)

func (a Serve) scheduled(req *http.Request) ([]models.Incident, error) {
	scheduled := make([]models.Incident, 0)

	y, m, d := time.Now().Date()
	from, err := a.parseDate(req, "from", time.Date(y, m, d, 0, 0, 0, 0, a.Location(req)))
	if err != nil {
		return []models.Incident{}, err
	}

	to, err := a.parseDate(req, "to", from.AddDate(0, 0, 26))
	if err != nil {
		return []models.Incident{}, err
	}

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

func (a Serve) incidentsByParamsDate(req *http.Request) ([]models.Incident, error) {
	var err error
	y, m, d := time.Now().Date()
	from, err := a.parseDate(req, "from", time.Date(y, m, d, 0, 0, 0, 0, a.Location(req)))
	if err != nil {
		return []models.Incident{}, err
	}

	to, err := a.parseDate(req, "to", from.Add(7*24*time.Hour))
	if err != nil {
		return []models.Incident{}, err
	}

	allType := false
	dateQuery := req.URL.Query().Get("all_types")
	if dateQuery != "" {
		allType = true
	}

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
