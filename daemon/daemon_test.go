package daemon_test

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/helto4real/go-daemon/daemon"
	"github.com/sirupsen/logrus"

	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
)

func TestGetAllApplicationConfigFilePaths(t *testing.T) {
	h.Equals(t, true, true)
	x := client.HassEntity{}
	h.NotEquals(t, nil, x)
}

func TestNewApplicationDaemon(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	h.NotEquals(t, nil, d)
}

func TestStartAndStop(t *testing.T) {
	d := daemon.NewApplicationDaemon()

	fake := newFakeHomeAssistant()
	defer func() {
		d.Stop()
		h.Equals(t, 1, fake.nrOfCallsStop)
	}()

	d.Start("testdata/ok", fake, newAvailableApps())
	h.Equals(t, 1, fake.nrOfCallsStart)
}

func TestStartFailConfigNotExist(t *testing.T) {
	d := daemon.NewApplicationDaemon()

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/okk", fake, newAvailableApps())

	h.Equals(t, 0, fake.nrOfCallsStart)
}

func TestBasicDeamonHelperFunctions(t *testing.T) {
	d := daemon.NewApplicationDaemon()

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	helper := d.(daemon.DaemonAppHelper)
	h.NotEquals(t, nil, helper.GetCancelContext())
	h.NotEquals(t, nil, helper.GetCancelFunction())
	helper.TurnOn("any")
	h.Equals(t, 1, fake.nrOfCallsCallService)
	helper.TurnOff("any")
	h.Equals(t, 2, fake.nrOfCallsCallService)
	helper.Toggle("any")
	h.Equals(t, 3, fake.nrOfCallsCallService)
}

func TestBasicGetEntity(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)

	defer d.Stop()
	d.Start("testdata/ok", newFakeHomeAssistant(), newAvailableApps())

	e, ok := hlpr.GetEntity("entity1")

	h.Equals(t, true, ok)
	h.NotEquals(t, nil, e)

	h.Equals(t, "entityname", e.Name)
}

func TestAtSunset(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	ch := make(chan bool)
	tmr := hlpr.AtSunset(time.Duration(0), ch)
	h.NotEquals(t, nil, tmr)
	res := <-ch
	h.Equals(t, true, res)

}

func TestAtSunsetErrors(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	ch := make(chan bool)

	fake.fakeNoSunEntity = true
	tmr := hlpr.AtSunset(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	fake.fakeNoSunEntity = false
	fake.fakeNoAttribute = true
	tmr = hlpr.AtSunset(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	fake.fakeNoAttribute = false
	fake.fakeMalformatedDates = true
	tmr = hlpr.AtSunset(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	level := logrus.GetLevel()
	logrus.SetLevel(logrus.DebugLevel)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(level)
	}()
	fake.fakeMalformatedDates = false
	fake.fakeTimeBeforeNow = true
	tmr = hlpr.AtSunset(time.Duration(0), ch)
	h.NotEquals(t, (*time.Timer)(nil), tmr)
	s := mockStdErr.String()
	log.Print(s)
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "We are before in time, adding 24 hours"))
}

func TestAtSunRise(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	ch := make(chan bool)
	tmr := hlpr.AtSunrise(time.Duration(0), ch)
	h.NotEquals(t, nil, tmr)
	res := <-ch
	h.Equals(t, true, res)

}
func TestAtSunriseErrors(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	ch := make(chan bool)

	fake.fakeNoSunEntity = true
	tmr := hlpr.AtSunrise(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	fake.fakeNoSunEntity = false
	fake.fakeNoAttribute = true
	tmr = hlpr.AtSunrise(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	fake.fakeNoAttribute = false
	fake.fakeMalformatedDates = true
	tmr = hlpr.AtSunrise(time.Duration(0), ch)
	h.Equals(t, (*time.Timer)(nil), tmr)

	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	level := logrus.GetLevel()
	logrus.SetLevel(logrus.DebugLevel)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(level)
	}()
	fake.fakeMalformatedDates = false
	fake.fakeTimeBeforeNow = true
	tmr = hlpr.AtSunrise(time.Duration(0), ch)

	h.NotEquals(t, (*time.Timer)(nil), tmr)
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "We are before in time, adding 24 hours"))

}

