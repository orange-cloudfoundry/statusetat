package log

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
)

func init() {
	notifiers.RegisterNotifier(&Log{})
}

type Log struct {
}

func (n Log) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (notifiers.Notifier, error) {
	return &Log{}, nil
}

func (n Log) Name() string {
	return "log"
}

func (n Log) Id() string {
	return "log"
}

func (n Log) Notify(notifyReq *models.NotifyRequest) error {
	RequestToEntry(notifyReq).Info("notify")
	return nil
}

func RequestToEntry(notifyReq *models.NotifyRequest) *logrus.Entry {

	incident := notifyReq.Incident
	entryMap := map[string]interface{}{
		"main_title":         incident.MainMessage().Title,
		"last_title":         incident.LastMessage().Title,
		"last_message":       incident.LastMessage().Content,
		"created_at":         incident.CreatedAt,
		"updated_at":         incident.UpdatedAt,
		"component_state":    models.TextState(incident.ComponentState),
		"incident_state":     models.TextIncidentState(incident.State),
		"is_scheduled":       incident.IsScheduled,
		"triggerred_by_user": notifyReq.TriggerByUser,
		"nb_subscribers":     len(notifyReq.Subscribers),
	}
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		entryMap["subscribers"] = strings.Join(notifyReq.Subscribers, ", ")
	}

	if incident.IsScheduled {
		entryMap["scheduled_end"] = incident.ScheduledEnd
	}
	return logrus.WithFields(entryMap)
}
