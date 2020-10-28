package extemplate

import (
	"encoding/json"
	"html/template"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/orange-cloudfoundry/statusetat/locations"
	"github.com/orange-cloudfoundry/statusetat/markdown"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/dustin/go-humanize"
)

func iconState(state models.ComponentState) string {
	switch state {
	case models.DegradedPerformance:
		return "error"
	case models.PartialOutage:
		return "remove_circle"
	case models.UnderMaintenance:
		return "watch_later"
	case models.MajorOutage:
		return "cancel"
	}
	return "check_circle"
}

func colorState(state models.ComponentState) string {
	switch state {
	case models.DegradedPerformance:
		return "purple"
	case models.PartialOutage:
		return "deep-orange"
	case models.UnderMaintenance:
		return "grey"
	case models.MajorOutage:
		return "red"
	}
	return "green"
}

func colorIncidentState(state models.IncidentState) string {
	switch state {
	case models.Unresolved:
		return "deep-orange"
	case models.Monitoring:
		return "blue"
	}
	return "green"
}

func colorHexState(state models.ComponentState) string {
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

func colorHexIncidentState(state models.IncidentState) string {
	switch state {
	case models.Unresolved:
		return "#ff5722"
	case models.Monitoring:
		return "#2196F3"
	}
	return "#4CAF50"
}

func timeFormat(t time.Time) string {

	return t.Format("Jan 02, 15:04 MST")
}

func timeFmtCustom(layout string, t time.Time) string {
	return t.Format(layout)
}

func timeStdFormat(t time.Time) string {
	return t.Format(time.RFC3339)
}

func stateFromIncidents(incidents []models.Incident) models.ComponentState {
	state := models.Operational

	for _, incident := range incidents {
		if incident.ComponentState > state {
			state = incident.ComponentState
		}
	}
	return state
}

func safeHTML(content string) template.HTML {
	return template.HTML(content)
}

func jsonify(content interface{}) template.JS {
	b, _ := json.Marshal(content)
	return template.JS(b)
}

func listMap(strs []string) template.JS {
	tags := make(map[string]interface{})
	for _, c := range strs {
		tags[c] = nil
	}
	return jsonify(tags)
}

func ref(d interface{}) interface{} {
	if reflect.TypeOf(d).Kind() != reflect.Ptr {
		return d
	}
	value := reflect.ValueOf(d)
	if !value.IsZero() {
		return value.Elem().Interface()
	}
	return reflect.New(value.Type().Elem()).Elem().Interface()
}

func tagify(strs []string) template.JS {
	tags := make([]map[string]interface{}, 0)
	for _, c := range strs {
		tags = append(tags, map[string]interface{}{
			"tag": c,
		})
	}
	return jsonify(tags)
}

func humanTime(t time.Time) string {
	return humanize.Time(t)
}

func timeNow() time.Time {
	return time.Now().In(locations.DefaultLocation())
}

func isAfterNow(t time.Time) bool {
	return time.Now().After(t)
}

func netUrl(baseUrl string) *url.URL {
	u, _ := url.Parse(baseUrl)
	return u
}

func markdownNoParaph(content string) template.HTML {
	b := markdown.Convert([]byte(content))
	content = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(string(b)), "<p>"), "</p>")
	return template.HTML(content)
}
