package notifiers

import (
	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/models"
)

type Notifier interface {
	Creator(params map[string]interface{}, baseInfo config.BaseInfo) (Notifier, error)
	Name() string
	Id() string
	Notify(incident models.Incident) error
}

type NotifierSubscriber interface {
	NotifySubscriber(incident models.Incident, subscribers []string) error
}
