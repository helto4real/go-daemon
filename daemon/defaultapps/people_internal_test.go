package defaultapps

import (
	"context"
	"testing"
	"time"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-daemon/daemon/config"
	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
)

func TestInitialize(t *testing.T) {
	app := PeopleApp{}
	fake := fakeDaemonAppHelper{}
	app.Initialize(&fake, d.DeamonAppConfig{
		App:        "fake_app",
		Properties: make(map[string]string),
	})
	h.Equals(t, fake.listenState, 4)
	h.Equals(t, fake.getEntity, 4)
	h.Equals(t, fake.setEntity, 2)
}

type fakeDaemonAppHelper struct {
	listenState int
	getEntity   int
	setEntity   int
}

func (a *fakeDaemonAppHelper) GetCancelContext() context.Context {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) GetCancelFunction() context.CancelFunc {
	panic("not implemented")
}

func (a *fakeDaemonAppHelper) GetEntity(entity string) (*client.HassEntity, bool) {
	a.getEntity = a.getEntity + 1
	if entity == "tracker.person1_one" {
		return &client.HassEntity{
			ID:   "tracker.person1_one",
			Name: "tracker.person1_one"}, true
	} else if entity == "tracker.person1_two" {
		return &client.HassEntity{
			ID:   "tracker.person1_two",
			Name: "tracker.person1_two"}, true
	} else if entity == "tracker.person2_one" {
		return &client.HassEntity{
			ID:   "tracker.person2_one",
			Name: "tracker.person2_one"}, true
	} else if entity == "tracker.person2_two" {
		return &client.HassEntity{
			ID:   "tracker.person2_two",
			Name: "tracker.person2_two"}, true
	}
	return nil, false
}

func (a *fakeDaemonAppHelper) SetEntity(entity *client.HassEntity) bool {
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
	fakePeopleConfig := make(map[string]*config.PeopleConfig)

	fakePeopleConfig["person1"] = &config.PeopleConfig{
		FriendlyName: "friendly1",
		State:        "home",
		Devices:      []string{"tracker.person1_one", "tracker.person1_two"},
		Attributes:   make(map[string]interface{}),
	}

	fakePeopleConfig["person2"] = &config.PeopleConfig{
		FriendlyName: "friendly2",
		State:        "not_home",
		Devices:      []string{"tracker.person2_one", "tracker.person2_two"},
		Attributes:   make(map[string]interface{}),
	}

	return fakePeopleConfig
}

func (a *fakeDaemonAppHelper) GetLocation() d.Location {
	panic("not implemented")
}
