/*
Package defaultapps implements the default built in applications in go-daemon

Applications

The following applications are implemented
- presence:
	Implements the people app that keeps track of people like presence information
*/
package defaultapps

import (
	"context"
	"sort"
	"strings"

	"github.com/helto4real/go-hassclient/client"

	"time"

	d "github.com/helto4real/go-daemon/daemon"
	c "github.com/helto4real/go-daemon/daemon/config"
	"github.com/sirupsen/logrus"
)

const (
	justArrivedState string        = "Just Arrived"
	justLeftState    string        = "Just Left"
	homeState        string        = "Home"
	awayState        string        = "Away"
	justTimer        time.Duration = 300 //5 minutes
)

type personState struct {
	state      string
	attributes map[string]interface{}
}

func newState(state string) *personState {
	return &personState{
		state:      state,
		attributes: map[string]interface{}{}}
}

// PeopleApp implements an go-appdaemon app that keeps track
// of peoples presence information. Data is available in daemon it self
// for other apps
type PeopleApp struct {
	deamon        d.DaemonAppHelper
	conf          map[string]*c.PeopleConfig
	cancel        context.CancelFunc
	cancelContext context.Context
	timer         *time.Timer
	// trackerChannel is the channel where tracker updates will come
	trackerChannel      chan client.HassEntity
	stateChangedChannel chan string
}

// Initialize is called when an application is started
//
// Use this to initialize your application, like subscribe to
// changes in entities and events
func (a *PeopleApp) Initialize(helper d.DaemonAppHelper, config d.DeamonAppConfig) bool {
	// Save the daemon helper and config to variables for later use
	a.deamon = helper
	a.conf = helper.GetPeople()

	// Make a cancelation context to use when the application need to close
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.cancelContext = ctx

	a.trackerChannel = make(chan client.HassEntity, 10)
	a.stateChangedChannel = make(chan string, 2)
	// Update state for all persons
	for name := range a.conf {
		a.handleUpdatedDeviceForPerson(name)
	}

	a.listenToDevices()

	// Run loop in own goroutine
	go a.loop()
	log.Println("Default people app initialized!")

	return true
}

func (a *PeopleApp) loop() {

	for {
		select {
		case entity, ok := <-a.trackerChannel:
			if !ok {
				return
			}
			a.handleUpdatedDevice(entity.ID)

		case person, ok := <-a.stateChangedChannel:
			if !ok {
				return
			}
			a.handleUpdatedDeviceForPerson(person)
		// Listen to the cancelation context and leave when canceled
		case <-a.cancelContext.Done():
			return
		}
	}
}
func (a *PeopleApp) handleUpdatedDevice(entityID string) {
	// Get the person owning device
	person := a.getPersonOwningDevice(entityID)
	a.handleUpdatedDeviceForPerson(person)
}

