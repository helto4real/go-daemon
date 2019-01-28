package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-daemon/daemon/config"
	"github.com/helto4real/go-daemon/daemon/defaultapps"
	"github.com/helto4real/go-hassclient/client"
	c "github.com/helto4real/go-hassclient/client"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var log *logrus.Entry

type DaemonCommand int

const (
	StartApplications   DaemonCommand = 0
	StopApplications    DaemonCommand = 1
	ReStartApplications DaemonCommand = 2
)

type ApplicationDaemon struct {
	hassClient                c.HomeAssistant
	config                    *config.Config
	cancel                    context.CancelFunc
	cancelContext             context.Context
	configPath                string
	commandChannel            chan DaemonCommand
	applications              []d.DaemonApplication
	availableApps             map[string]interface{}
	stateListeners            map[string][]chan client.HassEntity
	callServiceEventListeners map[string]map[string][]chan client.HassCallServiceEvent
}

// Start the daemon, use in main function
func (a *ApplicationDaemon) Start(configPath string, hassClient c.HomeAssistant, availableApps map[string]interface{}) bool {
	a.hassClient = hassClient
	ctx, cancel := context.WithCancel(context.Background())
	a.commandChannel = make(chan DaemonCommand)
	a.cancelContext = ctx
	a.cancel = cancel
	a.configPath = configPath
	a.applications = []d.DaemonApplication{}
	a.availableApps = availableApps

	a.stateListeners = make(map[string][]chan client.HassEntity)
	a.callServiceEventListeners =
		make(map[string]map[string][]chan client.HassCallServiceEvent)

	configuration := config.NewConfiguration(filepath.Join(configPath, "go-daemon.yaml"))
	conf, err := configuration.Open()

	if err != nil {
		log.Error("Failed to open config file, ending -> ", err)
		return false
	}
	a.config = conf
	if a.config.HomeAssistant.IP == "hassio" {
		// It is a hassio plugin
		a.checkHassioOptionsConfig()
	}
	go a.receiveHassLoop()
	go a.applicationDaemonLoop()
	if len(conf.HomeAssistant.Token) == 0 {
		// Check if we have hassio env set
		envHassioToken := os.Getenv("HASSIO_TOKEN")
		if len(envHassioToken) == 0 {
			log.Warn("Token empty and hassio token not present. API wont be accessable if anonomous access not allowed!")
		}
		conf.HomeAssistant.Token = envHassioToken
	}
	a.hassClient.Start(conf.HomeAssistant.IP, conf.HomeAssistant.SSL, conf.HomeAssistant.Token)

	return true
}

// Stop the daemon, only use in main function
func (a *ApplicationDaemon) Stop() {
	a.cancel()
	a.hassClient.Stop()

}

var optionsPath = "/data/options.json"

func (a *ApplicationDaemon) checkHassioOptionsConfig() {

	confBytes, err := ioutil.ReadFile(fmt.Sprintf(optionsPath))
	if err != nil {
		log.Errorln(err)
		return
	}
	result := &config.HassioOptionsConfig{}
	err = json.Unmarshal(confBytes, result)
	if err != nil {
		log.Errorln(err)
		return
	}
	a.config.People = map[string]*config.PeopleConfig{}
	for _, person := range result.Persons {
		a.config.People[person.ID] = &config.PeopleConfig{
			FriendlyName: person.FriendlyName,
			Devices:      person.Devices,
			Attributes:   map[string]interface{}{},
		}
	}
	// Set the correct logger level
	if result.LogLevel == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	} else if result.LogLevel == "info" {
		logrus.SetLevel(logrus.InfoLevel)
	} else if result.LogLevel == "trace" {
		logrus.SetLevel(logrus.TraceLevel)
	} else if result.LogLevel == "warning" {
		logrus.SetLevel(logrus.WarnLevel)
	} else if result.LogLevel == "error" {
		logrus.SetLevel(logrus.ErrorLevel)
	} else if result.LogLevel == "fatal" {
		logrus.SetLevel(logrus.FatalLevel)
	}
}

