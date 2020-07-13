package notifiers

import (
	"fmt"

	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/emitter"
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

func Notify() {
	if len(toNotifies) == 0 {
		return
	}
	for event := range emitter.On() {
		incident := emitter.ToIncident(event)
		for _, toNotif := range toNotifies {
			n := toNotif.Notifier
			if !toNotif.For.MatchComponents(*incident.Components) {
				continue
			}
			entry := log.WithField("notifier", n.Name()).WithField("id", n.Id())
			err := n.Notify(incident)
			if err != nil {
				entry.Errorf("Could not send notify: %s", err.Error())
			}
		}
	}
}
