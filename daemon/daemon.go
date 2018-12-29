package daemon

import (
	"context"
	"log"

	"github.com/helto4real/go-daemon/daemon/config"
	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
)

type ApplicationDaemon struct {
	hassClient    c.HomeAssistant
	config        *config.Config
	cancel        context.CancelFunc
	cancelContext context.Context
}

// Start the daemon, use in main function
func (a *ApplicationDaemon) Start(configPath string, hassClient c.HomeAssistant) bool {
	a.hassClient = hassClient
	ctx, cancel := context.WithCancel(context.Background())

	a.cancelContext = ctx
	a.cancel = cancel

	configuration := config.NewConfiguration(configPath)
	conf, err := configuration.Open()

	if err != nil {
		log.Print("Failed to open config file, ending", err)
		return false
	}
	a.config = conf
	a.hassClient.Start(conf.HomeAssistant.IP, conf.HomeAssistant.SSL, conf.HomeAssistant.Token)
	return true
}

// Stop the daemon, only use in main function
func (a *ApplicationDaemon) Stop() {
	a.cancel()
	a.hassClient.Stop()

}

// RegisterApplication registers a new daemon application in the appdaemon
func (a *ApplicationDaemon) RegisterApplication(app DaemonApplication) {

}

func NewApplicationDaemon() ApplicationDaemonRunner {
	return &ApplicationDaemon{}
}

// GetCancelContext gets the context for goroutines to use as cancel context
func (a *ApplicationDaemon) GetCancelContext() context.Context {
	return a.cancelContext
}

// GetCancelFunction gets the cancel function for the whole daemon
// Should not be used unless intend to close the daemon from an app
func (a *ApplicationDaemon) GetCancelFunction() context.CancelFunc {
	return a.cancel
}

// GetEntity returns the state of a entity
func (a *ApplicationDaemon) GetEntity(entity string) (client.HassEntity, bool) {
	return a.hassClient.GetEntity(entity)
}

// TurnsOn turns on an entity with no attributes
func (a *ApplicationDaemon) TurnOn(entity string) {
	a.hassClient.CallService("turn_on", map[string]string{"entity_id": entity})
}

// TurnOff turns off an entity with no attributes
func (a *ApplicationDaemon) TurnOff(entity string) {
	a.hassClient.CallService("turn_off", map[string]string{"entity_id": entity})
}

// Toggle toggles an entity with no attributes
func (a *ApplicationDaemon) Toggle(entity string) {
	a.hassClient.CallService("toggle", map[string]string{"entity_id": entity})
}
