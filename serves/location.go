package serves

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/orange-cloudfoundry/statusetat/locations"
)

const (
	LocationContextKey LocationContextType = iota
)

type LocationContextType int

type LocationHandler struct {
	store      *sessions.CookieStore
	defaultLoc *time.Location
}

func SetLocationContext(req *http.Request, location *time.Location) {
	parentContext := req.Context()
	ctxValueReq := req.WithContext(context.WithValue(parentContext, LocationContextKey, location))
	*req = *ctxValueReq
}

func (a Serve) Location(req *http.Request) *time.Location {
	val := req.Context().Value(LocationContextKey)
	if val == nil {
		return locations.DefaultLocation()
	}
	return val.(*time.Location)
}

func (a Serve) IsDefaultLocation(req *http.Request) bool {
	val := req.Context().Value(LocationContextKey)
	return val == nil
}

func NewLocationHandler(sessKey string) *LocationHandler {
	return &LocationHandler{
		store:      sessions.NewCookieStore([]byte(sessKey)),
		defaultLoc: locations.DefaultLocation(),
	}
}

func (s LocationHandler) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		session, _ := s.store.Get(req, "time-location")
		timezone := req.URL.Query().Get("timezone")
		if timezone == "" {
			timezoneTmp, ok := session.Values["timezone"]
			if ok {
				timezone = timezoneTmp.(string)
			}
		}
		if timezone == "" {
			next.ServeHTTP(w, req)
			return
		}
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = s.defaultLoc
		}
		SetLocationContext(req, loc)
		session.Values["timezone"] = timezone
		err = session.Save(req, w)
		if err != nil {
			HTMLError(w, err, http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, req)
	})
}
