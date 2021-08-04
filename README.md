# Statusetat [![ginkgo test badge](https://github.com/orange-cloudfoundry/statusetat/workflows/ginkgo-test/badge.svg)](https://github.com/orange-cloudfoundry/statusetat/actions?query=workflow%3Aginkgo-test)

status page with very HA in mind and notification system to send information about incident(s) or scheduled task(s) 
to external system like slack, email or whatever plugin you implement.

[![home one incident and maintenance](/screenshots/home_one_incident_and_maintenance.png)](/screenshots/home_one_incident_and_maintenance.png)

Features:
- Easy admin web interface to manage incidents and scheduled tasks
- Send notifications to external system (slack, email, plugins...) when incident come
- Full rest api to create/read/update/delete incident/scheduled tasks
- Subscribes systems

## Why another status page ?

A status page **must** be more resilient than the service you want to be monitored because you need to have the page 
always available even when your a part of your infra is partially down to give the information to your client.

When you're a hosting service be more resilient on your status page become complicated. 
This implementation allow you to set multiple storage system in same time to ensure you always can store/retrieve incidents and 
scheduled tasks.

Statusetat will replicate in all storage what you write and ensure that the data can be retrieved.

## Getting started

1. Download latest release for your system
2. Create a `config.yml` file with this content for a dev deployment

```yaml
listen: 0.0.0.0:8080 # this will listen in http on port 8080
targets:
  - sqlite://./store.db
log:
  level: debug
components:
  - name: my-service
    description: description of my service
    group: ~
notifiers:
- type: slack
  params:
    endpoint: http://my-mattermost.com/hooks/aj6na7hz5ib1mfs3eggc1bu8tk
- type: plugin
  params:
    path: /path/to/my/plugin.so
username: admin
password: admin
cookie_key: encryption-key-for-cookies
base_info:
  base_url: http://localhost:8080
  support: https://doc.to.my.service.com
  contact: http://my-mattermost.com/channel/town
  time_zone: UTC
```

## Configuration

For understanding config definition format:
- `[]` means optional (by default parameter is required)
- `<>` means type to use

### Root configuration in config.yml


```yaml
# Listen address for listening in http
# If empty and env var `PORT` set default listen will 0.0.0.0:${PORT}
[ listen: <string> | default = 0.0.0.0:8080 ]
log:
  # log level to use for server
  # you can chose: `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `panic`
  [ level: <string> | default = info ]
  # Set to true to force not have color when seeing logs
  [ no_color: <bool> ]
  # et to true to see logs as json format
  [ in_json: <bool> ]
# username for basic authentication to access admin page or api
username: <string>
# password for basic authentication to access admin page or api
password: <string>
# cookie key for cookie encryption (generate a random value and set it here)
cookie_key: <string>
base_info:
  # Url of your statusetat for forging urls on notifications
  base_url: <string>
  # Title of your status page
  [ title: <string> | default = "Statusetat" ]
  # Url for your user to know where they can found support/doc page
  [ support: <string> ]
  # Url for your user to know where they can contact you
  [ contact: <bool> ]
  # Default timezone when creating/updating/show incident(s)
  [ timezone: <string> | default = "UTC" ]
# customize web page
theme:
  # Markdown content to show before status boxes
  [ pre_status: <string> ]
  # Markdown content to show after status boxes
  [ post_status: <string> ]
  # Markdown content to show before timeline
  [ pre_timeline: <string> ]
  # Markdown content to show after timeline
  [ post_timeline: <string> ]
  # Markdown content to show before maintenance/scheduled tasks box
  [ pre_maintenance: <string> ]
  # Markdown content to show after maintenance/scheduled tasks box
  [ post_maintenance: <string> ]
  # Markdown content to put in footer
  [ footer: <string> ]
  # Persistent name to display, default to Persistent Incident but you could set Known Issues for example
  [ persistent_display_name: <string | default = "Persistent Incident"> ]
  # Markdown content to show before persistent incident tasks box
  [ pre_persistent: <string> ]
  # Markdown content to show after persistent incident tasks box
  [ post_persistent: <string> ]
# list of targets store to use, this can be in form of:
# mysql://user:password@host:3306?options
# mariadb://user:password@host:3306?options
# postgres://user:password@host:3306?options
# sqlite://:memory:
# sqlite:///a/path
# s3://access_key_id:access_key_secret@host.com/bucket?region=us-east-1&insecure-skip-verify=false
# file:///a/path/to/a/folder
targets:
- <uri>
notifiers:
[ - <notifier> ]
```

### notifiers configuration

```yaml
# type of notifier to use, you can found them in notifiers section of this doc
type: <string>
# only trigger for particular component(s)
[ for: <for_component> ]
# map of params for the notifier you use
params:
  [ <string>: <any> ]
