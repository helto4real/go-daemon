package daemon

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/helto4real/go-daemon/daemon/config"
	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
	yaml "gopkg.in/yaml.v2"
)

type DaemonCommand int

const (
	StartApplications   DaemonCommand = 0
	StopApplications    DaemonCommand = 1
	ReStartApplications DaemonCommand = 2
)

type DeamonAppConfig struct {
	App        string            `yaml:"app"`
	Properties map[string]string `yaml:"properties"`
}

type ApplicationDaemon struct {
	hassClient     c.HomeAssistant
	config         *config.Config
	cancel         context.CancelFunc
	cancelContext  context.Context
	configPath     string
	commandChannel chan DaemonCommand
	applications   []DaemonApplication
	availableApps  map[string]interface{}
}

// Start the daemon, use in main function
func (a *ApplicationDaemon) Start(configPath string, hassClient c.HomeAssistant, availableApps map[string]interface{}) bool {
	a.hassClient = hassClient
	ctx, cancel := context.WithCancel(context.Background())
	a.commandChannel = make(chan DaemonCommand)
	a.cancelContext = ctx
	a.cancel = cancel
	a.configPath = configPath
	a.applications = []DaemonApplication{}
	a.availableApps = availableApps
	configuration := config.NewConfiguration(filepath.Join(configPath, "go-daemon.yaml"))
	conf, err := configuration.Open()

	if err != nil {
		log.Print("Failed to open config file, ending", err)
		return false
	}
	a.config = conf
	go a.receiveHassLoop()
	go a.applicationDaemonLoop()
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

func (a *ApplicationDaemon) NewDaemonApp(appName string) (DaemonApplication, bool) {
	if f, exist := a.availableApps[appName]; exist {
		instance := reflect.New(reflect.TypeOf(f)).Interface()
		dApp, ok := instance.(DaemonApplication)
		if ok {
			return dApp, true
		}
	}
	return nil, false
}

func (a *ApplicationDaemon) receiveHassLoop() {
	hassStatusChannel := a.hassClient.GetStatusChannel()
	hassEntityChannel := a.hassClient.GetEntityChannel()
	commandChannel := a.commandChannel
	for {
		select {
		case status, mc := <-hassStatusChannel:
			if mc {
				if status {
					// We got connected
					//a.loadDaemonApplications()
					commandChannel <- StartApplications
				} else {
					// We disconnected
					//a.unloadDaemonApplications()
					commandChannel <- StopApplications
				}
			}
		case _, mc := <-hassEntityChannel:
			if mc {
				//log.Println("Got entity: ", status)
			}
		case <-a.cancelContext.Done():
			return
		}
	}
}
func (a *ApplicationDaemon) applicationDaemonLoop() {
	commandChannel := a.commandChannel
	for {
		select {

		case command, mc := <-commandChannel:
			if mc {
				switch command {
				case StartApplications:
					a.loadDaemonApplications()
				case StopApplications:
					a.unloadDaemonApplications()
				}
			}
		case <-a.cancelContext.Done():
			return
		}
	}
}

func (a *ApplicationDaemon) loadDaemonApplications() {
	log.Println("Loading applications...")
	if len(a.applications) > 0 {
		a.unloadDaemonApplications()
	}
	a.applications = a.instanceAllApplications()
}
func (a *ApplicationDaemon) unloadDaemonApplications() {
	log.Println("Unloading applications...")
	// Remove all subscriptions here

	// Remove the applications
	if len(a.applications) > 0 {
		for _, app := range a.applications {
			app.Cancel()
		}
		// Get new instance of empty list
		a.applications = []DaemonApplication{}
	}
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

func (a *ApplicationDaemon) instanceAllApplications() []DaemonApplication {
	applicationInstances := []DaemonApplication{}

	allApplicationConfigs := a.getAllApplicationConfigFilePaths()

	for _, configFile := range allApplicationConfigs {
		cfgList, ok := a.getConfigFromFile(configFile)
		if ok {
			for _, appCfg := range cfgList {
				app, ok := a.NewDaemonApp(appCfg.App)
				if ok {
					log.Println("Loading application: ", appCfg.App)
					applicationInstances = append(applicationInstances, app)
					app.Initialize(a, appCfg)
				} else {
					log.Printf("Did not find the application {%s}, please check config in [%s] ", appCfg.App, configFile)
				}
			}
		}

	}
	return applicationInstances
}