func (a *ApplicationDaemon) GetLocation() d.Location {
	return d.Location{
		Longitude: a.hassClient.GetConfig().Longitude,
		Latitude:  a.hassClient.GetConfig().Latitude,
		Elevation: a.hassClient.GetConfig().Elevation,
	}
}

// AtSunset sends a message on provided channel at sunset
//
// You can set a positive or negative offset from sunset
func (a *ApplicationDaemon) AtSunset(offset time.Duration, sunsetChannel chan bool) *time.Timer {
	sun, ok := a.GetEntity("sun.sun")
	if !ok {
		log.Errorln("Failed to get sun.sun entity, cant set AtSunset!")
		return nil
	}

	sunset, ok := sun.New.Attributes["next_setting"].(string)
	if !ok {
		log.Errorln("Failed to get the attribute 'next_setting', catn set AtSunset!")
		return nil
	}
	t, err := time.Parse(time.RFC3339, sunset)

	if err != nil {
		log.Error("Failed to parse date", sunset)
		return nil
	}
	toffset := t.Add(offset)
	if toffset.Before(time.Now()) {
		// In some situations the time can be less that current time if using
		// negative offsets and the rescheduling is done in the right after
		// this event is set, we just add a day to the time if that happens
		toffset = toffset.Add(time.Hour * 24)
		log.Debug("We are before in time, adding 24 hours")
	}
	// Calculate duration until sunset
	dur := toffset.Sub(time.Now())
	log.Debugf("Next sunset event at %v, in %v ", toffset.Format("2006-01-02 15:04:05"), dur.Round(time.Second))

	return time.AfterFunc(dur, func() {
		sunsetChannel <- true
	})
}

// AtSunrise sends a message on provided channel at sunset
//
// You can set a positive or negative offset from sunset
func (a *ApplicationDaemon) AtSunrise(offset time.Duration, sunriseChannel chan bool) *time.Timer {
	sun, ok := a.GetEntity("sun.sun")
	if !ok {
		log.Errorln("Failed to get sun.sun entity, cant set AtSunrise!")
		return nil
	}

	sunrise, ok := sun.New.Attributes["next_rising"].(string)
	if !ok {
		log.Errorln("Failed to get the attribute 'next_rising', catn set AtSunrise!")
		return nil
	}
	t, err := time.Parse(time.RFC3339, sunrise)

	if err != nil {
		log.Errorln("Failed to parse date", sunrise)
		return nil
	}
	toffset := t.Add(offset)
	if toffset.Before(time.Now()) {
		// In some situations the time can be less that current time if using
		// negative offsets and the rescheduling is done in the right after
		// this event is set, we just add a day to the time if that happens
		toffset = toffset.Add(time.Hour * 24)
		log.Debug("We are before in time, adding 24 hours")
	}
	// Calculate duration until sunset
	dur := toffset.Sub(time.Now())

	log.Debugf("Next surise event at %v, in %v ", toffset.Format("2006-01-02 15:04:05"), dur.Round(time.Second))
	return time.AfterFunc(dur, func() {
		sunriseChannel <- true
	})
}

// ListenState start listen to state changes from entity
//
// Any changes is reported back to the provided channel
func (a *ApplicationDaemon) ListenCallServiceEvent(domain string, service string, callServiceChannel chan client.HassCallServiceEvent) {
	// Convert to lower case if some noob wrote it wrong
	domain = strings.ToLower(domain)
	service = strings.ToLower(service)

	domainCallServiceEventChannels, ok := a.callServiceEventListeners[domain]
	if !ok {
		domainCallServiceEventChannels =
			map[string][]chan client.HassCallServiceEvent{}
		a.callServiceEventListeners[domain] = domainCallServiceEventChannels
	}

	serviceChannels, ok := domainCallServiceEventChannels[service]
	if !ok {
		// First time we need to create the array
		domainCallServiceEventChannels[service] =
			[]chan client.HassCallServiceEvent{callServiceChannel}
		return
	}
	// We have existing, make sure channel not registered already
	for _, csChannel := range serviceChannels {
		if csChannel == callServiceChannel {
			// Allreade registered so return
			log.Errorf("ListenCallServiceEvent: Already registered on %s on current channel", service)
			return
		}
	}

	// Add the new channel
	domainCallServiceEventChannels[service] = append(serviceChannels, callServiceChannel)
}

