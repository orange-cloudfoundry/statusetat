package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ArthurHlt/statusetat/common"
	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/models"
	"github.com/mitchellh/mapstructure"
)

type SlackOpts struct {
	Endpoint           string `mapstructure:"endpoint"`
	Channel            string `mapstructure:"channel"`
	Username           string `mapstructure:"username"`
	IconEmojiIncident  string `mapstructure:"icon_emoji_incident"`
	IconEmojiScheduled string `mapstructure:"icon_emoji_scheduled"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

type slackRequest struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	LinkNames   bool              `json:"link_names,omitempty"`
	Attachments []slackAttachment `json:"attachments"`
}

// slackAttachment is used to display a richly-formatted message block.
type slackAttachment struct {
	Title      string        `json:"title,omitempty"`
	TitleLink  string        `json:"title_link,omitempty"`
	Pretext    string        `json:"pretext,omitempty"`
	Text       string        `json:"text"`
	Fallback   string        `json:"fallback"`
	CallbackID string        `json:"callback_id"`
	Fields     []slackField  `json:"fields,omitempty"`
	Actions    []slackAction `json:"actions,omitempty"`
	ImageURL   string        `json:"image_url,omitempty"`
	ThumbURL   string        `json:"thumb_url,omitempty"`
	Footer     string        `json:"footer"`
	Color      string        `json:"color,omitempty"`
}

type slackField struct {
	Title string `yaml:"title,omitempty" json:"title,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	Short *bool  `yaml:"short,omitempty" json:"short,omitempty"`
}

type slackAction struct {
	Type         string                  `yaml:"type,omitempty"  json:"type,omitempty"`
	Text         string                  `yaml:"text,omitempty"  json:"text,omitempty"`
	URL          string                  `yaml:"url,omitempty"   json:"url,omitempty"`
	Style        string                  `yaml:"style,omitempty" json:"style,omitempty"`
	Name         string                  `yaml:"name,omitempty"  json:"name,omitempty"`
	Value        string                  `yaml:"value,omitempty"  json:"value,omitempty"`
	ConfirmField *SlackConfirmationField `yaml:"confirm,omitempty"  json:"confirm,omitempty"`
}

type SlackConfirmationField struct {
	Text        string `yaml:"text,omitempty"  json:"text,omitempty"`
	Title       string `yaml:"title,omitempty"  json:"title,omitempty"`
	OkText      string `yaml:"ok_text,omitempty"  json:"ok_text,omitempty"`
	DismissText string `yaml:"dismiss_text,omitempty"  json:"dismiss_text,omitempty"`
}

type Slack struct {
	baseUrl    string
	httpClient *http.Client
	id         string
	opts       SlackOpts
	loc        *time.Location
}

func (n Slack) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (Notifier, error) {
	var opts SlackOpts
	err := mapstructure.Decode(params, &opts)
	if err != nil {
		return nil, err
	}
	if opts.Username == "" {
		opts.Username = baseInfo.Title
	}

	loc, err := time.LoadLocation(baseInfo.TimeZone)
	if err != nil {
		return nil, err
	}

	return &Slack{
		baseUrl: baseInfo.BaseURL,
		httpClient: &http.Client{
			Transport: common.MakeHttpTransport(opts.InsecureSkipVerify),
			Timeout:   5 * time.Second,
		},
		id:   opts.Endpoint,
		opts: opts,
		loc:  loc,
	}, nil
}

func (n Slack) Name() string {
	return "slack"
}

func (n Slack) Id() string {
	return n.id
}

func (n Slack) colorState(state models.ComponentState) string {
	switch state {
	case models.DegradedPerformance:
		return "#9c27b0"
	case models.PartialOutage:
		return "#ff5722"
	case models.UnderMaintenance:
		return "#9e9e9e"
	case models.MajorOutage:
		return "#e51c23"
	}
	return "#4CAF50"
}

func (n Slack) colorStateIncident(state models.IncidentState) string {
	switch state {
	case models.Unresolved:
		return "#ff5722"
	case models.Monitoring:
		return "#2196F3"
	}
	return "#4CAF50"
}

