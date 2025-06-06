package main

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/alecthomas/kingpin/v2"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/statusetat/v2/config"
	"github.com/orange-cloudfoundry/statusetat/v2/locations"
	"github.com/orange-cloudfoundry/statusetat/v2/notifiers"
	_ "github.com/orange-cloudfoundry/statusetat/v2/notifiers/email"
	_ "github.com/orange-cloudfoundry/statusetat/v2/notifiers/grafana"
	_ "github.com/orange-cloudfoundry/statusetat/v2/notifiers/log"
	_ "github.com/orange-cloudfoundry/statusetat/v2/notifiers/plugin"
	_ "github.com/orange-cloudfoundry/statusetat/v2/notifiers/slack"
	"github.com/orange-cloudfoundry/statusetat/v2/serves"
	"github.com/orange-cloudfoundry/statusetat/v2/storages"
)

var (
	httpTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "statusetat_http_total_requests",
		Help: "Duration of HTTP requests.",
	}, []string{"code", "method", "path"})

	configFile = kingpin.Flag("config", "Path to Configuration File").Short('c').String()
)

//go:embed serves/website/assets
var assetsContent embed.FS

func main() {
	kingpin.Version(version.Print("statusetat"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var (
		err error
		c   config.Config
	)

	if configFile != nil && *configFile != "" {
		c, err = config.LoadConfigFromFile(*configFile)
	} else {
		c, err = config.LoadConfigFromEnv()
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	urls := make([]*url.URL, len(c.Targets))
	for i, target := range c.Targets {
		u, err := target.Validate()
		if err != nil {
			log.Fatalf("unable to parse target url '%s': %s", string(target), err.Error())
		}
		urls[i] = u
	}
	store, err := storages.Factory(urls)
	if err != nil {
		log.Fatal(err.Error())
	}
	assetsFs, err := fs.Sub(assetsContent, "serves/website/assets")
	if err != nil {
		log.Fatal(err.Error())
	}
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
	err = serves.Register(store, router, url.UserPassword(c.Username, c.Password), c)
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
		http.StripPrefix("/assets/", serves.NewMinifyMiddleware(http.FileServer(http.FS(assetsFs)))),
	)

	for _, n := range c.Notifiers {
		err := notifiers.AddNotifier(n.Type, n.Params, n.For, *c.BaseInfo)
		if err != nil {
			log.Fatalf("error when loading notifiers: %s", err.Error())
		}
	}

	go notifiers.Notify(store)

	protocol := "http://"
	if c.TlsConfig != nil {
		protocol = "https://"
	}

	log.Infof("Listening on address %s%s ...", protocol, c.Listen)

	if c.TlsConfig != nil {
		err = http.ListenAndServeTLS(c.Listen, c.TlsConfig.CertFile, c.TlsConfig.KeyFile, router)
	} else {
		err = http.ListenAndServe(c.Listen, router)
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