// ListenState start listen to state changes from entity
//
// Any changes is reported back to the provided channel
func (a *ApplicationDaemon) ListenState(entity string, stateChannel chan client.HassEntity) {
	// Convert to lower case if some noob wrote it wrong
	entityLower := strings.ToLower(entity)

	stateChannels, ok := a.stateListeners[entityLower]
	if !ok {
		// First time we need to create the array
		a.stateListeners[entityLower] = []chan client.HassEntity{stateChannel}
		return
	}
	// We have existing, make sure channel not registered already
	for _, sChannel := range stateChannels {
		if sChannel == stateChannel {
			// Allreade registered so return
			log.Errorf("Listen state already registered on %s on current channel", entity)
			return
		}
	}

	// Add the new channel
	a.stateListeners[entityLower] = append(stateChannels, stateChannel)
}

func NewApplicationDaemon() d.ApplicationDaemonRunner {
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
func (a *ApplicationDaemon) GetEntity(entity string) (*client.HassEntity, bool) {
	return a.hassClient.GetEntity(entity)
}

// SetEntity creates or updates existing entity
func (a *ApplicationDaemon) SetEntity(entity *client.HassEntity) bool {
	return a.hassClient.SetEntity(entity)
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

func (a *ApplicationDaemon) GetPeople() map[string]*config.PeopleConfig {

	// First make sure that the yaml config doesent make nil map
	for _, personConfig := range a.config.People {
		if personConfig.Attributes == nil {
			personConfig.Attributes = map[string]interface{}{}
		}

	}
	return a.config.People
}

func (a *ApplicationDaemon) NewDaemonApp(appName string) (d.DaemonApplication, bool) {
	if f, exist := a.availableApps[appName]; exist {
		instance := reflect.New(reflect.TypeOf(f)).Interface()
		dApp, ok := instance.(d.DaemonApplication)
		if ok {
			return dApp, true
		}
	}
	return nil, false
}

func (a *ApplicationDaemon) receiveHassLoop() {
	hassStatusChannel := a.hassClient.GetStatusChannel()
	hassChannel := a.hassClient.GetHassChannel()
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
		case message, mc := <-hassChannel:
			if mc {
				//log.Info(message)
				switch m := message.(type) {
				case c.HassEntity:
					if m.Old.State != "" {
						// We do this in own go-routine so we never block main thread
						go a.handleEntity(&m)
					}

				case c.HassCallServiceEvent:
					go a.handleCallServiceEvent(&m)
				default:
					log.Errorf("Unexpected message type: %v", message)
				}

			} else {
				//
				log.Error("We should never get here!")
			}
		case <-a.cancelContext.Done():
			return
		}
	}
}

var defaultTimeoutForFullChannel = 5