func (a *PeopleApp) handleUpdatedDeviceForPerson(person string) {
	// Get devices
	devices := a.getDeviceEntities(person)

	state := a.getHassDeviceState(devices)

	if state != "home" {
		if a.conf[person].State == "" ||
			a.conf[person].State == justArrivedState ||
			a.conf[person].State == justLeftState {
			a.setState(person, state, devices)
		} else if a.conf[person].State == homeState {
			// We were home and just left
			a.setState(person, justLeftState, devices)

			time.AfterFunc(time.Second*justTimer, func() {
				if a.conf[person].State == justLeftState {
					a.stateChangedChannel <- person
				}
			})

		} else {
			a.setState(person, state, devices)
		}
	} else {
		//Home
		if a.conf[person].State == "" || a.conf[person].State == justArrivedState || a.conf[person].State == justLeftState {
			a.setState(person, state, devices)
		} else if a.conf[person].State != homeState {
			// We were home and just left
			a.setState(person, justArrivedState, devices)

			time.AfterFunc(time.Second*justTimer, func() {
				if a.conf[person].State == justArrivedState {
					a.stateChangedChannel <- person
				}
			})
		} else {
			a.setState(person, state, devices)
		}
	}

}
func (a *PeopleApp) setState(person string, state string, devices []*client.HassEntity) {
	var personState string
	if state == "home" {
		personState = homeState
	} else if state == "not_home" {
		personState = awayState
	} else {
		personState = state
	}
	a.conf[person].State = personState

	sortedDevices := devices
	sort.Slice(sortedDevices, func(i, j int) bool { return devices[i].New.LastChanged.After(devices[j].New.LastUpdated) })

	for _, device := range sortedDevices {
		if device.New.Attributes["source_type"] == "gps" {
			// Copy attributes if exists
			longitude, ok := device.New.Attributes["longitude"]
			if ok {
				a.conf[person].Attributes["longitude"] = longitude
			}
			latitude, ok := device.New.Attributes["latitude"]
			if ok {
				a.conf[person].Attributes["latitude"] = latitude
			}
			picture, ok := device.New.Attributes["entity_picture"]
			if ok {
				a.conf[person].Attributes["entity_picture"] = picture
			}
			address, ok := device.New.Attributes["address"]
			if ok {
				a.conf[person].Attributes["address"] = address
			}
			batteryLevel, ok := device.New.Attributes["battery_level"]
			if ok {
				a.conf[person].Attributes["battery_level"] = batteryLevel
			}
			gpsAccuracy, ok := device.New.Attributes["gps_accuracy"]
			if ok {
				a.conf[person].Attributes["gps_accuracy"] = gpsAccuracy
			}
			break
		}
	}
	a.conf[person].Attributes["friendly_name"] = a.conf[person].FriendlyName

	deviceID := getDeviceID(person)
	entity := client.NewHassEntity(deviceID, deviceID, client.HassEntityState{}, client.HassEntityState{
		State:      a.conf[person].State,
		Attributes: a.conf[person].Attributes,
	})
	a.deamon.SetEntity(entity)
	log.Warn(entity)
}
func getDeviceID(person string) string {
	return "sensor." + strings.ToLower(person) + "_presence"
}
func (a *PeopleApp) getHassDeviceState(devices []*client.HassEntity) string {
	sortedDevices := devices
	// Get devices
	sort.Slice(sortedDevices, func(i, j int) bool { return devices[i].New.LastChanged.After(devices[j].New.LastChanged) })

	for _, device := range sortedDevices {
		sourceType, ok := device.New.Attributes["source_type"]
		if ok {
			if device.New.State == "home" &&
				(sourceType == "bluetooth" || sourceType == "router") {
				// Ether bt or wifi are home, device always home, this will make
				// the tracking alot more stable
				return "home"
			}
		}
	}
	return sortedDevices[0].New.State
}

func (a *PeopleApp) getPersonOwningDevice(device string) string {
	for name, person := range a.conf {
		for _, dev := range person.Devices {
			if dev == device {
				return name
			}
		}
	}
	log.Panicf("Device unknown, please check configuration [%s]", device)
	return ""
}

func (a *PeopleApp) listenToDevices() {
	if !a.peopleConfigured() {
		return
	}
	for _, person := range a.conf {
		for _, device := range person.Devices {
			a.deamon.ListenState(device, a.trackerChannel)
		}
	}
}
func (a *PeopleApp) getDeviceEntities(person string) []*client.HassEntity {
	if !a.peopleConfigured() {
		return nil
	}
	ret := []*client.HassEntity{}
	for _, device := range a.conf[person].Devices {
		entity, ok := a.deamon.GetEntity(device)
		if ok {
			ret = append(ret, entity)
		} else {
			log.Errorf("Device [%s] does not exist!", device)
		}
	}
	return ret
}

func (a *PeopleApp) peopleConfigured() bool {
	if a.conf == nil || len(a.conf) == 0 {
		return false
	}
	return true
}

// Cancel the application during shutdown the go-appdaemon
//
// Implement any cancelation logic here if needed
func (a *PeopleApp) Cancel() {
	// Cancel the goroutine select
	a.cancel()
}

var log *logrus.Entry

// init is called first in all packages. This setup the logging to use prefix
func init() {
	log = logrus.WithField("prefix", "default people_app")
}
