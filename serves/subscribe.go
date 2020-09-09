package serves

import (
	"fmt"
	"net/http"
	"net/mail"
)

func (a Serve) SubscribeEmail(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if email == "" {
		JSONError(w, fmt.Errorf("You must set an email"), http.StatusPreconditionRequired)
		return
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		JSONError(w, err, http.StatusBadRequest)
		return
	}
	err = a.store.Subscribe(email)
	if err != nil {
		JSONError(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a Serve) UnsubscribeEmail(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if email == "" {
		HTMLError(w, fmt.Errorf("You must set an email"), http.StatusPreconditionRequired)
		return
	}
	err := a.store.Unsubscribe(email)
	if err != nil {
		HTMLError(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("you have successfully unsubscribed"))
}