```

### for_component configurations

```yaml
# Set to true to mandatory match all filter
# Think that you want to use an AND condition on true and OR condition on false
[require_all: <bool> | default = false ]
# Group which match component(s)
groups:
[ - <string> ]
# Names which match component(s)
names:
[ - <string> ]
```

## Notifiers

### Slack

Notify on a slack channel or mattermost channel new incidents or scheduled tasks.

- **Type name**: `slack`
- **Params**:

```yaml
# incoming endpoint url
endpoint: <string>
# Specify channel to send notif
[ channel: <string> ]
# Specify username to show on message
[ channel: <string> ]
# Specify icon emoji incident to show on message
[ icon_emoji_incident: <string> | default = "bell" ]
# Specify icon emoji scheduled tasks to show when planified
[ icon_emoji_scheduled: <string> | default = "clock1" ]
# Specify icon emoji scheduled tasks to show when started
[ icon_emoji_scheduled_started: <string> | default = "clock1" ]
# Specify icon emoji scheduled tasks to show when finished
[ icon_emoji_scheduled_finish: <string> | default = "clock1" ]
# skip ssl verification when sending notification
[ insecure_skip_verify: <bool> ]
```

### Grafana annotation

Put a grafana annotation when incident start and when it's finished on a graph.

- **Type name**: `grafana_annotation`
- **Params**:

```yaml
# Grafana api endpoint
endpoint: <string>
# Grafana api key to use
api_key: <string>
# Dashboard id to set annotation
dashboard_id: <string>
# Panel id in dashboard to set annotation
panel_id: <string>
# tag used in grafana to filter global annotations
[ tag: <string> ]
# time zone to use when sending annotation
[ time_zone: <string> ]
# skip ssl verification when sending notification
[ insecure_skip_verify: <bool> ]
```

### Email

Send incidents and scheduled tasks to subscribers via subscribe api and/or emails in notifier param.
**Type name**: `email`
**Params**:

```yaml
# Email server host
host: <string>
# Email server port
[ port: <int> | default = 25 ]
# Username to use when connect to email server
[ username: <string> ]
# Password to use when connect to email server
[ password: <string> ]
# Set to true to use ssl/tls when connect to email server
[ use_ssl: <bool> ]
# Email to insert in from field
[ From: <string> | default = "no-reply@local" ]
# List of emails address to send email, e.g.: admins of the monitored service
subscribers:
[ - <string> ]
# Go template to use to write subject for incident
[ subject_incident: <string> | default = "[{{ .TitleSite }} {{ .Incident.State | textIncidentState | title }} Incident] {{ .IncidentTitle | title }}" ]
# Go template to use to write subject for scheduled tasks
[ subject_scheduled: <string> | default = "[{{ .TitleSite }} Scheduled task] {{ .IncidentTitle | title }}" ]
# Go template to use to write content of email when get an incident
[ txt_incident: <string | default = "see https://github.com/orange-cloudfoundry/statusetat/blob/master/notifiers/email/email.go#L26-L38"]
# Go template to use to write content of email when get an scheduled task
[ txt_incident: <string | default = "see https://github.com/orange-cloudfoundry/statusetat/blob/master/notifiers/email/email.go#L39-L51"]
# skip ssl verification when sending notification
[ insecure_skip_verify: <bool> ]
```

For customize content of email you can see data passed to template at https://github.com/orange-cloudfoundry/statusetat/blob/master/notifiers/email/email.go#L223-L236 
and template function available at https://github.com/orange-cloudfoundry/statusetat/blob/master/extemplate/template.go#L25-L51

## Plugin

You can create a plugin for notification for your own need.

Plugin are grpc golang plugin which doc can be found here: https://github.com/hashicorp/go-plugin .

You must implement [plugin.Notifier](/notifiers/plugin/interface.go).

Example of implementation:

```go
package main

import (
	pluginhc "github.com/hashicorp/go-plugin"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers/plugin"
	log "github.com/sirupsen/logrus"
	"strings"
)

type LogNotifier struct {
}

func (n LogNotifier) PreCheck(incident models.Incident) error {
	return nil
}

func (n *LogNotifier) Init(baseInfo config.BaseInfo, params map[string]interface{}) error {
	return nil
}

func (n LogNotifier) Name() (string, error) {
	return "log_notifier", nil
}

func (n LogNotifier) Id() (string, error) {
	return "log_notifier1", nil
}

func (n LogNotifier) MetadataFields() ([]models.MetadataField, error) {
	return []models.MetadataField{}, nil
}

func (n LogNotifier) Notify(notifyReq *models.NotifyRequest) error {
	log.
		WithField("subscribers", strings.Join(notifyReq.Subscribers, ", ")).
		WithField("triggerred_by_user", notifyReq.TriggerByUser).
		Info(notifyReq.Incident)
	return nil
}

func main() {
	pluginhc.Serve(&pluginhc.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]pluginhc.Plugin{
			"notifier": &plugin.NotifierGRPCPlugin{
				Impl: &LogNotifier{},
			},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: pluginhc.DefaultGRPCServer,
	})
}
```

You can build it with command line `go build -o log-notif .`

You can now use it in your config with:

- **Type name**: `plugin` or notifier name you have set in func signature `Name() string`
- **Params**:
```yaml
# path to .so to load
# in our example it would be `./log-notif`
path: <string>
# all remaining arguments will be pass at Init
```


## Api 

To document

## Credits

This project was heavily inspired by [statusfy](https://github.com/juliomrqz/statusfy) mostly on the design part and 
models of incidents/components.

[Cachethq](https://cachethq.io/) is another source of inspiration for some details on the admin design part.
