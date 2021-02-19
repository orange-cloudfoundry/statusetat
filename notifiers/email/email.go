package email

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . EmailDialer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/extemplate"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	"gopkg.in/gomail.v2"
)

func init() {
	notifiers.RegisterNotifier(&Email{})
}

const (
	DefaultSubjectIncident = "[{{ .TitleSite }} {{ .Incident.State | textIncidentState | title }} Incident] {{ .IncidentTitle | title }}"
	DefaultTxtIncident     = `<h1>{{ .IncidentTitle | title }} - <span style="color: {{ .Incident.State | colorHexIncidentState }};">{{ .Incident.State | textIncidentState | title }}</span></h1>
<ul>
	<li><b>Impacted components</b>: {{ join .Incident.Components.Inline ", " }}</li>
	<li><b>Impact</b>: {{ .Incident.ComponentState | textState }}</li>
	<li><b>Trigger at</b>: {{ .Incident.CreatedAt | timeFormat }}</li>
	<li><b>Link</b>: <a href="{{ .Link }}">{{ .Link }}</a></li>
</ul>

<h2>Message</h2>
<p>
{{ .Content | markdown }}
</p>
`
	DefaultTxtScheduled = `<h1>Scheduled Maintenance: {{ .IncidentTitle | title }}</h1>
<ul>
	<li><b>Components involved</b>: {{ join .Incident.Components.Inline ", " }}</li>
	<li><b>Scheduled at</b>: {{ .Incident.CreatedAt | timeFormat }}</li>
	<li><b>Duration</b>: {{ humanDuration .Incident.CreatedAt .Incident.ScheduledEnd }}</li>
	<li><b>Link</b>: <a href="{{ .Link }}">{{ .Link }}</a></li>
</ul>

<h2>Message</h2>
<p>
{{ .Content | markdown }}
</p>
`
	DefaultSubjectScheduled = "[{{ .TitleSite }} Scheduled task] {{ .IncidentTitle | title }}"
	DefaultFrom             = "no-reply@local"
)

type EmailDialer interface {
	DialAndSend(m ...*gomail.Message) error
}

type OptsEmail struct {
	Host             string   `mapstructure:"host"`
	Port             int      `mapstructure:"port"`
	Username         string   `mapstructure:"username"`
	Password         string   `mapstructure:"password"`
	UseSSl           bool     `mapstructure:"use_ssl"`
	SubjectIncident  string   `mapstructure:"subject_incident"`
	TxtIncident      string   `mapstructure:"txt_incident"`
	SubjectScheduled string   `mapstructure:"subject_scheduled"`
	TxtScheduled     string   `mapstructure:"txt_scheduled"`
	Subscribers      []string `mapstructure:"subscribers"`
	From             string   `mapstructure:"from"`
	SkipInsecure     bool     `mapstructure:"insecure_skip_verify"`
}

type Email struct {
	dialer              EmailDialer
	id                  string
	opts                OptsEmail
	titleSite           string
	tplSubjectIncident  *template.Template
	tplTxtIncident      *template.Template
	tplSubjectScheduled *template.Template
	tplTxtScheduled     *template.Template
	baseUrl             string
}

func (n *Email) SetDialer(dialer EmailDialer) {
	n.dialer = dialer
}

type SmtpInfo struct {
	User     string
	Password string
	Host     string
	Port     int
}

