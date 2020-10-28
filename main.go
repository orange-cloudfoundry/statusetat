package main

import (
	"flag"
	"net/http"
	"net/url"
	"strings"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/locations"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	_ "github.com/orange-cloudfoundry/statusetat/notifiers/email"
	_ "github.com/orange-cloudfoundry/statusetat/notifiers/grafana"
	_ "github.com/orange-cloudfoundry/statusetat/notifiers/plugin"
	_ "github.com/orange-cloudfoundry/statusetat/notifiers/slack"
	"github.com/orange-cloudfoundry/statusetat/serves"
	"github.com/orange-cloudfoundry/statusetat/storages"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var (
	httpTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statusetat_http_total_requests",
		Help: "Duration of HTTP requests.",
	}, []string{"code", "method", "path"})
)

func main() {
	configPtr := flag.String("config", "config.yml", "path to config file")
	c, err := config.LoadConfig(*configPtr)
	if err != nil {
		log.Fatal(err.Error())
	}

	urls := make([]*url.URL, len(c.Targets))
	for i, target := range c.Targets {
		urls[i] = target.URL
	}
	store, err := storages.Factory(urls)
	if err != nil {
		log.Fatal(err.Error())
	}
	box := packr.New("assets", "./website/assets")
	if c.BaseInfo.TimeZone != "" {
		err = locations.LoadByTimezone(c.BaseInfo.TimeZone)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	router := mux.NewRouter()

	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowedMethods:     []string{http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodPost, http.MethodPatch},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		Debug:              log.IsLevelEnabled(log.DebugLevel),
	}).Handler)
	router.Use(serves.NewLocationHandler(c.CookieKey).Handler)
	err = serves.Register(store, router, *c.BaseInfo,
		url.UserPassword(c.Username, c.Password), c.Components, c.Theme,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/assets") || strings.HasPrefix(req.URL.Path, "/metrics") {
				handler.ServeHTTP(w, req)
				return
			}
			promhttp.InstrumentHandlerCounter(httpTotalRequests.MustCurryWith(map[string]string{"path": req.URL.Path}), handler).ServeHTTP(w, req)
		})
	})
	router.Handle("/metrics", promhttp.Handler())
	router.PathPrefix("/assets").Handler(
		http.StripPrefix("/assets", serves.NewMinifyMiddleware(http.FileServer(box))),
	)

	for _, n := range c.Notifiers {
		err := notifiers.AddNotifier(n.Type, n.Params, n.For, *c.BaseInfo)
		if err != nil {
			log.Fatalf("error when loading notifiers: %s", err.Error())
		}
	}

	go notifiers.Notify(store)

	log.Infof("Listening on address %s ...", c.Listen)
	err = http.ListenAndServe(c.Listen, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}
