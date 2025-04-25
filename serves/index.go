package serves

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"

	"github.com/orange-cloudfoundry/statusetat/v2/config"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
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

func (a *Serve) Index(w http.ResponseWriter, req *http.Request) {

	from, to, err := a.periodFromReq(req, -6, 0)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	incidents, err := a.incidentsByParamsDate(from, to, false)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
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
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	scheduled, err := a.scheduled(fromScheduled, toScheduled)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	timezone := ""
	if !a.IsDefaultLocation(req) {
		timezone = a.Location(req).String()
	}

	persistents, err := a.store.Persistents()
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}

	sort.Sort(sort.Reverse(timelineDates))
	err = a.xt.ExecuteTemplate(w, "incidents.gohtml", IndexData{
		BaseInfo:            a.BaseInfo(),
		GroupComponentState: compStateGroup,
		ComponentStatesData: componentStatesByGroup,
		Timeline:            timeline,
		TimelineDates:       timelineDates,
		Scheduled:           scheduled,
		PersistentIncidents: persistents,
		Timezone:            timezone,
		Theme:               *a.config.Theme,
	})
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
