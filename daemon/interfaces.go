package daemon

import (
	"context"
	"time"

	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
)

// DaemonApplication represents an application
type DaemonAppHelper interface {
	// GetCancelContext gets the context for goroutines to use as cancel context
	GetCancelContext() context.Context
	// GetCancelFunction gets the cancel function for the whole daemon
	// Should not be used unless intend to close the daemon from an app
	GetCancelFunction() context.CancelFunc

	// GetEntity returns the state of a entity
	GetEntity(entity string) (client.HassEntity, bool)

	// TurnsOn turns on an entity with no attributes
	TurnOn(entity string)

	// TurnOff turns off an entity with no attributes
	TurnOff(entity string)

	// Toggle toggles an entity with no attributes
	Toggle(entity string)

	// ListenState start listen to state changes from entity
	//
	// Any changes is reported back to the provided channel
	ListenState(entity string, stateChannel chan client.HassEntity)

	// AtSunset sends a message on provided channel at sunset
	//
	// You can set a positive or negative offset from sunset
	AtSunset(offset time.Duration, sunsetChannel chan bool) *time.Timer

	// AtSunrise sends a message on provided channel at sunset
	//
	// You can set a positive or negative offset from sunset
	AtSunrise(offset time.Duration, sunriseChannel chan bool) *time.Timer
}

// DaemonApplication represents an application
type ApplicationDaemonRunner interface {
	// Start daemon only use in main
	Start(configPath string, hassClient c.HomeAssistant, availableApps map[string]interface{}) bool
	// Stop daemon only use in main
	Stop()
}

// DaemonApplication represents an application
type DaemonApplication interface {
	Initialize(helper DaemonAppHelper, config DeamonAppConfig) bool
	Cancel()
}
