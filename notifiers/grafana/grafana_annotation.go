package grafana

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
	"github.com/ArthurHlt/statusetat/notifiers"
	"github.com/mitchellh/mapstructure"
)

func init(){
	notifiers.RegisterNotifier(&GrafanaAnnotation{})
}

type ReqGrafanaAnnotation struct {
	DashboardID int      `json:"dashboardId,omitempty"`
	PanelID     int      `json:"panelId,omitempty"`
	Time        int64    `json:"time"`
	TimeEnd     int64    `json:"timeEnd,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Text        string   `json:"text"`
}

type OptsGrafanaAnnotation struct {
	ApiKey             string `mapstructure:"api_key"`
	Endpoint           string `mapstructure:"endpoint"`
	DashboardId        int    `mapstructure:"dashboard_id"`
	PanelId            int    `mapstructure:"panel_id"`
	Location           string `mapstructure:"location"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

type GrafanaAnnotation struct {
	httpClient *http.Client
	id         string
	opts       OptsGrafanaAnnotation
	loc        *time.Location
}

func (n GrafanaAnnotation) Creator(params map[string]interface{}, baseInfo config.BaseInfo) (notifiers.Notifier, error) {
	var opts OptsGrafanaAnnotation
	err := mapstructure.Decode(params, &opts)
	if err != nil {
		return nil, err
	}

	if opts.Location == "" {
		opts.Location = baseInfo.TimeZone
	}
	opts.Endpoint = strings.TrimSuffix(opts.Endpoint, "/")

	loc, err := time.LoadLocation(opts.Location)
	if err != nil {
		return nil, err
	}

	return &GrafanaAnnotation{
		httpClient: &http.Client{
			Transport: common.MakeHttpTransportWithHeader(opts.InsecureSkipVerify, "Authorization", "Bearer "+opts.ApiKey),
			Timeout:   5 * time.Second,
		},
		id:   opts.Endpoint,
		opts: opts,
		loc:  loc,
	}, nil
}

func (n GrafanaAnnotation) Name() string {
	return "grafana_annotation"
}

func (n GrafanaAnnotation) Id() string {
	return n.id
}

func (n GrafanaAnnotation) deleteNotify(incident models.Incident) error {
	req, err := http.NewRequest(http.MethodGet, n.opts.Endpoint+"/api/annotations", nil)
	if err != nil {
		return err
	}
	query := req.URL.Query()
	query.Add("tags", n.incidentTag(incident))
	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = query.Encode()
	respFind, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer respFind.Body.Close()
	b, err := ioutil.ReadAll(respFind.Body)
	if err != nil {
		return fmt.Errorf("Get error code %d", respFind.StatusCode)
	}
	if respFind.StatusCode > 399 {
		return fmt.Errorf("Get error code %d: %s", respFind.StatusCode, string(b))
	}

	type notify struct {
		ID int `json:"id"`
	}

	notifies := make([]notify, 0)
	err = json.Unmarshal(b, &notifies)
	if err != nil {
		return err
	}

	if len(notifies) == 0 {
		return nil
	}

	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/annotations/%d", n.opts.Endpoint, notifies[0].ID), nil)
	if err != nil {
		return err
	}
	respDelete, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer respDelete.Body.Close()
	err = common.ExtractHttpError(respDelete)
	if err != nil {
		return err
	}
	return nil
}

func (n GrafanaAnnotation) incidentTag(incident models.Incident) string {
	return "incident-guid-" + incident.GUID
}

func (n GrafanaAnnotation) Notify(incident models.Incident) error {
	if incident.IsScheduled {
		return nil
	}
	if len(incident.Messages) > 1 && incident.State != models.Resolved {
		return nil
	}
	err := n.deleteNotify(incident)
	if err != nil {
		return err
	}
	msg := incident.MainMessage()
	end := int64(0)
	if incident.State == models.Resolved {

		end = incident.UpdatedAt.In(n.loc).Unix() * 1000
	}
	b, _ := json.Marshal(ReqGrafanaAnnotation{
		DashboardID: n.opts.DashboardId,
		PanelID:     n.opts.PanelId,
		Time:        incident.CreatedAt.In(n.loc).Unix() * 1000,
		TimeEnd:     end,
		Tags:        []string{n.incidentTag(incident)},
		Text:        fmt.Sprintf("%s -- %s", msg.Title, msg.Content),
	})
	req, err := http.NewRequest(http.MethodPost, n.opts.Endpoint+"/api/annotations", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

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
