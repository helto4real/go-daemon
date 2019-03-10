package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	d "github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-daemon/daemon/config"
	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
	"github.com/sirupsen/logrus"
)

func TestGetAllApplicationConfigFilePaths(t *testing.T) {
	daemon := ApplicationDaemon{
		configPath: "testdata/ok"}

	files := daemon.getAllApplicationConfigFilePaths()
	h.Equals(t, 3, len(files))
	h.Equals(t, filepath.Join("testdata/ok", "app", "app.yaml"), files[0])
	h.Equals(t, filepath.Join("testdata/ok", "app", "folder", "myapp.yaml"), files[1])
	h.Equals(t, filepath.Join("testdata/ok", "app", "folder2", "myapp2.yaml"), files[2])
}

func TestGetConfigFromFile(t *testing.T) {

	daemon := ApplicationDaemon{
		configPath: "testdata/ok"}

	files := daemon.getAllApplicationConfigFilePaths()
	h.Equals(t, 3, len(files))

	c, ok := daemon.getConfigFromFile(files[1])
	h.Equals(t, true, ok)
	h.Equals(t, 1, len(c))

	instance, ok := c["testapp_instance"]
	h.Equals(t, true, ok)
	h.NotEquals(t, nil, instance)
	h.Equals(t, "switch.switch1", instance.Properties["theswitch"])
	h.Equals(t, "light.light1", instance.Properties["thelight"])
}

func TestGetInstance(t *testing.T) {
	daemon := ApplicationDaemon{
		availableApps: map[string]interface{}{
			"testapp": testapp{}}}

	app, ok := daemon.NewDaemonApp("testapp")
	h.Equals(t, true, ok)
	h.NotEquals(t, nil, app)
}

func TestGetInstanceTypeMissing(t *testing.T) {
	daemon := ApplicationDaemon{
		availableApps: map[string]interface{}{
			"testapp": testapp{}}}

	app, ok := daemon.NewDaemonApp("name_not_exist")
	h.Equals(t, false, ok)
	h.Equals(t, nil, app)
}

func TestHandleEntity(t *testing.T) {
	entity := client.HassEntity{
		ID:   "light.testentity",
		Name: "Hello"}

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 2),
				make(chan client.HassEntity, 2)}},
		cancelContext: context.Background()}

	daemon.handleEntity(&entity)

	e := <-daemon.stateListeners["light.testentity"][0]
	e2 := <-daemon.stateListeners["light.testentity"][1]
	h.NotEquals(t, nil, e)
	h.NotEquals(t, nil, e2)
}

func TestHandleEntityFullChannel(t *testing.T) {
	oldTimeout := defaultTimeoutForFullChannel
	defaultTimeoutForFullChannel = 1 // 1 second for test
	defer func() { defaultTimeoutForFullChannel = oldTimeout }()

	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
	}()

	entity := client.HassEntity{
		ID:   "light.testentity",
		Name: "Hello"}

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}},
		cancelContext: context.Background()}

	daemon.handleEntity(&entity)
	daemon.handleEntity(&entity)

	<-daemon.stateListeners["light.testentity"][0]
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "Channel full, please check recevicer channel"))

}

func TestCheckHassioOptionsConfig(t *testing.T) {
	oldOptionsPath := optionsPath
	defer func() { optionsPath = oldOptionsPath }()
	optionsPath = "testdata/options/options.json"
	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}},
		cancelContext: context.Background(),
		config:        &config.Config{}}
	daemon.checkHassioOptionsConfig()

	h.Equals(t, len(daemon.config.People), 2)
	h.Equals(t, logrus.GetLevel(), logrus.InfoLevel)
}

