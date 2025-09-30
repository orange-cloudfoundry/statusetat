package notifiers

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/statusetat/v2/config"
	"github.com/orange-cloudfoundry/statusetat/v2/emitter"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
	"github.com/orange-cloudfoundry/statusetat/v2/storages"
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
	return fmt.Errorf("could not find notifier with name '%s'", name)
}

func MetadataFields() models.MetadataFields {
	return metadataFields
}

func PreCheckers(components models.Components) []NotifierPreCheck {
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

func ListAll() map[string][]Notifier {
	notifierMap := make(map[string][]Notifier)
	for _, toNotif := range toNotifies {
		notifier := toNotif.Notifier
		if ls, ok := notifierMap[notifier.Name()]; ok {
			notifierMap[notifier.Name()] = append(ls, notifier)
			continue
		}
		notifierMap[notifier.Name()] = []Notifier{notifier}
	}
	return notifierMap
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
		notifyReq := emitter.ToNotifyRequest(event)
		notifyReq.Subscribers = subscribers

		// Use a wait group to make notify calls concurrently and wait for all to complete
		var wg sync.WaitGroup
		for _, toNotif := range toNotifies {
			n := toNotif.Notifier
			if !toNotif.For.MatchComponents(*notifyReq.Incident.Components) {
				continue
			}
			wg.Add(1)
			go func(n Notifier, notifyReq *models.NotifyRequest) {
				defer wg.Done()
				entry := log.WithField("notifier", n.Name()).WithField("id", n.Id())
				err := n.Notify(notifyReq)
				if err != nil {
					entry.Errorf("could not send notify: %s", err.Error())
				}
			}(n, notifyReq)
		}
		wg.Wait()
	}
}
