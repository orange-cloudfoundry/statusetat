package serves

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HtmlTemplater

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/markdown"
	"github.com/orange-cloudfoundry/statusetat/models"
	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/statusetat/storages"
)

var registeredFuncs = template.FuncMap{
	"iconState":             iconState,
	"colorState":            colorState,
	"colorIncidentState":    colorIncidentState,
	"textIncidentState":     models.TextIncidentState,
	"textScheduledState":    models.TextScheduledState,
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
	"isAfterNow":            isAfterNow,
	"markdownNoParaph":      markdownNoParaph,
	"tagify":                tagify,
	"ref":                   ref,
	"timeFmtCustom":         timeFmtCustom,
	"colorHexState":         colorHexState,
	"colorHexIncidentState": colorHexIncidentState,
	"join":                  strings.Join,
	"netUrl":                netUrl,
	"timeNow":               timeNow,
	"dict":                  dict,
	"metadataValue":         metadataValue,
	"timeAddDay":            timeAddDay,
	"stringReplace":         stringReplace,
	"sanitizeUrl":           sanitizeUrl,
}

type HtmlTemplater interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

type Serve struct {
	store          storages.Store
	xt             HtmlTemplater
	config         config.Config
	adminMenuItems []menuItem
}

//go:embed website/templates/*
var templateContent embed.FS

func getAllFilenames(efs embed.FS, pathFiles string) ([]string, error) {
	files, err := fs.ReadDir(efs, pathFiles)
	if err != nil {
		return nil, err
	}

	// only file name
	// 1131 0001-01-01 00:00:00 foo.gohtml -> foo.gohtml
	arr := make([]string, 0, len(files))
	for _, file := range files {
		arr = append(arr, file.Name())
	}

	return arr, nil
}

func Register(
	store storages.Store,
	router *mux.Router,
	userInfo *url.Userinfo,
	config config.Config,
) error {
	xt, err := template.New("").Funcs(registeredFuncs).ParseFS(templateContent,
		"website/templates/*.gohtml")
	// "website/templates/admin/*.gohtml",
	// "website/templates/components/*.gohtml")
	if err != nil {
		return err
	}
	fs.WalkDir(templateContent, "website/templates", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			log.Println(path)
			log.Println(d.Name())
			log.Println(d.Type())
		}
		return nil
	},
	)
	temple, _ := getAllFilenames(templateContent, "website/templates")
	log.Infoln("read templates", temple)
	return RegisterWithHtmlTemplater(store, router, userInfo, xt, config)
}

func RegisterWithHtmlTemplater(
	store storages.Store,
	router *mux.Router,
	userInfo *url.Userinfo,
	htmlTemplater HtmlTemplater,
	config config.Config,
) error {

	api := &Serve{
		store:  store,
		config: config,
		adminMenuItems: []menuItem{
			{
				ID:          "incident",
				DisplayName: "incident",
			},
			{
				ID:          "maintenance",
				DisplayName: "maintenance",
			},
			{
				ID:          "persistent_incident",
				DisplayName: config.Theme.PersistentDisplayName,
			},
			{
				ID:          "info",
				DisplayName: "info",
			},
		},
	}
	api.xt = htmlTemplater

	router.HandleFunc("/", api.Index)
	router.HandleFunc("/index", api.Index)
	router.HandleFunc("/history", api.History)
	router.HandleFunc("/incidents/{guid}", api.ShowIncident)
	router.HandleFunc("/rss.xml", api.Rss)
	router.HandleFunc("/atom.xml", api.Atom)
	router.HandleFunc("/cal.ics", api.Ical)
	router.HandleFunc("/healthy", api.HealthCheck)
	subRouter := router.PathPrefix("/v1").Subrouter()

	subRouter.HandleFunc("/subscribe", api.SubscribeEmail).Methods(http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodPut)
	subRouter.HandleFunc("/unsubscribe", api.UnsubscribeEmail).Methods(http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodPut)

	subRouter.HandleFunc("/components", api.ShowComponents).Methods(http.MethodGet)
	subRouter.HandleFunc("/flags/incident_states", api.ShowFlagIncidentStates).Methods(http.MethodGet)
	subRouter.HandleFunc("/flags/component_states", api.ShowFlagComponentStates).Methods(http.MethodGet)
	subRouter.HandleFunc("/markdown/preview", api.preview).Methods(http.MethodPost)
	subRouter.HandleFunc("/incidents/{guid}", api.Incident).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents", api.ByDate).Methods(http.MethodGet)
	subRouter.HandleFunc("/persistent_incidents", api.Persistents).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents/{incident_guid}/messages", api.ReadMessages).Methods(http.MethodGet)
	subRouter.HandleFunc("/incidents/{incident_guid}/messages/{message_guid}", api.ReadMessage).Methods(http.MethodGet)

	pass, _ := userInfo.Password()
	bauthHandler := httpauth.SimpleBasicAuth(userInfo.Username(), pass)
	subRouter.Handle("/subscribers", bauthHandler(http.HandlerFunc(api.ListSubscribers))).Methods(http.MethodGet)
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
	subrouterAdmin.HandleFunc("/persistent_incident", api.AdminPersistentIncidents)
	subrouterAdmin.HandleFunc("/maintenance", api.AdminMaintenance)
	subrouterAdmin.HandleFunc("/info", api.AdminInfo)
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
	//nolint:errcheck
	w.Write([]byte(fmt.Sprintf("%d %s: %s", code, http.StatusText(code), err.Error())))
}
