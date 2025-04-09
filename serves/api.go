package serves

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	"github.com/nicklaw5/go-respond"

	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/emitter"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
)

func (a *Serve) CreateIncident(w http.ResponseWriter, req *http.Request) {
	guid := uuid.NewString()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	var incident models.Incident
	err = json.Unmarshal(b, &incident)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}

	if incident.Components == nil || len(*incident.Components) == 0 {
		JSONError(w, fmt.Errorf("components must be set"), http.StatusPreconditionFailed)
		return
	}

	if incident.ComponentState < 0 {
		JSONError(w, fmt.Errorf("component state must be set"), http.StatusPreconditionFailed)
		return
	}

	if incident.State < 0 {
		JSONError(w, fmt.Errorf("incident state must be set"), http.StatusPreconditionFailed)
		return
	}

	if len(incident.Messages) <= 0 {
		JSONError(w, fmt.Errorf("at least one message must be set"), http.StatusPreconditionFailed)
		return
	}

	if incident.IsScheduled && incident.ScheduledEnd.IsZero() {
		JSONError(w, fmt.Errorf("if is scheduled, it must have a scheduled end"), http.StatusPreconditionFailed)
		return
	}

	if incident.IsScheduled {
		incident.ComponentState = models.UnderMaintenance
	}

	loc := a.Location(req)

	incident.Origin = a.BaseURL()
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = time.Now().In(loc)
	}

	incident.UpdatedAt = incident.CreatedAt

	if incident.IsScheduled && incident.CreatedAt.After(incident.ScheduledEnd) {
		JSONError(w, fmt.Errorf("start date of scheduled maintenance can't be before end date"), http.StatusPreconditionFailed)
		return
	}

	incident.GUID = guid
	incident.Messages = a.messagesGuid(guid, incident.Messages, loc)

	err = a.runPreCheck(&incident)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionFailed)
		return
	}

	incident, err = a.store.Create(incident)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitter.Emit(models.NewNotifyRequest(incident, false))
	respond.NewResponse(w).Created(incident)
}

func (a *Serve) messagesGuid(incidentGuid string, messages []models.Message, loc *time.Location) []models.Message {
	for i, msg := range messages {
		msg.IncidentGUID = incidentGuid

		if msg.GUID != "" {
			messages[i] = msg
			continue
		}
		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = time.Now().In(loc)
		}
		msg.GUID = uuid.NewString()
		messages[i] = msg
	}
	return messages
}

func (a *Serve) parseDate(req *http.Request, key string, defaultTime time.Time) (time.Time, error) {
	if req == nil {
		return defaultTime, nil
	}
	var err error
	date := defaultTime
	dateQuery := req.URL.Query().Get(key)
	if dateQuery != "" {
		date, err = time.Parse(time.RFC3339, dateQuery)
		if err != nil {
			return date, err
		}
	}
	return date, nil
}

func (a *Serve) ByDate(w http.ResponseWriter, req *http.Request) {
	var err error
	from, to, err := a.periodFromReq(req, -7, 0)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	incidents, err := a.incidentsByParamsDate(from, to, a.isAllType(req))
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	respond.NewResponse(w).Ok(incidents)
}

func (a *Serve) Persistents(w http.ResponseWriter, req *http.Request) {
	var err error
	incidents, err := a.store.Persistents()
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	respond.NewResponse(w).Ok(incidents)
}

func (a *Serve) Incident(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	guid := v["guid"]
	incident, err := a.store.Read(guid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	respond.NewResponse(w).Ok(incident)
}

func (a *Serve) Update(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	guid := v["guid"]

	b, err := io.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	var incidentUpdate models.IncidentUpdateRequest
	err = json.Unmarshal(b, &incidentUpdate)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}

	incident, err := a.store.Read(guid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}

	if incidentUpdate.ComponentState != nil {
		incident.ComponentState = *incidentUpdate.ComponentState
	}

	if incidentUpdate.State != nil {
		incident.State = *incidentUpdate.State
	}

	if incidentUpdate.IsScheduled != nil {
		incident.IsScheduled = *incidentUpdate.IsScheduled
	}

	if incidentUpdate.Persistent != nil {
		incident.Persistent = *incidentUpdate.Persistent
	}

	if !incidentUpdate.ScheduledEnd.IsZero() {
		incident.ScheduledEnd = incidentUpdate.ScheduledEnd
	}

	if !incidentUpdate.CreatedAt.IsZero() {
		incident.CreatedAt = incidentUpdate.CreatedAt
	}

	if incidentUpdate.Components != nil {
		incident.Components = incidentUpdate.Components
	}

	if incidentUpdate.Messages != nil && len(*incidentUpdate.Messages) == 0 {
		incident.Messages = []models.Message{}
	}
	if incidentUpdate.Messages != nil && len(*incidentUpdate.Messages) > 0 {

		if _, ok := req.URL.Query()["partial_update_message"]; ok &&
			len(*incidentUpdate.Messages) == 1 &&
			(*incidentUpdate.Messages)[0].GUID != "" {

			mainMsg := (*incidentUpdate.Messages)[0]
			for i, msg := range incident.Messages {
				if msg.GUID == mainMsg.GUID {
					msg.Content = mainMsg.Content
					msg.Title = mainMsg.Title
					incident.Messages[i] = msg
				}
			}

		} else {
			incident.Messages = a.messagesGuid(guid, *incidentUpdate.Messages, a.Location(req))
		}
	}

	if incidentUpdate.Metadata != nil {
		incident.Metadata = *incidentUpdate.Metadata
	}
	if incident.IsScheduled && incident.CreatedAt.After(incident.ScheduledEnd) {
		JSONError(w, fmt.Errorf("start date of scheduled maintenance must be before end date"), http.StatusPreconditionFailed)
		return
	}

	incident.Origin = a.BaseURL()

	incident.UpdatedAt = time.Now()

	err = a.runPreCheck(&incident)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionFailed)
		return
	}

	incident, err = a.store.Update(guid, incident)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	if !incidentUpdate.NoNotify {
		emitter.Emit(models.NewNotifyRequest(incident, false))
	}
	respond.NewResponse(w).Ok(incident)
}

