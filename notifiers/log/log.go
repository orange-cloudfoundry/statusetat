package log

import (
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	"github.com/sirupsen/logrus"
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

func (n Log) Notify(incident models.Incident) error {
	IncidentToEntry(incident).Info("notify")
	return nil
}

func (n Log) NotifySubscriber(incident models.Incident, subscriber []string) error {
	if len(subscriber) == 0 {
		return nil
	}
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}

	for _, sub := range subscriber {
		IncidentToEntry(incident).WithField("subscriber", sub).Info("notify subscribers")
	}
	return nil
}

func IncidentToEntry(incident models.Incident) *logrus.Entry {
	entryMap := map[string]interface{}{
		"main_title":      incident.MainMessage().Title,
		"last_title":      incident.LastMessage().Title,
		"last_message":    incident.LastMessage().Content,
		"created_at":      incident.CreatedAt,
		"updated_at":      incident.UpdatedAt,
		"component_state": models.TextState(incident.ComponentState),
		"incident_state":  models.TextIncidentState(incident.State),
		"is_scheduled":    incident.IsScheduled,
	}
	if incident.IsScheduled {
		entryMap["scheduled_end"] = incident.ScheduledEnd
	}
	return logrus.WithFields(entryMap)
}
