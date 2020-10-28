package emitter

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . emitterInterface

import (
	"github.com/olebedev/emitter"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type emitterInterface interface {
	Emit(topic string, args ...interface{}) chan struct{}
	On(topic string, middlewares ...func(*emitter.Event)) <-chan emitter.Event
	Off(topic string, channels ...<-chan emitter.Event)
	Listeners(topic string) []<-chan emitter.Event
}

var e emitterInterface = emitter.New(uint(100))

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

// this is only made for testing purpose
func SetEmitter(emit emitterInterface) {
	e = emit
}
