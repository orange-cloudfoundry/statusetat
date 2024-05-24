package serves_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/emitter"
	"github.com/orange-cloudfoundry/statusetat/emitter/emitterfakes"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/serves"
	"github.com/orange-cloudfoundry/statusetat/serves/servesfakes"
	"github.com/orange-cloudfoundry/statusetat/storages"
	"github.com/orange-cloudfoundry/statusetat/storages/storagesfakes"
)

var router *mux.Router

var BaseInfo = config.BaseInfo{
	BaseURL:  "http://localhost",
	Support:  "http://localhost",
	Contact:  "http://localhost",
	Title:    "MyStatus",
	TimeZone: "UTC",
}

var UserInfo = url.UserPassword("admin", "admin")

var Component1 = config.Component{
	Name:        "component1",
	Description: "",
	Group:       "",
}

var Component2 = config.Component{
	Name:        "component1",
	Description: "",
	Group:       "",
}

var Components = config.Components{Component1, Component2}

var Theme = config.Theme{
	PreStatus:       "",
	PostStatus:      "",
	PreTimeline:     "",
	PostTimeline:    "",
	PreMaintenance:  "",
	PostMaintenance: "",
	Footer:          "",
}

var fakeStoreMem *storagesfakes.FakeStore
var fakeHtmlTemplater *servesfakes.FakeHtmlTemplater
var fakeEmitter *emitterfakes.FakeEmitterInterface

var _ = BeforeSuite(func() {

})

var _ = BeforeEach(func() {
	router = mux.NewRouter()
	fakeEmitter = &emitterfakes.FakeEmitterInterface{}
	fakeHtmlTemplater = &servesfakes.FakeHtmlTemplater{}
	fakeStoreMem = &storagesfakes.FakeStore{}
	u, _ := url.Parse("sqlite://:memory:")
	crea := (&storages.DB{}).Creator()
	var err error
	dbStore, err := crea(u)
	Expect(err).ToNot(HaveOccurred())

	fakeStoreMem.DetectStub = dbStore.Detect
	fakeStoreMem.CreatorStub = dbStore.Creator

	fakeStoreMem.CreateStub = dbStore.Create
	fakeStoreMem.UpdateStub = dbStore.Update
	fakeStoreMem.DeleteStub = dbStore.Delete
	fakeStoreMem.ReadStub = dbStore.Read
	fakeStoreMem.ByDateStub = dbStore.ByDate

	fakeStoreMem.SubscribeStub = dbStore.Subscribe
	fakeStoreMem.UnsubscribeStub = dbStore.Unsubscribe
	fakeStoreMem.SubscribersStub = dbStore.Subscribers

	err = serves.RegisterWithHtmlTemplater(fakeStoreMem, router, UserInfo, fakeHtmlTemplater, config.Config{
		Targets:    config.Targets{},
		Listen:     "",
		Log:        &config.Log{},
		Components: Components,
		BaseInfo:   &BaseInfo,
		Username:   "",
		Password:   "",
		CookieKey:  "",
		Notifiers:  []config.Notifier{},
		Theme:      &Theme,
	})
	Expect(err).ToNot(HaveOccurred())
	emitter.SetEmitter(fakeEmitter)
})

var _ = AfterSuite(func() {
})

func TestServes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serves Suite")
}

func TemplateUnmarshalIn(expectName string, v interface{}) func(wr io.Writer, name string, data interface{}) error {
	return func(wr io.Writer, name string, data interface{}) error {
		Expect(name).To(Equal(expectName))
		b, err := json.Marshal(data)
		Expect(err).ToNot(HaveOccurred())
		err = json.Unmarshal(b, v)
		Expect(err).ToNot(HaveOccurred())
		return nil
	}
}

func AssertTemplateStruct(templateNameExpect string, vExpected interface{}) func(name string, data map[string]interface{}) {
	return func(name string, data map[string]interface{}) {

		vExpectedMap, ok := vExpected.(map[string]interface{})
		if !ok {
			b, err := json.Marshal(vExpected)
			Expect(err).ToNot(HaveOccurred())
			vExpectedMap = make(map[string]interface{})
			err = json.Unmarshal(b, &vExpectedMap)
			Expect(err).ToNot(HaveOccurred())
		}
		Expect(data).To(BeEquivalentTo(vExpectedMap))
	}
}

func AssertTemplateFunc(assert func(name string, data map[string]interface{})) func(wr io.Writer, name string, data interface{}) error {
	return func(wr io.Writer, name string, data interface{}) error {
		b, err := json.Marshal(data)
		Expect(err).ToNot(HaveOccurred())
		dataMap := make(map[string]interface{})
		err = json.Unmarshal(b, &dataMap)
		Expect(err).ToNot(HaveOccurred())
		assert(name, dataMap)
		return nil
	}
}

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (rr TestResponseRecorder) UnmarshalToIncidents() (models.Incidents, error) {
	incs := make(models.Incidents, 0)
	err := rr.Unmarshal(&incs)
	if err != nil {
		return incs, err
	}
	return incs, nil
}

func (rr TestResponseRecorder) UnmarshalToIncident() (models.Incident, error) {
	var inc models.Incident
	err := rr.Unmarshal(&inc)
	if err != nil {
		return models.Incident{}, err
	}
	return inc, nil
}

func (rr TestResponseRecorder) UnmarshalToMessages() (models.Messages, error) {
	messages := make(models.Messages, 0)
	err := rr.Unmarshal(&messages)
	if err != nil {
		return messages, err
	}
	return messages, nil
}

func (rr TestResponseRecorder) UnmarshalToMessage() (models.Message, error) {
	var mess models.Message
	err := rr.Unmarshal(&mess)
	if err != nil {
		return models.Message{}, err
	}
	return mess, nil
}

func (rr TestResponseRecorder) UnmarshalToError() (serves.HttpError, error) {
	var httpErr serves.HttpError
	err := rr.Unmarshal(&httpErr)
	if err != nil {
		return serves.HttpError{}, err
	}
	return httpErr, nil
}

func (rr TestResponseRecorder) Unmarshal(v interface{}) error {
	return json.Unmarshal(rr.Body.Bytes(), v)
}

func (rr TestResponseRecorder) IsInError() bool {
	return rr.Code >= 400
}

func (rr TestResponseRecorder) CheckError() error {
	if !rr.IsInError() {
		return nil
	}
	if strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		he, err := rr.UnmarshalToError()
		if err != nil {
			return err
		}
		return he
	}

	return fmt.Errorf("Error with code %d: %s", rr.Code, rr.Body.String())
}

func CallRequest(req *http.Request) *TestResponseRecorder {
	rr := &TestResponseRecorder{httptest.NewRecorder()}
	router.ServeHTTP(rr, req)
	return rr
}

func NewRequestInt(method, target string, v interface{}) *http.Request {
	if v == nil {
		return httptest.NewRequest(method, target, nil)
	}
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(b))
	serves.SetLocationContext(req, time.UTC)
	return req
}

func NewRequestIntAdmin(method, target string, v interface{}) *http.Request {
	req := NewRequestInt(method, target, v)
	req.SetBasicAuth("admin", "admin")
	return req
}