func TestListenState(t *testing.T) {
	d := daemon.NewApplicationDaemon()
	hlpr := d.(daemon.DaemonAppHelper)
	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/ok", fake, newAvailableApps())

	// Set to 2 deep so it wont block
	hchan1 := make(chan client.HassEntity, 2)
	hchan2 := make(chan client.HassEntity, 2)
	hlpr.ListenState("entity1", hchan1)
	hlpr.ListenState("entity1", hchan2)

	go func() {
		// Fake coming a new message from hass
		fake.entityChannel <- &client.HassEntity{
			ID:   "entity1",
			Name: "entityname",
			Old: client.HassEntityState{
				State: "anystate"}}

		// Wait for it to return to us

	}()
	ev1 := <-hchan1
	ev2 := <-hchan2
	h.Equals(t, "entityname", ev1.Name)
	h.Equals(t, "entityname", ev2.Name)

	// Last check if we double register channel
	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
	}()
	hlpr.ListenState("entity1", hchan1)
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "Listen state already registered on "))

}

// Check the testdata/badformat/go-daemon.yaml
func TestStartFailMalformatedConfig(t *testing.T) {
	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
	}()

	d := daemon.NewApplicationDaemon()

	fake := newFakeHomeAssistant()

	defer d.Stop()
	d.Start("testdata/badformat", fake, newAvailableApps())

	h.Equals(t, 0, fake.nrOfCallsStart)
	h.Equals(t, true, strings.Contains(mockStdErr.String(),
		"Failed to open config file, ending -> yaml: line 6: could not find expected ':'"))
}

/*
Fake testapp
*/
type testapp struct {
}

func (a testapp) Initialize(helper daemon.DaemonAppHelper, config daemon.DeamonAppConfig) bool {
	return true
}

func (a testapp) Cancel() {

}

func newAvailableApps() map[string]interface{} {
	r := make(map[string]interface{})

	r["testapp"] = testapp{}
	r["testapp2"] = testapp{}
	return r
}

/*
Fake home assistant client
*/
type fakeHomeAssistant struct {
	nrOfCallsStart       int
	nrOfCallsStop        int
	nrOfCallsCallService int
	fakeNoSunEntity      bool
	fakeNoAttribute      bool
	fakeMalformatedDates bool
	fakeTimeBeforeNow    bool

	entityChannel chan *client.HassEntity
	statusChannel chan bool
}

func newFakeHomeAssistant() *fakeHomeAssistant {
	f := fakeHomeAssistant{
		entityChannel: make(chan *client.HassEntity),
		statusChannel: make(chan bool)}

	return &f
}

// Start daemon only use in main
func (a *fakeHomeAssistant) Start(host string, ssl bool, token string) bool {
	a.nrOfCallsStart = a.nrOfCallsStart + 1
	return true
}

// Stop daemon only use in main
func (a *fakeHomeAssistant) Stop() {
	a.nrOfCallsStop = a.nrOfCallsStop + 1
}
func (a *fakeHomeAssistant) GetEntity(entity string) (*client.HassEntity, bool) {
	dur := time.Duration(time.Second * 1)
	if a.fakeTimeBeforeNow {
		// One second back in time
		dur = time.Duration(time.Minute * -1)
	}
	if entity == "sun.sun" && !a.fakeNoSunEntity && !a.fakeNoAttribute && !a.fakeMalformatedDates {
		return &client.HassEntity{
			ID:   "sun.sun",
			Name: "sun.sun",
			New: client.HassEntityState{
				State: "below_horizon",
				Attributes: map[string]string{
					"next_setting": time.Now().Add(dur).Format(time.RFC3339),
					"next_rising":  time.Now().Add(dur).Format(time.RFC3339)}}}, true
	}

	if entity == "sun.sun" && a.fakeMalformatedDates {
		return &client.HassEntity{
			ID:   "sun.sun",
			Name: "sun.sun",
			New: client.HassEntityState{
				State: "below_horizon",
				Attributes: map[string]string{
					"next_setting": "not a date",
					"next_rising":  "not a date"}}}, true
	}
	if entity == "sun.sun" && a.fakeNoAttribute {
		return &client.HassEntity{
			ID:   "sun.sun",
			Name: "sun.sun",
			New: client.HassEntityState{
				State:      "below_horizon",
				Attributes: map[string]string{}}}, true
	}

	if entity == "entity1" {
		return &client.HassEntity{
			ID:   "entity1",
			Name: "entityname"}, true
	}

	return nil, false
}
func (a *fakeHomeAssistant) CallService(service string, serviceData map[string]string) {
	a.nrOfCallsCallService = a.nrOfCallsCallService + 1
}
func (a *fakeHomeAssistant) GetEntityChannel() chan *client.HassEntity {
	return a.entityChannel
}
func (a *fakeHomeAssistant) GetStatusChannel() chan bool {
	return a.statusChannel
}
