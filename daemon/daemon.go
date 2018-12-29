package daemon

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/helto4real/go-daemon/daemon/config"
	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
	yaml "gopkg.in/yaml.v2"
)

type DeamonAppConfig struct {
	App        string            `yaml:"app"`
	Properties map[string]string `yaml:"properties"`
}

type ApplicationDaemon struct {
	hassClient    c.HomeAssistant
	config        *config.Config
	cancel        context.CancelFunc
	cancelContext context.Context
	configPath    string
}

// Start the daemon, use in main function
func (a *ApplicationDaemon) Start(configPath string, hassClient c.HomeAssistant) bool {
	a.hassClient = hassClient
	ctx, cancel := context.WithCancel(context.Background())

	a.cancelContext = ctx
	a.cancel = cancel
	a.configPath = configPath
	configuration := config.NewConfiguration(filepath.Join(configPath, "go-daemon.yaml"))
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

// TurnOn turns on an entity with no attributes
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

func (a *ApplicationDaemon) getAllApplicationConfigFilePaths() []string {
	fileList := []string{}
	pathAppDir := filepath.Join(a.configPath, "app")
	err := filepath.Walk(pathAppDir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			fileList = append(fileList, path)
		}

		return nil
	})

	if err != nil {
		log.Print("Failed to get all the configuration files.", err)
	}

	return fileList
}

func (a *ApplicationDaemon) getConfigFromFile(path string) (map[string]DeamonAppConfig, bool) {
	i := make(map[string]DeamonAppConfig, 1)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Failed to read app yaml file", path, err)
		return nil, false
	}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		log.Print("Failed to parse app yaml file", path, err)
		return nil, false
	}
	return i, true
}
