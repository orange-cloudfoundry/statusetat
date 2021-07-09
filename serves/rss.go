package serves

import (
	"net/http"
	"time"

	"github.com/gorilla/feeds"
)

func (a Serve) Rss(w http.ResponseWriter, req *http.Request) {
	feed, err := a.feed(req)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	rss, err := feed.ToRss()
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rss))
}

func (a Serve) Atom(w http.ResponseWriter, req *http.Request) {
	feed, err := a.feed(req)
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	atom, err := feed.ToAtom()
	if err != nil {
		HTMLError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/atom+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(atom))
}

func (a Serve) feed(req *http.Request) (*feeds.Feed, error) {
	loc := a.Location(req)
	incidents, err := a.store.ByDate(time.Now().Add(-7*24*time.Hour).In(loc), time.Now().In(loc))
	if err != nil {
		return nil, err
	}
	for i, incident := range incidents {
		incident.Messages = a.convertMessageToHtml(incident.Messages)
		incidents[i] = incident
	}
	feed := &feeds.Feed{
		Title:       a.baseInfo.Title,
		Link:        &feeds.Link{Href: a.baseInfo.BaseURL},
		Description: "Get the status",
		Created:     time.Now().In(loc),
	}

	feed.Items = make([]*feeds.Item, len(incidents))
	for i, incident := range incidents {
		mainMsg := incident.MainMessage()
		content := ""

		if incident.Components != nil {
			content += "<b>Impacted components</b>:<br/><ul>"
			for _, comp := range *incident.Components {
				content += "<li>" + comp.String() + "</li>"
			}
			content += "</ul>"
		}
		for _, msg := range incident.UpdateMessages() {
			content += "<p>" + msg.Content + "</p>"
		}
		feed.Items[i] = &feeds.Item{
			Title:       mainMsg.Title,
			Link:        &feeds.Link{Href: a.baseInfo.BaseURL + "/" + incident.GUID},
			Description: mainMsg.Content,
			Created:     incident.CreatedAt,
			Content:     content,
		}
	}
	return feed, nil
}
