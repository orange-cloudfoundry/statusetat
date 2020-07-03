package serves

import (
	"net/http"
	"time"

	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/models"
	"github.com/gorilla/mux"
)

var adminMenuItems = []string{"incident", "maintenance"}

type adminDefaultData struct {
	BaseInfo   config.BaseInfo
	MenuItems  []string
	ActiveItem string
	Timezone   string
}

func (a Serve) AdminIncidents(w http.ResponseWriter, req *http.Request) {
	loc := a.Location(req)
	from, err := a.parseDate(req, "from", time.Now().In(loc))
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	after := from.Add(7 * 24 * time.Hour)
	before := from.AddDate(0, 0, -7)

	incidents, err := a.incidentsByParamsDate(req)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	err = a.xt.ExecuteTemplate(w, "admin/incidents.html", struct {
		adminDefaultData
		Incidents      []models.Incident
		IncidentStates []models.IncidentState
		Before         time.Time
		After          time.Time
	}{
		adminDefaultData: adminDefaultData{
			BaseInfo:   a.baseInfo,
			ActiveItem: "incident",
			MenuItems:  adminMenuItems,
			Timezone:   timezone,
		},
		Incidents:      incidents,
		IncidentStates: models.AllIncidentState,

		After:  after,
		Before: before,
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a Serve) AdminAddEditIncident(w http.ResponseWriter, req *http.Request) {
	a.AdminAddEditIncidentByType(w, req, "incident")
}

func (a Serve) AdminAddEditIncidentByType(w http.ResponseWriter, req *http.Request, typ string) {
	var incident models.Incident
	var err error
	v := mux.Vars(req)
	guid := v["guid"]
	if guid != "" {
		incident, err = a.store.Read(guid)
		if err != nil {
			HTMLError(w, err, http.StatusInternalServerError)
			return
		}
	} else {
		incident.ComponentState = models.MajorOutage
		incident.CreatedAt = time.Now().In(a.Location(req))
		incident.ScheduledEnd = incident.CreatedAt.Add(2 * time.Hour)
	}
	components := make([]string, len(a.components))
	for i, c := range a.components {
		components[i] = c.String()
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	err = a.xt.ExecuteTemplate(w, "admin/add_edit_"+typ+".html", struct {
		adminDefaultData
		Components      []string
		IncidentStates  []models.IncidentState
		ComponentStates []models.ComponentState
		Incident        models.Incident
	}{
		adminDefaultData: adminDefaultData{
			BaseInfo:   a.baseInfo,
			ActiveItem: typ,
			MenuItems:  adminMenuItems,
			Timezone:   timezone,
		},
		Components: components,

		IncidentStates:  models.AllIncidentState,
		ComponentStates: models.AllComponentState,
		Incident:        incident,
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a Serve) AdminMaintenance(w http.ResponseWriter, req *http.Request) {
	maintenance, err := a.scheduled(req)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}
	err = a.xt.ExecuteTemplate(w, "admin/maintenance.html", struct {
		adminDefaultData
		Maintenance    []models.Incident
		IncidentStates []models.IncidentState
	}{
		adminDefaultData: adminDefaultData{
			BaseInfo:   a.baseInfo,
			ActiveItem: "maintenance",
			MenuItems:  adminMenuItems,
			Timezone:   timezone,
		},
		Maintenance:    maintenance,
		IncidentStates: models.AllIncidentState,
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a Serve) AdminAddEditMaintenance(w http.ResponseWriter, req *http.Request) {
	a.AdminAddEditIncidentByType(w, req, "maintenance")
}
