package notifiers

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . NotifierAllInOne

import (
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type Notifier interface {
	Creator(params map[string]interface{}, baseInfo config.BaseInfo) (Notifier, error)
	Name() string
	Id() string
	Notify(notifyRequest *models.NotifyRequest) error
}

type NotifierMetadataField interface {
	MetadataFields() []models.MetadataField
}

type NotifierPreCheck interface {
	PreCheck(incident models.Incident) error
}

// special interface for creating a moke
type NotifierAllInOne interface {
	Notifier
	NotifierMetadataField
	NotifierPreCheck
}