func TestCheckHassioOptionsConfigLogLevels(t *testing.T) {
	oldOptionsPath := optionsPath
	defer func() { optionsPath = oldOptionsPath }()
	oldLogLevel := logrus.GetLevel()
	defer logrus.SetLevel(oldLogLevel)

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}},
		cancelContext: context.Background(),
		config:        &config.Config{}}

	optionsPath = "testdata/options/options-debug.json"
	daemon.checkHassioOptionsConfig()
	h.Equals(t, logrus.GetLevel(), logrus.DebugLevel)

	optionsPath = "testdata/options/options-trace.json"
	daemon.checkHassioOptionsConfig()
	h.Equals(t, logrus.GetLevel(), logrus.TraceLevel)

	optionsPath = "testdata/options/options-warning.json"
	daemon.checkHassioOptionsConfig()
	h.Equals(t, logrus.GetLevel(), logrus.WarnLevel)

	optionsPath = "testdata/options/options-error.json"
	daemon.checkHassioOptionsConfig()
	h.Equals(t, logrus.GetLevel(), logrus.ErrorLevel)

	optionsPath = "testdata/options/options-fatal.json"
	daemon.checkHassioOptionsConfig()
	h.Equals(t, logrus.GetLevel(), logrus.FatalLevel)
}

func TestCheckHassioOptionsDefaultValues(t *testing.T) {
	oldOptionsPath := optionsPath
	defer func() { optionsPath = oldOptionsPath }()

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}},
		cancelContext: context.Background(),
		config:        &config.Config{}}

	optionsPath = "testdata/options/options-default-values.json"
	daemon.setDefaultSettings()
	daemon.checkHassioOptionsConfig()

	h.Equals(t, 300, daemon.config.Settings.TrackingSettings.JustArrivedTime)
	h.Equals(t, 60, daemon.config.Settings.TrackingSettings.JustLeftTime)
	h.Equals(t, "Home", daemon.config.Settings.TrackingSettings.HomeState)
	h.Equals(t, "Away", daemon.config.Settings.TrackingSettings.AwayState)
	h.Equals(t, "Just arrived", daemon.config.Settings.TrackingSettings.JustArrivedState)
	h.Equals(t, "Just left", daemon.config.Settings.TrackingSettings.JustLeftState)
}

func TestCheckHassioOptionsTracking(t *testing.T) {
	oldOptionsPath := optionsPath
	defer func() { optionsPath = oldOptionsPath }()

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"light.testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}},
		cancelContext: context.Background(),
		config:        &config.Config{}}

	optionsPath = "testdata/options/options-tracking.json"
	daemon.setDefaultSettings()
	daemon.checkHassioOptionsConfig()

	h.Equals(t, 301, daemon.config.Settings.TrackingSettings.JustArrivedTime)
	h.Equals(t, 61, daemon.config.Settings.TrackingSettings.JustLeftTime)
	h.Equals(t, "home", daemon.config.Settings.TrackingSettings.HomeState)
	h.Equals(t, "away", daemon.config.Settings.TrackingSettings.AwayState)
	h.Equals(t, "just_arrived", daemon.config.Settings.TrackingSettings.JustArrivedState)
	h.Equals(t, "just_left", daemon.config.Settings.TrackingSettings.JustLeftState)
}
func TestListenCallServiceEvent(t *testing.T) {
	daemon := NewApplicationDaemon()
	callServiceEvent := make(chan client.HassCallServiceEvent, 2)
	daemon.ListenCallServiceEvent("domain1", "service1", callServiceEvent)

	_, ok := daemon.callServiceEventListeners["domain1"]
	h.Equals(t, ok, true)
	_, ok = daemon.callServiceEventListeners["domain1"]["service1"]
	h.Equals(t, ok, true)
}

func TestListenCallServiceEventAlreadyRegistered(t *testing.T) {
	daemon := NewApplicationDaemon()
	callServiceEvent := make(chan client.HassCallServiceEvent, 2)
	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	level := logrus.GetLevel()
	logrus.SetLevel(logrus.DebugLevel)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(level)
	}()
	daemon.ListenCallServiceEvent("domain1", "service1", callServiceEvent)
	daemon.ListenCallServiceEvent("domain1", "service1", callServiceEvent)

	s := mockStdErr.String()
	log.Print(s)
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "ListenCallServiceEvent: Already registered on"))

}

type testapp struct {
}

func (a testapp) Initialize(helper d.DaemonAppHelper, config d.DeamonAppConfig) bool {
	return true
}

func (a testapp) Cancel() {

}
