package defaultapps

import (
	"context"
	"io/ioutil"
	"path"
	"sync"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-daemon/daemon/config"
	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
)

func TestInitialize(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase1.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 6)
	h.Equals(t, fake.getEntity, 6)
	h.Equals(t, fake.setEntity, 2)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Home")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Home")
}

func TestInitializeAway(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase2.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 3)
	h.Equals(t, fake.getEntity, 3)
	h.Equals(t, fake.setEntity, 1)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Away")
}

func TestInitializeOnOffTrueFalse(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase3.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 5)
	h.Equals(t, fake.getEntity, 5)
	h.Equals(t, fake.setEntity, 5)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Home")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Away")
	h.Equals(t, fake.fakePeopleConfig["person3"].State, "Home")
	h.Equals(t, fake.fakePeopleConfig["person4"].State, "Away")
	h.Equals(t, fake.fakePeopleConfig["person5"].State, "somestate")
}

func TestInitializeJustArrivedJustLeft(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase4.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 2)
	h.Equals(t, fake.getEntity, 2)
	h.Equals(t, fake.setEntity, 2)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just arrived")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Just left")
}

func TestInitializeJustArrivedJustLeftSameState(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase5.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 2)
	h.Equals(t, fake.getEntity, 2)
	h.Equals(t, fake.setEntity, 2)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just arrived")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Just left")
}

func TestInitializeInZone(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase6.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 2)
	h.Equals(t, fake.getEntity, 2)
	h.Equals(t, fake.setEntity, 2)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just arrived")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Away")
}

func TestInitializeNamedSourceType(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase7.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 3)
	h.Equals(t, fake.getEntity, 3)
	h.Equals(t, fake.setEntity, 1)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Home")
}

func TestHandleUpdatedDevice(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase7.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	fake.fakeDevices["device_tracker.bt"].New.State = "not_home"
	fake.fakeDevices["device_tracker.wifi"].New.State = "not_home"

	app.handleUpdatedDevice("device_tracker.bt", false)
	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just left")
}

func TestHandleHomeWhenHome(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase8.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Home")
}

func TestManageTime(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase10.yml")
	fake.fakeDevices["device_tracker.gps"].New.LastUpdated = time.Now().UTC()
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just arrived")
}

func TestTackUpdatedDevice(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase9.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	app.trackerChannel <- client.HassEntity{
		ID:   "device_tracker.gps",
		Name: "device_tracker.gps",
		New: client.HassEntityState{
			State: "home",
			Attributes: map[string]interface{}{
				"source_type": "gps",
			},
		},
	}
	time.Sleep(time.Millisecond * 100)
	fake.confMutex.Lock()
	defer fake.confMutex.Unlock()
	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Just left")
}

func TestInitializeJustArrivedJustLeftGoBackToState(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase4.yml")
	// Set the timer to a zero value
	fake.timeChangedState = 0

	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	time.Sleep(time.Millisecond * 100)
	fake.confMutex.Lock()
	defer fake.confMutex.Unlock()
	h.Equals(t, fake.listenState, 2)
	h.Equals(t, fake.getEntity, 4)
	h.Equals(t, fake.setEntity, 4)

	h.Equals(t, fake.fakePeopleConfig["person1"].State, "Home")
	h.Equals(t, fake.fakePeopleConfig["person2"].State, "Away")
}

func TestInitializeAndCancel(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase4.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	time.Sleep(time.Millisecond * 100)
	app.Cancel()
	<-app.cancelContext.Done()
}

func TestGetPersonOwningDevice(t *testing.T) {
	app := PeopleApp{}
	fake := newFakeDaemonHelperTestCase("testcase2.yml")
	app.Initialize(fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})

	person := app.getPersonOwningDevice("device_tracker.bt")

	h.Equals(t, person, "person1")

	person = app.getPersonOwningDevice("device_tracker.bt.noexist")

	h.Equals(t, person, "")
}

func TestPeopleConfigured(t *testing.T) {
	app := PeopleApp{}
	h.Equals(t, app.peopleConfigured(), false)
}

func TestNewState(t *testing.T) {
	stateData := newState("a state")
	h.Equals(t, stateData.state, "a state")
}
func TestDistance(t *testing.T) {

	// Same coordinate should be zero on both
	h.Equals(t, distance(1.0, 1.0, 1.0, 1.0, "K"), 0.0)
	h.Equals(t, distance(1.0, 1.0, 1.0, 1.0, "N"), 0.0)
	h.Equals(t, distance(1.0, 1.0, 2.0, 2.0, "K"), 157.2178677858707)
	h.Equals(t, distance(1.0, 1.0, 2.0, 2.0, "N"), 84.83456388767729)

	h.Equals(t, distance(1.0, 1.0, 5.0, 5.0, "K"), 628.4879299059344)

}

type fakeDaemonAppHelper struct {
	listenState      int
	getEntity        int
	setEntity        int
	timeChangedState int
	fakePeopleConfig map[string]*config.PeopleConfig
	fakeDevices      map[string]*client.HassEntity
	confMutex        *sync.Mutex
}

