package plugin

import (
	pluginhc "github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	"github.com/sirupsen/logrus"
	"os/exec"
)

func init() {
	notifiers.RegisterNotifier(&Plugin{})
}

const RegisterFuncName = "Register"

type optsPlugin struct {
	Path   string                 `mapstructure:"path"`
	Params map[string]interface{} `mapstructure:",remain"`
}

type Plugin struct {
	notifier    Notifier
	BaseRequest Base
}

func (n Plugin) loadPlugin(path string) (Notifier, error) {
	client := pluginhc.NewClient(&pluginhc.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]pluginhc.Plugin{
			"notifier": &NotifierGRPCPlugin{},
		},
		Cmd:              exec.Command(path),
		Logger:           common.NewLogrusHclogger(logrus.StandardLogger()),
		AllowedProtocols: []pluginhc.Protocol{pluginhc.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("notifier")
	if err != nil {
		return nil, err
	}

	return raw.(Notifier), nil

}

func (n Plugin) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (notifiers.Notifier, error) {
	var opts optsPlugin
	err := mapstructure.Decode(params, &opts)
	if err != nil {
		return nil, err
	}
	p, err := n.loadPlugin(opts.Path)
	if err != nil {
		return nil, err
	}
	err = p.Init(baseInfo, opts.Params)
	if err != nil {
		return nil, err
	}
	return &Plugin{
		notifier: p,
		BaseRequest: Base{
			BaseInfo: baseInfo,
			Params:   opts.Params,
		},
	}, nil
}

func (n Plugin) Name() string {
	if n.notifier == nil {
		return "plugin"
	}
	name, err := n.notifier.Name()
	if err != nil {
		logrus.Errorf("Error from plugin: %s", err.Error())
		return "plugin"
	}
	return name
}

func (n Plugin) Id() string {
	if n.notifier == nil {
		return "plugin"
	}
	id, err := n.notifier.Id()
	if err != nil {
		logrus.Errorf("Error from plugin: %s", err.Error())
		return "plugin"
	}
	return id
}

func (n Plugin) Notify(incident models.Incident) error {
	return n.notifier.Notify(incident)
}

func (n Plugin) NotifySubscriber(incident models.Incident, subscribers []string) error {
	return n.notifier.NotifySubscriber(incident, subscribers)
}

func (n Plugin) MetadataFields() []models.MetadataField {
	fields, err := n.notifier.MetadataFields()
	if err != nil {
		logrus.Errorf("Error from plugin: %s", err.Error())
		return nil
	}
	return fields
}
