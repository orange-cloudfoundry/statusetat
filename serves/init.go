package serves

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ArthurHlt/statusetat/common"
	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/extemplate"
	"github.com/ArthurHlt/statusetat/markdown"
	"github.com/ArthurHlt/statusetat/models"
	"github.com/ArthurHlt/statusetat/storages"
	"github.com/gobuffalo/packr/v2"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Serve struct {
	store      storages.Store
	xt         *extemplate.Extemplate
	baseInfo   config.BaseInfo
	components config.Components
	loc        *time.Location
	funcs      template.FuncMap
	theme      config.Theme
}

func Register(
	store storages.Store,
	router *mux.Router,
	baseInfo config.BaseInfo,
	userInfo *url.Userinfo,
	components config.Components,
	loc *time.Location,
	theme config.Theme,
) error {

	api := &Serve{
		store:      store,
		baseInfo:   baseInfo,
		components: components,
		loc:        loc,
		theme:      theme,
	}

	funcs := template.FuncMap{
		"iconState":             iconState,
		"colorState":            colorState,
		"colorIncidentState":    colorIncidentState,
		"textIncidentState":     models.TextIncidentState,
		"textState":             models.TextState,
		"timeFormat":            timeFormat,
		"timeStdFormat":         timeStdFormat,
		"title":                 common.Title,
		"markdown":              markdown.ConvertSafeTemplate,
		"stateFromIncidents":    stateFromIncidents,
		"safeHTML":              safeHTML,
		"humanTime":             humanTime,
		"jsonify":               jsonify,
		"listMap":               listMap,
		"humanDuration":         common.HumanDuration,
		"timeNow":               api.timeNow,
		"isAfterNow":            isAfterNow,
		"baseUrl":               api.baseUrl,
		"markdownNoParaph":      markdownNoParaph,
		"tagify":                tagify,
		"ref":                   ref,
		"timeFmtCustom":         timeFmtCustom,
		"colorHexState":         colorHexState,
		"colorHexIncidentState": colorHexIncidentState,
		"join":                  strings.Join,
	}
	extemplate.SetFuncs(funcs)

	api.xt = extemplate.New()
	box := packr.New("templates", "../website/templates")
	err := api.xt.ParseDir(box, []string{".html"})
	if err != nil {
		return err
	}

	router.HandleFunc("/", api.Index)
	router.HandleFunc("/index", api.Index)
	router.HandleFunc("/history", api.History)
	router.HandleFunc("/incidents/{guid}", api.ShowIncident)
	router.HandleFunc("/rss.xml", api.Rss)
	router.HandleFunc("/atom.xml", api.Atom)
	router.HandleFunc("/cal.ics", api.Ical)
	subRouter := router.PathPrefix("/v1").Subrouter()

	subRouter.HandleFunc("/subscribe", api.SubscribeEmail).Methods(http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodPut)
	subRouter.HandleFunc("/unsubscribe", api.UnsubscribeEmail).Methods(http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodPut)

	subRouter.HandleFunc("/components", api.ShowComponents).Methods(http.MethodGet)
	subRouter.HandleFunc("/flags/incident_states", api.ShowFlagIncidentStates).Methods(http.MethodGet)
	subRouter.HandleFunc("/flags/component_states", api.ShowFlagComponentStates).Methods(http.MethodGet)
	subRouter.HandleFunc("/markdown/preview", api.preview).Methods(http.MethodPost)
	subRouter.HandleFunc("/incidents/{guid}", api.Incident).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents", api.ByDate).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents/{incident_guid}/messages", api.ReadMessages).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents/{incident_guid}/messages/{message_guid}", api.ReadMessage).Methods(http.MethodGet)

	pass, _ := userInfo.Password()
	bauthHandler := httpauth.SimpleBasicAuth(userInfo.Username(), pass)
	subRouter.Handle("/incidents", bauthHandler(http.HandlerFunc(api.CreateIncident))).Methods(http.MethodPost)
	subRouter.Handle("/incidents/{guid}", bauthHandler(http.HandlerFunc(api.Update))).Methods(http.MethodPut)
	subRouter.Handle("/incidents/{guid}", bauthHandler(http.HandlerFunc(api.Delete))).Methods(http.MethodDelete)
	subRouter.Handle("/incidents/{incident_guid}/messages", bauthHandler(http.HandlerFunc(api.AddMessage))).Methods(http.MethodPost)
	subRouter.Handle("/incidents/{incident_guid}/messages/{message_guid}", bauthHandler(http.HandlerFunc(api.UpdateMessage))).Methods(http.MethodPut)
	subRouter.Handle("/incidents/{incident_guid}/messages/{message_guid}", bauthHandler(http.HandlerFunc(api.DeleteMessage))).Methods(http.MethodDelete)

	subrouterAdmin := router.PathPrefix("/admin").Subrouter()
	subrouterAdmin.Use(bauthHandler)
	subrouterAdmin.HandleFunc("/dashboard", api.AdminIncidents)
	subrouterAdmin.HandleFunc("/incident", api.AdminIncidents)
	subrouterAdmin.HandleFunc("/maintenance", api.AdminMaintenance)
	subrouterAdmin.HandleFunc("/incident/add", api.AdminAddEditIncident)
	subrouterAdmin.HandleFunc("/incident/edit/{guid}", api.AdminAddEditIncident)
	subrouterAdmin.HandleFunc("/maintenance/add", api.AdminAddEditMaintenance)
	subrouterAdmin.HandleFunc("/maintenance/edit/{guid}", api.AdminAddEditMaintenance)

	return nil
}

type HttpError struct {
	Description string `json:"description"`
	Detail      string `json:"detail"`
	Status      int    `json:"status"`
}

func JSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	LogError(err, code)
	json.NewEncoder(w).Encode(HttpError{
		Status:      code,
		Description: http.StatusText(code),
		Detail:      err.Error(),
	})
}

func LogError(err error, code int) {
	log.WithField("code", code).WithField("status_text", http.StatusText(code)).Debug(err.Error())
}

func HTMLError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	LogError(err, code)
	w.Write([]byte(fmt.Sprintf("%d %s: %s", code, http.StatusText(code), err.Error())))
}