func newFakeDaemonHelper() *fakeDaemonAppHelper {
	returnVal := &fakeDaemonAppHelper{
		fakePeopleConfig: map[string]*config.PeopleConfig{},
		fakeDevices:      map[string]*client.HassEntity{},
		confMutex:        &sync.Mutex{},
		timeChangedState: 300,
	}

	return returnVal
}

func newFakeDaemonHelperTestCase(filename string) *fakeDaemonAppHelper {
	returnVal := &fakeDaemonAppHelper{
		fakePeopleConfig: map[string]*config.PeopleConfig{},
		fakeDevices:      map[string]*client.HassEntity{},
		confMutex:        &sync.Mutex{},
		timeChangedState: 300,
	}

	returnVal.loadTestCase(filename)

	return returnVal
}

func (a *fakeDaemonAppHelper) GetCancelContext() context.Context {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) GetCancelFunction() context.CancelFunc {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) GetEntity(entity string) (*client.HassEntity, bool) {
	a.confMutex.Lock()
	defer a.confMutex.Unlock()

	a.getEntity = a.getEntity + 1

	if a.fakeDevices != nil && len(a.fakeDevices) > 0 {
		enityFromTestCase, ok := a.fakeDevices[entity]
		if !ok {
			panic("failed to get device " + entity)
		}
		return enityFromTestCase, true
	}

	return nil, false
}

func (a *fakeDaemonAppHelper) SetEntity(entity *client.HassEntity) bool {
	a.confMutex.Lock()
	defer a.confMutex.Unlock()
	a.setEntity = a.setEntity + 1
	return true
}

func (a *fakeDaemonAppHelper) TurnOn(entity string) {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) TurnOff(entity string) {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) Toggle(entity string) {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) ListenCallServiceEvent(domain string, service string, callServiceChannel chan client.HassCallServiceEvent) {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) ListenState(entity string, stateChannel chan client.HassEntity) {
	a.listenState = a.listenState + 1
}

func (a *fakeDaemonAppHelper) AtSunset(offset time.Duration, sunsetChannel chan bool) *time.Timer {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) AtSunrise(offset time.Duration, sunriseChannel chan bool) *time.Timer {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) NewEntity(id string, daemonHelper d.DaemonAppHelper, autoRespondServiceCall bool, changedEntityChannel chan d.DaemonEntity) d.DaemonEntity {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) GetPeople() map[string]*config.PeopleConfig {
	if a.fakePeopleConfig != nil {
		// Open ups for other fakes to test different things
		return a.fakePeopleConfig
	}

	return nil
}

func (a *fakeDaemonAppHelper) GetSettings() *config.SettingsConfig {
	return &config.SettingsConfig{
		TrackingSettings: &config.TrackingStateSettingsConfig{
			JustArrivedTime:  a.timeChangedState,
			JustLeftTime:     a.timeChangedState,
			HomeState:        "Home",
			JustArrivedState: "Just arrived",
			JustLeftState:    "Just left",
			AwayState:        "Away",
		},
	}
}

func (a *fakeDaemonAppHelper) GetLocation() d.Location {
	return d.Location{
		Longitude: 1.0,
		Latitude:  1.0,
		Elevation: 0,
	}
}

func (a *fakeDaemonAppHelper) loadTestCase(filename string) {
	caseData := testCaseConfig{}
	data, error := ioutil.ReadFile(path.Join("testdata/people", filename))
	if error != nil {
		panic(error)
	}
	error = yaml.Unmarshal(data, &caseData)

	if error != nil {
		panic(error)
	}

	if caseData.People != nil {
		for id, person := range caseData.People {
			a.fakePeopleConfig[id] = &config.PeopleConfig{
				FriendlyName: person.FriendlyName,
				Devices:      person.Devices,
				Attributes:   person.Attributes,
				State:        person.State,
			}
			if a.fakePeopleConfig[id].Attributes == nil {
				a.fakePeopleConfig[id].Attributes = map[string]interface{}{}
			}
		}
	}

	if caseData.Devices != nil {
		for deviceID, device := range caseData.Devices {
			a.fakeDevices[deviceID] = &client.HassEntity{
				ID:   deviceID,
				Name: deviceID,
				New: client.HassEntityState{
					State:      device.State,
					Attributes: device.Attributes,
				},
			}
			if a.fakeDevices[deviceID].New.Attributes == nil {
				a.fakeDevices[deviceID].New.Attributes = map[string]interface{}{}
			}
		}
	}

}

// Config is the main configuration data structure
type testCaseConfig struct {
	People  map[string]*peopleConfig  `yaml:"people"`
	Devices map[string]*devicesConfig `yaml:"devices"`
}

// HomeAssistantConfig is the configuration for the Home Assistant platform integration
// type testCaseConfig struct {
// 	IP    string `yaml:"ip"`
// 	SSL   bool   `yaml:"ssl"`
// 	Token string `yaml:"token"`
// }

// PeopleConfig is the configuration for the Home Assistant platform integration
type peopleConfig struct {
	FriendlyName string                 `yaml:"friendly_name"`
	Devices      []string               `yaml:"devices"`
	State        string                 `yaml:"state"`
	Attributes   map[string]interface{} `yaml:"attributes"`
}

type devicesConfig struct {
	State      string                 `yaml:"state"`
	Attributes map[string]interface{} `yaml:"attributes"`
}
