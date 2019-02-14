package core

import (
	"context"
	"io/ioutil"
	"path"
	"sync"
	"testing"
	"time"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-daemon/daemon/config"
	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
	yaml "gopkg.in/yaml.v2"
)

func TestNewEntity(t *testing.T) {
	fake := newFakeDaemonHelperTestCase("testcase1.yml") //.(d.DaemonAppHelper)

	changedEntityChannel := make(chan d.DaemonEntity)
	entity := NewEntity("device_tracker.gps", fake, false, changedEntityChannel)

	h.Equals(t, "device_tracker.gps", entity.ID())
	h.Equals(t, "gps", entity.Attributes()["source_type"])
	h.Equals(t, "device_tracker.gps", entity.Entity().Name)
	fake.cancel()
}

func TestNewEntityUnknown(t *testing.T) {
	fake := newFakeDaemonHelperTestCase("testcase1.yml") //.(d.DaemonAppHelper)

	changedEntityChannel := make(chan d.DaemonEntity)
	entity := NewEntity("sensor.not_exist", fake, false, changedEntityChannel)

	h.Equals(t, "sensor.not_exist", entity.ID())
	h.Equals(t, "unknown", entity.State())
	fake.cancel()
}

func TestNewEntityFromChannel(t *testing.T) {
	fake := newFakeDaemonHelperTestCase("testcase1.yml") //.(d.DaemonAppHelper)
	defer fake.cancel()

	changedEntityChannel := make(chan d.DaemonEntity, 2)
	entity := NewEntity("device_tracker.gps", fake, false, changedEntityChannel)

	fake.stateChannel <- *client.NewHassEntity(
		entity.ID(),
		entity.ID(),
		client.HassEntityState{},
		client.HassEntityState{
			State:      "NewState",
			Attributes: map[string]interface{}{"latitude": 5.0},
		})

	time.Sleep(time.Millisecond * 100)
	fake.confMutex.Lock()
	defer fake.confMutex.Unlock()

	h.Equals(t, 5.0, entity.Entity().New.Attributes["latitude"])
}

type fakeDaemonAppHelper struct {
	listenState      int
	cancel           context.CancelFunc
	cancelContext    context.Context
	getEntity        int
	setEntity        int
	fakePeopleConfig map[string]*config.PeopleConfig
	fakeDevices      map[string]*client.HassEntity
	stateChannel     chan client.HassEntity
	confMutex        *sync.Mutex
}

func newFakeDaemonHelper() *fakeDaemonAppHelper {
	apphelper := &fakeDaemonAppHelper{
		fakePeopleConfig: map[string]*config.PeopleConfig{},
		fakeDevices:      map[string]*client.HassEntity{},
		confMutex:        &sync.Mutex{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	apphelper.cancelContext = ctx
	apphelper.cancel = cancel
	return apphelper
}

func newFakeDaemonHelperTestCase(filename string) *fakeDaemonAppHelper {
	apphelper := &fakeDaemonAppHelper{
		fakePeopleConfig: map[string]*config.PeopleConfig{},
		fakeDevices:      map[string]*client.HassEntity{},
		confMutex:        &sync.Mutex{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	apphelper.cancelContext = ctx
	apphelper.cancel = cancel

	apphelper.loadTestCase(filename)

	return apphelper
}

func (a *fakeDaemonAppHelper) GetCancelContext() context.Context {
	return a.cancelContext
}

func (a *fakeDaemonAppHelper) GetCancelFunction() context.CancelFunc {
	return a.cancel
}

func (a *fakeDaemonAppHelper) GetEntity(entity string) (*client.HassEntity, bool) {
	a.confMutex.Lock()
	defer a.confMutex.Unlock()

	a.getEntity = a.getEntity + 1

	if a.fakeDevices != nil && len(a.fakeDevices) > 0 {
		enityFromTestCase, ok := a.fakeDevices[entity]
		if !ok {
			return nil, false
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
	a.stateChannel = stateChannel
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