func (a *Serve) runPreCheck(incident *models.Incident) error {
	var result error
	for _, preChecker := range notifiers.PreCheckers(*incident.Components) {
		err := preChecker.PreCheck(incident)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	if result != nil {
		result.(*multierror.Error).ErrorFormat = common.ListFormatHTMLFunc
	}
	return result
}

func (a *Serve) Notify(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	guid := v["guid"]

	incident, err := a.store.Read(guid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	emitter.Emit(models.NewNotifyRequest(incident, true))
}

func (a *Serve) Delete(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	guid := v["guid"]

	err := a.store.Delete(guid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func (a *Serve) AddMessage(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	incidentGuid := v["incident_guid"]

	incident, err := a.store.Read(incidentGuid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	var message models.Message
	err = json.Unmarshal(b, &message)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}

	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now().In(a.Location(req))
	}

	message.GUID = uuid.NewString()

	incident.Messages = append(incident.Messages, message)

	incident, err = a.store.Update(incidentGuid, incident)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	incident.UpdatedAt = time.Now()

	emitter.Emit(models.NewNotifyRequest(incident, false))
	respond.NewResponse(w).Created(incident)
}

func (a *Serve) DeleteMessage(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	incidentGuid := v["incident_guid"]
	messageGuid := v["message_guid"]

	incident, err := a.store.Read(incidentGuid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	finalMessages := make(models.Messages, 0)
	for _, msg := range incident.Messages {
		if msg.GUID == messageGuid {
			continue
		}
		finalMessages = append(finalMessages, msg)
	}

	incident.Messages = finalMessages

	incident.UpdatedAt = time.Now()

	incident, err = a.store.Update(incidentGuid, incident)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	emitter.Emit(models.NewNotifyRequest(incident, false))
	respond.NewResponse(w).Ok(incident)
}

func (a *Serve) ReadMessage(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	incidentGuid := v["incident_guid"]
	messageGuid := v["message_guid"]

	incident, err := a.store.Read(incidentGuid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	for _, msg := range incident.Messages {
		if msg.GUID == messageGuid {
			respond.NewResponse(w).Ok(msg)
			return
		}
	}

	JSONError(w, fmt.Errorf("message not found"), http.StatusNotFound)
}

func (a *Serve) ReadMessages(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	incidentGuid := v["incident_guid"]

	incident, err := a.store.Read(incidentGuid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	respond.NewResponse(w).Ok(incident.Messages)
}

func (a *Serve) UpdateMessage(w http.ResponseWriter, req *http.Request) {
	v := mux.Vars(req)
	incidentGuid := v["incident_guid"]
	messageGuid := v["message_guid"]

	incident, err := a.store.Read(incidentGuid)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	var message models.Message
	err = json.Unmarshal(b, &message)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}

	for i, msg := range incident.Messages {
		if msg.GUID == messageGuid {
			msg.Content = message.Content
			msg.Title = message.Title
			incident.Messages[i] = msg
		}
	}

	incident, err = a.store.Update(incidentGuid, incident)
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	incident.UpdatedAt = time.Now()

	emitter.Emit(models.NewNotifyRequest(incident, false))
	respond.NewResponse(w).Ok(incident)
}

func (a *Serve) ShowComponents(w http.ResponseWriter, req *http.Request) {
	respond.NewResponse(w).Ok(a.config.Components.Inline())
}

func (a *Serve) ShowFlagIncidentStates(w http.ResponseWriter, req *http.Request) {
	type stateDetail struct {
		Value       models.IncidentState `json:"value"`
		Description string               `json:"description"`
	}
	states := []stateDetail{
		{
			Value:       models.Resolved,
			Description: models.TextIncidentState(models.Resolved),
		},
		{
			Value:       models.Unresolved,
			Description: models.TextIncidentState(models.Unresolved),
		},
		{
			Value:       models.Monitoring,
			Description: models.TextIncidentState(models.Monitoring),
		},
		{
			Value:       models.Idle,
			Description: models.TextIncidentState(models.Monitoring),
		},
	}
	respond.NewResponse(w).Ok(states)
}

func (a *Serve) ListSubscribers(w http.ResponseWriter, req *http.Request) {
	subs, err := a.store.Subscribers()
	if err != nil {
		if os.IsNotExist(err) {
			JSONError(w, err, http.StatusNotFound)
			return
		}
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	respond.NewResponse(w).Ok(subs)
}

func (a *Serve) ShowFlagComponentStates(w http.ResponseWriter, req *http.Request) {
	type stateDetail struct {
		Value       models.ComponentState `json:"value"`
		Description string                `json:"description"`
	}
	states := []stateDetail{
		{
			Value:       models.Operational,
			Description: models.TextState(models.Operational),
		},
		{
			Value:       models.MajorOutage,
			Description: models.TextState(models.MajorOutage),
		},
		{
			Value:       models.UnderMaintenance,
			Description: models.TextState(models.UnderMaintenance),
		},
		{
			Value:       models.PartialOutage,
			Description: models.TextState(models.PartialOutage),
		},
		{
			Value:       models.DegradedPerformance,
			Description: models.TextState(models.DegradedPerformance),
		},
	}
	respond.NewResponse(w).Ok(states)
}
