package notifiers

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Notifier
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . NotifierSubscriber

import (
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
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

type NotifierMetadataField interface {
	MetadataFields() []models.MetadataField
}
