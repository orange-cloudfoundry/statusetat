package serves

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HtmlTemplater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gobuffalo/packr/v2"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/extemplate"
	"github.com/orange-cloudfoundry/statusetat/storages"
	log "github.com/sirupsen/logrus"
)

type HtmlTemplater interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

type Serve struct {
	store      storages.Store
	xt         HtmlTemplater
	baseInfo   config.BaseInfo
	components config.Components
	theme      config.Theme
}

func Register(
	store storages.Store,
	router *mux.Router,
	baseInfo config.BaseInfo,
	userInfo *url.Userinfo,
	components config.Components,
	theme config.Theme,
) error {
	xt := extemplate.New()
	box := packr.New("templates", "../website/templates")
	err := xt.ParseDir(box, []string{".gohtml"})
	if err != nil {
		return err
	}
	return RegisterWithHtmlTemplater(store, router, baseInfo, userInfo, components, theme, xt)
}

func RegisterWithHtmlTemplater(
	store storages.Store,
	router *mux.Router,
	baseInfo config.BaseInfo,
	userInfo *url.Userinfo,
	components config.Components,
	theme config.Theme,
	htmlTemplater HtmlTemplater,
) error {

	api := &Serve{
		store:      store,
		baseInfo:   baseInfo,
		components: components,
		theme:      theme,
	}
	api.xt = htmlTemplater

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
	subRouter.Handle("/incidents/{guid}/notify", bauthHandler(http.HandlerFunc(api.Notify))).Methods(http.MethodPut)
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

func (he HttpError) Error() string {
	return fmt.Sprintf("Http error (code: %d), %s: %s", he.Status, he.Detail, he.Description)
}

func JSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	LogError(err, code)
	err = json.NewEncoder(w).Encode(HttpError{
		Status:      code,
		Description: http.StatusText(code),
		Detail:      err.Error(),
	})
	if err != nil {
		panic(err)
	}
}

func LogError(err error, code int) {
	log.WithField("code", code).WithField("status_text", http.StatusText(code)).Debug(err.Error())
}

func HTMLError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	LogError(err, code)
	w.Write([]byte(fmt.Sprintf("%d %s: %s", code, http.StatusText(code), err.Error())))
}
