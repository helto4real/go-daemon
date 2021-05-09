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
	"math"
	"sort"
	"strings"

	"github.com/helto4real/go-hassclient/client"

	"time"

	d "github.com/helto4real/go-daemon/daemon"
	c "github.com/helto4real/go-daemon/daemon/config"
	"github.com/sirupsen/logrus"
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
	settings      *c.SettingsConfig
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
	a.settings = helper.GetSettings()

	// Make a cancelation context to use when the application need to close
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.cancelContext = ctx

	a.trackerChannel = make(chan client.HassEntity, 10)
	a.stateChangedChannel = make(chan string, 2)
	// Update state for all persons
	for name := range a.conf {
		a.handleUpdatedDeviceForPerson(name, false)
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
			a.handleUpdatedDevice(entity.ID, false)

		case person, ok := <-a.stateChangedChannel:
			if !ok {
				return
			}
			a.handleUpdatedDeviceForPerson(person, true)
		// Listen to the cancelation context and leave when canceled
		case <-a.cancelContext.Done():
			return
		}
	}
}
func (a *PeopleApp) handleUpdatedDevice(entityID string, isFromTimeout bool) {
	// Get the person owning device
	person := a.getPersonOwningDevice(entityID)
	a.handleUpdatedDeviceForPerson(person, isFromTimeout)
}

func (a *PeopleApp) handleUpdatedDeviceForPerson(person string, isFromTimeout bool) {
	// Get devices
	devices := a.getDeviceEntities(person)

	state := a.getHassDeviceState(devices)

	if state != "home" {
		if a.conf[person].State == "" {
			a.setState(person, state, devices)
		} else if a.conf[person].State == a.settings.TrackingSettings.JustLeftState {
			if isFromTimeout {
				a.setState(person, state, devices)
			} else {
				// Use same state since we are not setting from timeout
				a.setState(person, a.conf[person].State, devices)
			}

		} else if a.conf[person].State == a.settings.TrackingSettings.HomeState {
			// We were home and just left
			a.setState(person, a.settings.TrackingSettings.JustLeftState, devices)

			time.AfterFunc(time.Second*time.Duration(a.settings.TrackingSettings.JustLeftTime), func() {
				if a.conf[person].State == a.settings.TrackingSettings.JustLeftState {
					a.stateChangedChannel <- person
				}
			})

		} else {
			a.setState(person, state, devices)
		}
	} else {
		//Home

		if a.conf[person].State == "" {
			a.setState(person, state, devices)
		} else if a.conf[person].State == a.settings.TrackingSettings.JustArrivedState {
			if isFromTimeout {
				a.setState(person, state, devices)
			} else {
				// Use same state since we are not setting from timeout
				a.setState(person, a.conf[person].State, devices)

			}
		} else if a.conf[person].State != a.settings.TrackingSettings.HomeState {
			// We were home and just left
			a.setState(person, a.settings.TrackingSettings.JustArrivedState, devices)

			time.AfterFunc(time.Second*time.Duration(a.settings.TrackingSettings.JustArrivedTime), func() {
				if a.conf[person].State == a.settings.TrackingSettings.JustArrivedState {
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
		personState = a.settings.TrackingSettings.HomeState
	} else if state == "not_home" {
		personState = a.settings.TrackingSettings.AwayState
	} else {
		personState = state
	}
	a.conf[person].State = personState

	sortedDevices := devices
	sort.Slice(sortedDevices, func(i, j int) bool { return devices[i].New.LastChanged.After(devices[j].New.LastUpdated) })
	hasLocation := false
	for _, device := range sortedDevices {
		if device.New.Attributes["source_type"] == "gps" {
			// Copy attributes if exists
			longitude, ok := device.New.Attributes["longitude"]
			if ok {
				hasLocation = true
				a.conf[person].Attributes["longitude"] = longitude
			}
			latitude, ok := device.New.Attributes["latitude"]
			if ok {
				hasLocation = true
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
	if hasLocation {
		longitude := a.conf[person].Attributes["longitude"].(float64)
		latitude := a.conf[person].Attributes["latitude"].(float64)

		homeLocation := a.deamon.GetLocation()
		distance := distance(latitude, longitude, homeLocation.Latitude, homeLocation.Longitude, "K")
		a.conf[person].Attributes["distance"] = math.Round(distance)

		a.conf[person].Attributes["source_type"] = "gps"
	}

	deviceID := getDeviceID(person)
	entity := client.NewHassEntity(deviceID, deviceID, client.HassEntityState{}, client.HassEntityState{
		State:      a.conf[person].State,
		Attributes: a.conf[person].Attributes,
	})
	a.deamon.SetEntity(entity)
	log.Debugln(entity)
}
func getDeviceID(person string) string {
	return "device_tracker." + strings.ToLower(person) + "_presence"
}
func (a *PeopleApp) getHassDeviceState(devices []*client.HassEntity) string {
	sortedDevices := devices
	// Get devices
	sort.Slice(sortedDevices, func(i, j int) bool { return devices[i].New.LastChanged.After(devices[j].New.LastChanged) })

	for _, device := range sortedDevices {
		if translateState(device.New.State) == "home" {
			sourceType, ok := device.New.Attributes["source_type"]
			if ok {
				if sourceType != "gps" {
					// Ether bt or wifi are home, device always home, this will make
					// the tracking alot more stable
					return "home"
				} else if time.Now().UTC().Sub(device.New.LastUpdated).Minutes() < 60 {
					// If the gps device was updated recently if the gps is not reporting
					// and last state was "home" we want to avoid getting stuck at home
					return "home"
				}
			} else {
				// No attributes, it is not a gps device
				return "home"
			}
		}
	}

	// If we reached this point all devices are considered not_home
	// Get the state from gps device
	gpsDevice := getGpsSourceTypeDevice(sortedDevices)

	if gpsDevice != nil {
		// Return the gps device state
		return translateState(gpsDevice.New.State)
	}
	if len(sortedDevices) > 0 {
		// Just return the last changed device state
		return translateState(sortedDevices[0].New.State)
	} else {
		return "not_home"
	}
}

func getGpsSourceTypeDevice(devices []*client.HassEntity) *client.HassEntity {
	// None of the devices is home, take value from gps device
	for _, device := range devices {
		sourceType, ok := device.New.Attributes["source_type"]
		if ok && sourceType == "gps" {
			return device
		}
	}
	return nil
}

func translateState(state string) string {
	stateLower := strings.ToLower(state)

	if stateLower == "home" {
		return "home"
	} else if stateLower == "not_home" {
		return "not_home"
	} else if stateLower == "on" {
		return "home"
	} else if stateLower == "off" {
		return "not_home"
	} else if stateLower == "true" {
		return "home"
	} else if stateLower == "false" {
		return "not_home"
	}
	return state
}

func (a *PeopleApp) getPersonOwningDevice(device string) string {
	for name, person := range a.conf {
		for _, dev := range person.Devices {
			if dev == device {
				return name
			}
		}
	}
	log.Errorf("Device unknown, please check configuration [%s]", device)
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

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = dist * 1.609344
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}

var log *logrus.Entry

// init is called first in all packages. This setup the logging to use prefix
func init() {
	log = logrus.WithField("prefix", "default people_app")
}