func (a *ApplicationDaemon) handleCallServiceEvent(callServiceEvent *c.HassCallServiceEvent) {
	domainServiceCallListeners, exists := a.callServiceEventListeners[callServiceEvent.Domain]
	if !exists {
		return
	}

	// Check listen to status changes
	csl, exists := domainServiceCallListeners[callServiceEvent.Service]
	if exists {
		for _, callServiceEventChannel := range csl {
			select {
			case callServiceEventChannel <- *callServiceEvent:
			case <-time.After(time.Second * time.Duration(defaultTimeoutForFullChannel)):
				// This should never happen incase the app does not read the messages
				log.Errorf("Channel full, please check recevicer channel: %s", callServiceEvent.Service)
			case <-a.cancelContext.Done():
				// Exit cause of exit to os
				return
			}
		}
	}
}
func (a *ApplicationDaemon) handleEntity(entity *c.HassEntity) {
	// Check listen to status changes
	sl, exists := a.stateListeners[entity.ID]
	if exists {
		for _, chEntity := range sl {
			select {
			case chEntity <- *entity:
			case <-time.After(time.Second * time.Duration(defaultTimeoutForFullChannel)):
				// This should never happen incase the app does not read the messages
				log.Errorf("Channel full, please check recevicer channel: %s", entity.ID)
			case <-a.cancelContext.Done():
				// Exit cause of exit to os
				return
			}
		}
	}
	// Also check for plattform entitites
	platform := strings.Split(entity.ID, ".")[0]

	pl, exists := a.stateListeners[platform]
	if exists {
		for _, chPlatform := range pl {
			select {
			case chPlatform <- *entity:
			case <-time.After(time.Second * time.Duration(defaultTimeoutForFullChannel)):
				// This should never happen incase the app does not read the messages
				log.Errorf("Platform channel full, please check recevicer channel: %s", platform)
			case <-a.cancelContext.Done():
				// Exit cause of exit to os
				return
			}
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
	log.Debugln("Loading applications...")
	if len(a.applications) > 0 {
		a.unloadDaemonApplications()
	}

	a.applications = a.instanceAllApplications()
}
func (a *ApplicationDaemon) unloadDaemonApplications() {
	log.Debugln("Unloading applications...")
	// Remove all subscriptions here
	a.stateListeners = make(map[string][]chan client.HassEntity)
	// Remove the applications
	if len(a.applications) > 0 {
		for _, app := range a.applications {
			app.Cancel()
		}
		// Get new instance of empty list
		a.applications = []d.DaemonApplication{}
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
		log.Error("Failed to get all the configuration files.", err)
	}

	return fileList
}

func (a *ApplicationDaemon) getConfigFromFile(path string) (map[string]d.DeamonAppConfig, bool) {
	i := make(map[string]d.DeamonAppConfig, 1)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("Failed to read app yaml file", path, err)
		return nil, false
	}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		log.Error("Failed to parse app yaml file", path, err)
		return nil, false
	}
	return i, true
}

func (a *ApplicationDaemon) NewEntity(id string, daemonHelper d.DaemonAppHelper, autoRespondServiceCall bool,
	changedEntityChannel chan d.DaemonEntity) d.DaemonEntity {
	return NewEntity(id, daemonHelper, autoRespondServiceCall, changedEntityChannel)
}

func (a *ApplicationDaemon) instanceAllApplications() []d.DaemonApplication {
	applicationInstances := []d.DaemonApplication{}

	allApplicationConfigs := a.getAllApplicationConfigFilePaths()

	// Add the standard applications
	if len(a.config.People) > 0 {
		var peopleApp d.DaemonApplication = &defaultapps.PeopleApp{}
		applicationInstances = append(applicationInstances, peopleApp)
		log.Infoln("Loading default application: people_app")
		peopleApp.Initialize(a, d.DeamonAppConfig{})
	}

	for _, configFile := range allApplicationConfigs {
		cfgList, ok := a.getConfigFromFile(configFile)
		if ok {
			for _, appCfg := range cfgList {
				app, ok := a.NewDaemonApp(appCfg.App)
				if ok {
					log.Infoln("Loading application: ", appCfg.App)
					applicationInstances = append(applicationInstances, app)
					app.Initialize(a, appCfg)
				} else {
					log.Errorf("Did not find the application {%s}, please check config in [%s] ", appCfg.App, configFile)
				}
			}
		}

	}
	return applicationInstances
}

func init() {

	log = logrus.WithField("prefix", "Daemon")

}
