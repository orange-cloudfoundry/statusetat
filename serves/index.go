package serves

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type IndexData struct {
	GroupComponentState map[string]models.ComponentState
	ComponentStatesData map[string][]*ComponentStateData
	Timeline            map[string][]models.Incident
	PersistentIncidents []models.Incident
	TimelineDates       []string
	Scheduled           []models.Incident
	BaseInfo            config.BaseInfo
	Timezone            string
	Theme               config.Theme
}

func (i *IndexData) ToJsonData() interface{} {
	type JsonIncident struct {
		GUID           string    `json:"guid"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		State          string    `json:"state"`
		ComponentState string    `json:"component_state"`
		Components     []string  `json:"components"`
		Messages       []struct {
			GUID         string    `json:"guid"`
			IncidentGUID string    `json:"incident_guid"`
			CreatedAt    time.Time `json:"created_at"`
			Title        string    `json:"title"`
			Content      string    `json:"content"`
		} `json:"messages"`
		Metadata    interface{} `json:"metadata"`
		IsScheduled bool        `json:"is_scheduled"`
		Persistent  bool        `json:"persistent"`
	}

	type JsonComponent struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		State       string `json:"state"`
	}

	type JsonGroup struct {
		Name       string          `json:"name"`
		Components []JsonComponent `json:"components"`
		State      string          `json:"state"`
	}

	res := struct {
		Groups              []JsonGroup    `json:"groups"`
		PersistentIncidents []JsonIncident `json:"persistent_incidents"`
		Incidents           []JsonIncident `json:"incidents"`
		ScheduledIncidents  []JsonIncident `json:"scheduled"`
		TimeZone            string         `json:"timezone"`
	}{
		Groups:              []JsonGroup{},
		PersistentIncidents: []JsonIncident{},
		Incidents:           []JsonIncident{},
		ScheduledIncidents:  []JsonIncident{},
		TimeZone:            i.Timezone,
	}

	// Groups
	for group, state := range i.GroupComponentState {
		jsonGroup := JsonGroup{
			Name:       group,
			State:      models.TextState(state),
			Components: []JsonComponent{},
		}
		for _, component := range i.ComponentStatesData[group] {
			jsonGroup.Components = append(jsonGroup.Components, JsonComponent{
				Name:        component.Name,
				Description: component.Description,
				State:       models.TextState(component.State),
			})
		}

		incidentToJsonIncident := func(incident models.Incident) JsonIncident {
			jsonIncident := JsonIncident{
				GUID:           incident.GUID,
				CreatedAt:      incident.CreatedAt,
				UpdatedAt:      incident.UpdatedAt,
				State:          models.TextIncidentState(incident.State),
				ComponentState: models.TextState(incident.ComponentState),
				Components:     []string{},
				Messages: []struct {
					GUID         string    `json:"guid"`
					IncidentGUID string    `json:"incident_guid"`
					CreatedAt    time.Time `json:"created_at"`
					Title        string    `json:"title"`
					Content      string    `json:"content"`
				}{},
				Metadata:    incident.Metadata,
				IsScheduled: incident.IsScheduled,
				Persistent:  incident.Persistent,
			}
			for _, component := range *incident.Components {
				jsonIncident.Components = append(jsonIncident.Components, component.String())
			}
			for _, message := range incident.Messages {
				jsonIncident.Messages = append(jsonIncident.Messages, struct {
					GUID         string    `json:"guid"`
					IncidentGUID string    `json:"incident_guid"`
					CreatedAt    time.Time `json:"created_at"`
					Title        string    `json:"title"`
					Content      string    `json:"content"`
				}{
					GUID:         message.GUID,
					IncidentGUID: message.IncidentGUID,
					CreatedAt:    message.CreatedAt,
					Title:        message.Title,
					Content:      message.Content,
				})
			}
			return jsonIncident
		}

		for _, incident := range i.PersistentIncidents {
			res.PersistentIncidents = append(res.PersistentIncidents, incidentToJsonIncident(incident))
		}
		for _, incident := range i.Scheduled {
			res.ScheduledIncidents = append(res.ScheduledIncidents, incidentToJsonIncident(incident))
		}

		for _, incidents := range i.Timeline {
			for _, incident := range incidents {
				res.Incidents = append(res.Incidents, incidentToJsonIncident(incident))
			}
		}

		res.Groups = append(res.Groups, jsonGroup)
	}

	return res
}

type timeSlice []string

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	first, err := time.Parse("Jan 02, 2006", p[i])
	if err != nil {
		fmt.Println(err.Error())
	}

	second, err := time.Parse("Jan 02, 2006", p[j])
	if err != nil {
		fmt.Println(err.Error())
	}
	return first.Before(second)
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type ComponentStateData struct {
	Name        string
	Description string
	State       models.ComponentState
}

func (a *Serve) getIndexData(w http.ResponseWriter, req *http.Request) (IndexData, error) {
	from, to, err := a.periodFromReq(req, -6, 0)
	if err != nil {
		return IndexData{}, err
	}

	incidents, err := a.incidentsByParamsDate(from, to, false)
	if err != nil {
		return IndexData{}, err
	}

	componentStatesByGroup := make(map[string][]*ComponentStateData)
	componentStateMap := make(map[string]*ComponentStateData)
	for _, component := range a.config.Components {
		componentState := &ComponentStateData{
			Name:        component.Name,
			Description: component.Description,
			State:       models.Operational,
		}
		_, ok := componentStatesByGroup[component.Group]
		if ok {
			componentStatesByGroup[component.Group] = append(componentStatesByGroup[component.Group], componentState)
		} else {
			componentStatesByGroup[component.Group] = []*ComponentStateData{componentState}
		}

		componentStateMap[component.String()] = componentState
	}

	compStateGroup := make(map[string]models.ComponentState)
	for k := range a.config.Components.Regroups() {
		compStateGroup[k] = models.Operational
	}

	timeline := make(map[string][]models.Incident)
	timelineDates := make(timeSlice, 0)
	for _, incident := range incidents {
		date := a.timelineFormat(incident.CreatedAt)
		if _, ok := timeline[date]; ok {
			timeline[date] = append(timeline[date], incident)
		} else {
			timeline[date] = []models.Incident{incident}
			timelineDates = append(timelineDates, date)
		}
		if incident.State == models.Resolved || incident.Components == nil {
			continue
		}

		for _, component := range *incident.Components {
			componentState, ok := componentStateMap[component.String()]
			if !ok {
				continue
			}
			if incident.ComponentState > componentState.State {
				componentState.State = incident.ComponentState
			}

			if compGroupState, ok := compStateGroup[component.Group]; ok && incident.ComponentState > compGroupState {
				compStateGroup[component.Group] = incident.ComponentState
			}
		}

	}

	for i := 0; i < 7; i++ {
		date := a.timelineFormat(from.AddDate(0, 0, i))
		if _, ok := timeline[date]; !ok {
			timeline[date] = []models.Incident{}
			timelineDates = append(timelineDates, date)
		}
	}

	fromScheduled, toScheduled, err := a.periodFromReq(req, 0, 26)
	if err != nil {
		return IndexData{}, err
	}

	scheduled, err := a.scheduled(fromScheduled, toScheduled)
	if err != nil {
		return IndexData{}, err
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	persistents, err := a.store.Persistents()
	if err != nil {
		return IndexData{}, err
	}

	sort.Sort(sort.Reverse(timelineDates))
	return IndexData{
		BaseInfo:            a.BaseInfo(),
		GroupComponentState: compStateGroup,
		ComponentStatesData: componentStatesByGroup,
		Timeline:            timeline,
		TimelineDates:       timelineDates,
		Scheduled:           scheduled,
		PersistentIncidents: persistents,
		Timezone:            timezone,
		Theme:               *a.config.Theme,
	}, nil
}

func (a *Serve) Statuses(w http.ResponseWriter, req *http.Request) {
	data, err := a.getIndexData(w, req)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data.ToJsonData()); err != nil {
		LogError(err, 500)
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a *Serve) Index(w http.ResponseWriter, req *http.Request) {
	data, err := a.getIndexData(w, req)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	err = a.xt.ExecuteTemplate(w, "incidents.gohtml", data)
	if err != nil {
		log.Println(err)
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a *Serve) ShowIncident(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	guid := v["guid"]
	incident, err := a.store.Read(guid)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}
	err = a.xt.ExecuteTemplate(w, "one_incident.gohtml", struct {
		BaseInfo config.BaseInfo
		Incident models.Incident
		Timezone string
		Theme    config.Theme
	}{
		BaseInfo: a.BaseInfo(),
		Incident: incident,
		Timezone: timezone,
		Theme:    *a.config.Theme,
	})
	if err != nil {
		log.Println(err)
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}

func (a *Serve) History(w http.ResponseWriter, req *http.Request) {
	from, to, err := a.periodFromReq(req, -6, 0)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	after := from.Add(7 * 24 * time.Hour)

	incidents, err := a.incidentsByParamsDate(from, to, a.isAllType(req))
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timeline := make(map[string][]models.Incident)
	timelineDates := make(timeSlice, 0)
	for _, incident := range incidents {
		date := a.timelineFormat(incident.CreatedAt)
		if _, ok := timeline[date]; ok {
			timeline[date] = append(timeline[date], incident)
		} else {
			timeline[date] = []models.Incident{incident}
			timelineDates = append(timelineDates, date)
		}
		if incident.State == models.Resolved || incident.Components == nil {
			continue
		}

	}

	var before time.Time
	var subDate time.Time
	for i := 0; i < 7; i++ {
		subDate = from.AddDate(0, 0, i)
		date := a.timelineFormat(subDate)
		if _, ok := timeline[date]; !ok {
			timeline[date] = []models.Incident{}
			timelineDates = append(timelineDates, date)
		}
	}
	before = from.AddDate(0, 0, -7)

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	sort.Sort(sort.Reverse(timelineDates))
	err = a.xt.ExecuteTemplate(w, "history.gohtml", struct {
		IndexData
		Before time.Time
		After  time.Time
	}{
		IndexData: IndexData{
			BaseInfo:      a.BaseInfo(),
			Timeline:      timeline,
			TimelineDates: timelineDates,
			Timezone:      timezone,
			Theme:         *a.config.Theme,
		},
		After:  after,
		Before: before,
	})
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
}
