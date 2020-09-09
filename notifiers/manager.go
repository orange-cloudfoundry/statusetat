package notifiers

import (
	"fmt"

	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/emitter"
	"github.com/ArthurHlt/statusetat/storages"
	log "github.com/sirupsen/logrus"
)

type ToNotifie struct {
	Notifier Notifier
	For      config.ForComponent
}

var toNotifies = []ToNotifie{}

var notifiers = []Notifier{}

func RegisterNotifier(notifier Notifier) {
	notifiers = append(notifiers, notifier)
}

func AddNotifier(name string, params map[string]interface{}, forComp config.ForComponent, baseInfo config.BaseInfo) error {
	for _, n := range notifiers {
		if n.Name() == name {
			toNotify, err := n.Creator(params, baseInfo)
			if err != nil {
				return err
			}
			toNotifies = append(toNotifies, ToNotifie{
				Notifier: toNotify,
				For:      forComp,
			})
			return nil
		}
	}
	return fmt.Errorf("Could not find notifier with name '%s' .", name)
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