func (n Slack) Notify(incident models.Incident) error {
	if incident.IsScheduled {
		return n.notifyScheduled(incident)
	}
	return n.notifyIncident(incident)
}

func (n Slack) notifyScheduled(incident models.Incident) error {
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}
	msg := incident.MainMessage()
	color := "#607d8b"
	pretext := fmt.Sprintf("Maintenance has been scheduled, follow it at [%s](%s).", n.baseUrl, n.baseUrl)
	if incident.State == models.Resolved {
		pretext = "Maintenance is finished."
		msg = incident.LastMessage()
		color = "#4CAF50"
	}
	icon := n.opts.IconEmojiScheduled
	if icon == "" {
		icon = "clock1"
	}
	short := true
	b, _ := json.Marshal(slackRequest{
		Channel:   n.opts.Channel,
		Username:  fmt.Sprintf("%s - Scheduled maintenance", n.opts.Username),
		IconEmoji: icon,
		IconURL:   "",
		LinkNames: false,
		Attachments: []slackAttachment{
			{
				Title:   common.Title(msg.Title),
				Pretext: pretext,
				Text:    msg.Content,
				Fields: []slackField{
					{
						Title: "Components placed in maintenance",
						Value: "`" + strings.Join(incident.Components.Inline(), "`, `") + "`",
						Short: nil,
					},
					{
						Title: "Scheduled at",
						Value: incident.CreatedAt.In(n.loc).String(),
						Short: &short,
					},
					{
						Title: "Duration",
						Value: fmt.Sprintf("%s (Finish at %s)",
							common.HumanDuration(incident.CreatedAt, incident.ScheduledEnd),
							incident.CreatedAt.In(n.loc).String(),
						),
						Short: &short,
					},
				},
				Color: color,
			},
		},
	})
	req, err := http.NewRequest(http.MethodPost, n.opts.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Get error code %d", resp.StatusCode)
		}
		return fmt.Errorf("Get error code %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (n Slack) notifyIncident(incident models.Incident) error {
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}
	msg := incident.MainMessage()
	pretext := fmt.Sprintf("Incident is firing, follow it at [%s](%s).", n.baseUrl, n.baseUrl)
	if incident.State == models.Resolved {
		pretext = "Incident has been resolved"
		msg = incident.LastMessage()
	}
	short := true
	icon := n.opts.IconEmojiIncident
	if icon == "" {
		icon = "bell"
	}
	title := common.Title(msg.Title) + " - " + common.Title(models.TextIncidentState(incident.State))

	fields := []slackField{
		{
			Title: "Impacted components",
			Value: "`" + strings.Join(incident.Components.Inline(), "` `") + "`",
			Short: &short,
		},
		{
			Title: "Impact",
			Value: models.TextState(incident.ComponentState),
			Short: &short,
		},
		{
			Title: "Trigger at",
			Value: incident.CreatedAt.In(n.loc).String(),
			Short: &short,
		},
	}

	if incident.State == models.Resolved {
		fields = append(fields, slackField{
			Title: "End at",
			Value: incident.UpdatedAt.In(n.loc).String(),
			Short: &short,
		})
	}

	b, _ := json.Marshal(slackRequest{
		Channel:   n.opts.Channel,
		Username:  fmt.Sprintf("%s - Incident", n.opts.Username),
		IconEmoji: icon,
		IconURL:   "",
		LinkNames: false,
		Attachments: []slackAttachment{
			{
				Title:     title,
				TitleLink: fmt.Sprintf("%s/incidents/%s", n.baseUrl, incident.GUID),
				Pretext:   pretext,
				Text:      msg.Content,
				Fields:    fields,
				Color:     n.colorStateIncident(incident.State),
			},
		},
	})
	req, err := http.NewRequest(http.MethodPost, n.opts.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = common.ExtractHttpError(resp)
	if err != nil {
		return err
	}
	return nil
}
