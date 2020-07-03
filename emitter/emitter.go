package emitter

import (
	"github.com/ArthurHlt/statusetat/models"
	"github.com/olebedev/emitter"
)

var e *emitter.Emitter = emitter.New(uint(100))

func Emit(incident models.Incident) {
	e.Emit("incident", incident)
}

func On() <-chan emitter.Event {

	return e.On("incident", emitter.Sync)
}
func Off(events ...<-chan emitter.Event) {
	e.Off("incident", events...)
}

func Listeners() []<-chan emitter.Event {
	return e.Listeners("incident")
}

func ToIncident(evt emitter.Event) models.Incident {
	return evt.Args[0].(models.Incident)
}
