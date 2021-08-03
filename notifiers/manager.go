package notifiers

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/emitter"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/storages"
)

type ToNotifie struct {
	Notifier Notifier
	For      config.ForComponent
}

var toNotifies = []ToNotifie{}

var notifiers = []Notifier{}

var metadataFields = make(models.MetadataFields, 0)

func RegisterNotifier(notifier Notifier) {
	notifiers = append(notifiers, notifier)
}

func AddNotifier(name string, params map[string]interface{}, forComp config.ForComponent, baseInfo config.BaseInfo) error {
	for _, n := range notifiers {
		if n.Name() == name {
			notifier, err := n.Creator(params, baseInfo)
			if err != nil {
				return err
			}
			if metanotif, ok := notifier.(NotifierMetadataField); ok {

				for _, field := range metanotif.MetadataFields() {
					err := field.Validate()
					if err != nil {
						return err
					}
					metadataFields = append(metadataFields, field)
				}
			}
			toNotifies = append(toNotifies, ToNotifie{
				Notifier: notifier,
				For:      forComp,
			})
			return nil
		}
	}
	return fmt.Errorf("Could not find notifier with name '%s' .", name)
}

func NotifiersMetadataFields() models.MetadataFields {
	return metadataFields
}

func NotifiersPreCheckers(components models.Components) []NotifierPreCheck {
	notifiers := make([]NotifierPreCheck, 0)
	for _, tn := range toNotifies {
		preChecker, ok := tn.Notifier.(NotifierPreCheck)
		if !ok || !tn.For.MatchComponents(components) {
			continue
		}
		notifiers = append(notifiers, preChecker)
	}
	return notifiers
}

func Notify(store storages.Store) {
	if len(toNotifies) == 0 {
		return
	}
	for event := range emitter.On() {
		subscribers, err := store.Subscribers()
		if err != nil {
			log.Warningf("Could not retrieve list of subscribers: %s", err.Error())
		}
		incident := emitter.ToIncident(event)
		for _, toNotif := range toNotifies {
			n := toNotif.Notifier
			if !toNotif.For.MatchComponents(*incident.Components) {
				continue
			}
			entry := log.WithField("notifier", n.Name()).WithField("id", n.Id())
			if snotif, ok := n.(NotifierSubscriber); ok {
				err := snotif.NotifySubscriber(incident, subscribers)
				if err != nil {
					entry.Errorf("Could not send notify to subscribers: %s", err.Error())
				}
			}
			err := n.Notify(incident)
			if err != nil {
				entry.Errorf("Could not send notify: %s", err.Error())
			}
		}
	}
}
