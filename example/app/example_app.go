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
	"log"
	"time"

	d "github.com/helto4real/go-daemon/daemon"
	c "github.com/helto4real/go-hassclient/client"
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
	a.deamon.ListenState(a.cfg.Properties["tomas_room_light"], a.state)
	//a.deamon.ListenState(a.cfg.Properties["tomas_motion_sensor"], a.state)
	//a.deamon.ListenState("sun.sun", a.state)
	a.deamon.AtSunset(time.Duration(-1)*time.Hour, a.sunset)
	a.deamon.AtSunrise(time.Duration(30)*time.Minute, a.sunrise)
	// Do state change logic in own go-routine and return from initializaiotn
	// Initialize function should never block
	go a.handleStateChanges()

	return true
}

func (a *ExampleApp) handleStateChanges() {

	for {
		select {
		case entity, ok := <-a.state:
			if ok {
				log.Print(entity)
				if entity.New.State != entity.Old.State {
					// Only changed states handled
					log.Printf("State of %s changed from %s to: %s", entity.ID, entity.Old.State, entity.New.State)
				} else {
					log.Printf("State of %s same from %s to: %s", entity.ID, entity.Old.State, entity.New.State)
				}
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

// Cancel the application during shutdown the go-appdaemon
//
// Implement any cancelation logic here if needed
func (a *ExampleApp) Cancel() {
	// Cancel the goroutine select
	a.cancel()
}
