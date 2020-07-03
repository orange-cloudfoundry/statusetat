package main

import (
	"flag"
	"net/http"
	"net/url"
	"time"

	"github.com/ArthurHlt/statusetat/config"
	"github.com/ArthurHlt/statusetat/notifiers"
	_ "github.com/ArthurHlt/statusetat/notifiers"
	"github.com/ArthurHlt/statusetat/serves"
	"github.com/ArthurHlt/statusetat/storages"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func main() {
	configPtr := flag.String("config", "config.yml", "path to c file")
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
	loc := time.Local
	if c.BaseInfo.TimeZone != "" {
		loc, err = time.LoadLocation(c.BaseInfo.TimeZone)
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
		Debug:              true,
	}).Handler)
	router.Use(serves.NewLocationHandler(c.CookieKey, loc).Handler)
	err = serves.Register(store, router, *c.BaseInfo,
		url.UserPassword(c.Username, c.Password), c.Components,
		loc,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	router.PathPrefix("/assets").Handler(
		http.StripPrefix("/assets", serves.NewMinifyMiddleware(http.FileServer(box))),
	)

	for _, n := range c.Notifiers {
		err := notifiers.AddNotifier(n.Type, n.Params, n.For, *c.BaseInfo)
		if err != nil {
			log.Fatalf("error when loading notifiers: %s", err.Error())
		}
	}

	go notifiers.Notify()

	log.Infof("Listening on address %s ...", c.Listen)
	err = http.ListenAndServe(c.Listen, router)
	if err != nil {
		log.Fatal(err.Error())
	}
}
