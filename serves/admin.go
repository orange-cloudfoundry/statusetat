package serves

import (
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
)

var adminMenuItems = []string{"incident", "maintenance"}

type adminDefaultData struct {
	BaseInfo   config.BaseInfo
	MenuItems  []string
	ActiveItem string
	Timezone   string
}

func (a Serve) AdminIncidents(w http.ResponseWriter, req *http.Request) {
	from, to, err := a.periodFromReq(req, -6, 0)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	after := from.Add(7 * 24 * time.Hour)
	before := from.AddDate(0, 0, -7)

	incidents, err := a.incidentsByParamsDate(from, to, false)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	err = a.xt.ExecuteTemplate(w, "admin/incidents.gohtml", struct {
		adminDefaultData
		Incidents      []models.Incident
		IncidentStates []models.IncidentState
		Before         time.Time
		After          time.Time
		From           time.Time
		To             time.Time
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
		From:   from,
		To:     to,
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

	err = a.xt.ExecuteTemplate(w, "admin/add_edit_"+typ+".gohtml", struct {
		adminDefaultData
		Components      []string
		IncidentStates  []models.IncidentState
		ComponentStates []models.ComponentState
		Incident        models.Incident
		MetadataFields  models.MetadataFields
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
		MetadataFields:  notifiers.NotifiersMetadataFields(),
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a Serve) AdminMaintenance(w http.ResponseWriter, req *http.Request) {
	from, to, err := a.periodFromReq(req, -26, 26)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	after := from.Add(27 * 24 * time.Hour)
	before := from.AddDate(0, 0, -27)

	maintenance, err := a.scheduled(from, to)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	err = a.xt.ExecuteTemplate(w, "admin/maintenance.gohtml", struct {
		adminDefaultData
		Maintenance    []models.Incident
		IncidentStates []models.IncidentState
		MetadataFields models.MetadataFields
		Before         time.Time
		After          time.Time
		From           time.Time
		To             time.Time
	}{
		adminDefaultData: adminDefaultData{
			BaseInfo:   a.baseInfo,
			ActiveItem: "maintenance",
			MenuItems:  adminMenuItems,
			Timezone:   timezone,
		},
		Maintenance:    maintenance,
		IncidentStates: models.AllIncidentState,
		MetadataFields: notifiers.NotifiersMetadataFields(),
		After:          after,
		Before:         before,
		From:           from,
		To:             to,
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a Serve) AdminAddEditMaintenance(w http.ResponseWriter, req *http.Request) {
	a.AdminAddEditIncidentByType(w, req, "maintenance")
}