func (n Email) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (notifiers.Notifier, error) {
	var opts OptsEmail
	err := mapstructure.Decode(params, &opts)
	if err != nil {
		return nil, err
	}

	if opts.SubjectIncident == "" {
		opts.SubjectIncident = DefaultSubjectIncident
	}

	tplSubjIncident, err := template.New("subjectIncident").Funcs(extemplate.Funcs()).Parse(opts.SubjectIncident)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing template subject incident: %s", err.Error())
	}

	if opts.SubjectScheduled == "" {
		opts.SubjectScheduled = DefaultSubjectScheduled
	}
	tplSubjScheduled, err := template.New("subjectScheduled").Funcs(extemplate.Funcs()).Parse(opts.SubjectScheduled)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing template subject scheduled: %s", err.Error())
	}

	if opts.TxtIncident == "" {
		opts.TxtIncident = DefaultTxtIncident
	}
	tplTxtIncident, err := template.New("TxtIncident").Funcs(extemplate.Funcs()).Parse(opts.TxtIncident)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing template text incident: %s", err.Error())
	}
	if opts.TxtScheduled == "" {
		opts.TxtScheduled = DefaultTxtScheduled
	}
	tplTxtScheduled, err := template.New("TxtScheduled").Funcs(extemplate.Funcs()).Parse(opts.TxtScheduled)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing template text scheduled: %s", err.Error())
	}

	if opts.Host == "" {
		return nil, fmt.Errorf("Host for email is mandatory")
	}

	if opts.From == "" {
		opts.From = DefaultFrom
	}

	dialer := loadDialer(opts)

	return &Email{
		dialer:              dialer,
		id:                  opts.Host,
		opts:                opts,
		titleSite:           baseInfo.Title,
		tplSubjectIncident:  tplSubjIncident,
		tplTxtIncident:      tplTxtIncident,
		tplSubjectScheduled: tplSubjScheduled,
		tplTxtScheduled:     tplTxtScheduled,
		baseUrl:             baseInfo.BaseURL,
	}, nil
}

func loadDialer(opts OptsEmail) *gomail.Dialer {
	port := 25
	if opts.Port > 0 {
		port = opts.Port
	}
	dialer := &gomail.Dialer{
		Host:     opts.Host,
		Port:     port,
		Username: opts.Username,
		Password: opts.Password,
		SSL:      opts.UseSSl,
	}
	if opts.SkipInsecure {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: opts.SkipInsecure}
	}
	return dialer
}

func (n Email) Name() string {
	return "email"
}

func (n Email) Id() string {
	return n.id
}

func (n Email) Notify(incident models.Incident) error {
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}
	return n.NotifySubscriber(incident, n.opts.Subscribers)
}

func (n Email) NotifySubscriber(incident models.Incident, subscriber []string) error {
	if len(subscriber) == 0 {
		return nil
	}
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}
	subject, text, err := n.incidentToMail(incident)
	if err != nil {
		return err
	}
	var result error
	for _, sub := range subscriber {
		finalText := text + fmt.Sprintf(`<br/><br/>
<hr/>
<a href="%s/v1/unsubscribe?email=%s">Click here for unsubscribe to email</a>`, n.baseUrl, sub)
		err := n.sendEmailTo(subject, finalText, sub)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (n Email) incidentToMail(incident models.Incident) (subject string, textHtml string, err error) {
	title := incident.MainMessage().Title
	msg := incident.MainMessage()
	if incident.State == models.Resolved {
		msg = incident.LastMessage()
	}
	incidentStruct := struct {
		Incident        models.Incident
		TitleSite       string
		IncidentTitle   string
		Link            string
		UnsubscribeLink string
		Content         string
	}{
		IncidentTitle: title,
		Incident:      incident,
		TitleSite:     n.titleSite,
		Link:          fmt.Sprintf("%s/incidents/%s", n.baseUrl, incident.GUID),
		Content:       msg.Content,
	}

	subjTpl := n.tplSubjectIncident
	txtTpl := n.tplTxtIncident
	if incident.IsScheduled {
		subjTpl = n.tplSubjectScheduled
		txtTpl = n.tplTxtScheduled
	}

	buf := &bytes.Buffer{}
	err = subjTpl.Execute(buf, incidentStruct)
	if err != nil {
		return "", "", err
	}
	subject = buf.String()
	buf.Reset()

	err = txtTpl.Execute(buf, incidentStruct)
	if err != nil {
		return "", "", err
	}
	textHtml = buf.String()

	return subject, textHtml, nil
}

func (n Email) sendEmailTo(subject, textHtml, to string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.opts.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetHeader("Auto-submitted", "auto-generated")
	m.SetBody("text/html", textHtml)
	return n.dialer.DialAndSend(m)
}
