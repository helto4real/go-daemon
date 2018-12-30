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
	d "github.com/helto4real/go-daemon/daemon"
)

// ExampleApp implements an go-appdaemon app
type ExampleApp struct {
	deamon d.DaemonAppHelper
	cfg    d.DeamonAppConfig
}

// Initialize is called when an application is started
//
// Use this to initialize your application, like subscribe to
// changes in entities and events
func (a *ExampleApp) Initialize(helper d.DaemonAppHelper, config d.DeamonAppConfig) bool {
	// Save the daemon helper and config to variables for later use
	a.deamon = helper
	a.cfg = config

	a.deamon.Toggle(a.cfg.Properties["tomas_room_light"])
	return true
}

// Cancel the application during shutdown the go-appdaemon
//
// Implement any cancelation logic here if needed
func (a *ExampleApp) Cancel() {

}
