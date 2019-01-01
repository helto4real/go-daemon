/*
Package app implements an application "ExampleApp"

Applications

Applications in go-daemon implements any functionality provided by go-language
in home control software compatible with Home Assistant websocket API. All
applications need to be pre-added in the "apps.go" file since go is a compiled
language and does not support dynamic loading.

You can use your own go routines to do work, make sure you exit them by using the
"Cancel" function to provide the nescessary cancellation logic.

The "DaemonAppHelper" contains all logic that integrates with Home Assistant and the
"DeamonAppConfig" contains the instance specific configuration found in the config.yaml

Note:
	The "Initialize" and "Cancel" functions have to be present to comply with interface
*/
package app

import (
	"context"

	"time"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
	"github.com/sirupsen/logrus"
)

// ExampleApp implements an go-appdaemon app
// This app takes a light and logs its state changes
type ExampleApp struct {
	deamon        d.DaemonAppHelper
	cfg           d.DeamonAppConfig
	state         chan c.HassEntity
	sunset        chan bool
	sunrise       chan bool
	cancel        context.CancelFunc
	cancelContext context.Context
	timer         *time.Timer
}

// Initialize is called when an application is started
//
// Use this to initialize your application, like subscribe to
// changes in entities and events
func (a *ExampleApp) Initialize(helper d.DaemonAppHelper, config d.DeamonAppConfig) bool {
	// Save the daemon helper and config to variables for later use
	a.deamon = helper
	a.cfg = config
	// Make the channel all state changes we listen too will be sent to
	a.state = make(chan c.HassEntity, 1)
	// Make the sunset and sunrise channels
	a.sunset = make(chan bool, 1)
	a.sunrise = make(chan bool, 1)

	// Make a cancelation context to use when the application need to close
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.cancelContext = ctx

	// Listen to state changes to the entity configured
	// in the config yaml file
	//a.deamon.ListenState(a.cfg.Properties["tomas_room_light"], a.state)
	a.deamon.ListenState(a.cfg.Properties["tomas_motion_sensor"], a.state)
	//a.deamon.ListenState("sun.sun", a.state)
	a.deamon.AtSunset(time.Duration(-1)*time.Hour, a.sunset)
	a.deamon.AtSunrise(time.Duration(30)*time.Minute, a.sunrise)
	// Do state change logic in own go-routine and return from initializaiotn
	// Initialize function should never block
	go a.handleStateChanges()
	log.Println("Example app initialized!")
	return true
}

func (a *ExampleApp) handleStateChanges() {

	for {
		select {
		case entity, ok := <-a.state:
			if !ok {
				return
			}
			if entity.New.State != entity.Old.State {
				a.handleEntityState(entity)
			}
		case <-a.sunrise:
			log.Println("SUNRISE!")
			// Reschedule
			a.deamon.AtSunrise(time.Duration(30)*time.Minute, a.sunrise)
		case <-a.sunset:
			log.Println("SUNSET!")
			// Reschedule
			a.deamon.AtSunset(time.Duration(-1)*time.Hour, a.sunset)
		// Listen to the cancelation context and leave when canceled
		case <-a.cancelContext.Done():
			return
		}
	}
}

func (a *ExampleApp) handleEntityState(entity client.HassEntity) {
	motionsensor := a.cfg.Properties["tomas_motion_sensor"]
	light := a.cfg.Properties["tomas_room_light"]

	if entity.ID == motionsensor {
		if entity.New.State == "on" {

			a.deamon.TurnOn(light)
			if a.timer != nil {
				log.Printf("Retting timer off to %v", time.Now().Add(time.Minute*2).Local())
				a.timer.Reset(time.Minute * 2)
			} else {
				log.Printf("Setting timer off to %v", time.Now().Add(time.Minute*2).Local())
				// Call turn off in 2 minute
				a.timer = time.AfterFunc(time.Minute*2, func() {
					a.deamon.TurnOff(light)
				})
			}

		}
	}
}

// Cancel the application during shutdown the go-appdaemon
//
// Implement any cancelation logic here if needed
func (a *ExampleApp) Cancel() {
	// Cancel the goroutine select
	a.cancel()
}

var log *logrus.Entry

// init is called first in all packages. This setup the logging to use prefix
func init() {
	log = logrus.WithField("prefix", "example_app")
}
