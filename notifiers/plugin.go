package notifiers

import (
	"fmt"
	"plugin"

	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/models"
	"github.com/mitchellh/mapstructure"
)

const RegisterFuncName = "Register"

type optsPlugin struct {
	Path   string                 `mapstructure:"path"`
	Params map[string]interface{} `mapstructure:",remain"`
}

type Plugin struct {
	notifier Notifier
}

func (n Plugin) loadPlugin(path string) (Notifier, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error on plugin %s: %s", path, err.Error())
	}
	registerPlugin, err := p.Lookup(RegisterFuncName)
	if err != nil {
		return nil, fmt.Errorf("Error on plugin %s: %s", path, err.Error())
	}
	notifierPlugin := registerPlugin.(func() Notifier)()
	name := notifierPlugin.Name()
	if name == "" {
		return nil, fmt.Errorf("Error on plugin %s: plugin must define its name.")
	}

	return notifierPlugin, nil

}

func (n Plugin) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (Notifier, error) {
	var opts optsPlugin
	err := mapstructure.Decode(params, &opts)
	if err != nil {
		return nil, err
	}
	p, err := n.loadPlugin(opts.Path)
	if err != nil {
		return nil, err
	}
	notifier, err := p.Creator(opts.Params, baseInfo)
	if err != nil {
		return nil, err
	}
	return &Plugin{
		notifier: notifier,
	}, nil
}

func (n Plugin) Name() string {
	if n.notifier == nil {
		return "plugin"
	}
	return n.notifier.Name()
}

func (n Plugin) Id() string {
	return n.notifier.Id()
}

func (n Plugin) Notify(incident models.Incident) error {
	return n.notifier.Notify(incident)
}
