package config_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	c "github.com/helto4real/go-daemon/daemon/config"
	h "github.com/helto4real/go-daemon/daemon/test"
)

func TestOpen(t *testing.T) {
	configuration := c.NewConfiguration("testdata/go-daemon.yaml")
	config, _ := configuration.Open()
	h.Equals(t, config.HomeAssistant.IP, "192.168.0.100")
	h.Equals(t, config.HomeAssistant.SSL, false)
	h.Equals(t, config.HomeAssistant.Token, "ABCDEFG1234567")

	h.Equals(t, 2, len(config.People))
	h.Equals(t, 3, len(config.People["fred"].Devices))

	h.Equals(t, (*c.SettingsConfig)(nil), config.Settings)
}

func TestOpenWithSettings(t *testing.T) {
	configuration := c.NewConfiguration("testdata/go-daemon-settings.yaml")
	config, _ := configuration.Open()
	h.Equals(t, "192.168.0.100", config.HomeAssistant.IP)
	h.Equals(t, false, config.HomeAssistant.SSL)
	h.Equals(t, "ABCDEFG1234567", config.HomeAssistant.Token)

	h.Equals(t, 2, len(config.People))
	h.Equals(t, 3, len(config.People["fred"].Devices))

	h.NotEquals(t, nil, config.Settings)
	h.NotEquals(t, nil, config.Settings.TrackingSettings)
	h.Equals(t, 300, config.Settings.TrackingSettings.JustArrivedTime)
	h.Equals(t, 60, config.Settings.TrackingSettings.JustLeftTime)
	h.Equals(t, "home", config.Settings.TrackingSettings.HomeState)
	h.Equals(t, "away", config.Settings.TrackingSettings.AwayState)
	h.Equals(t, "just_arrived", config.Settings.TrackingSettings.JustArrivedState)
	h.Equals(t, "just_left", config.Settings.TrackingSettings.JustLeftState)
}

func TestFailOpenConfigFile(t *testing.T) {
	configuration := c.NewConfiguration("nofileexists")
	config, err := configuration.Open()
	h.Assert(t, config == nil, "Configuration should return nil value")
	h.Assert(t, err != nil, "Configuration should return error value")
}
func TestNewConfiguration(t *testing.T) {
	h.Assert(t, c.NewConfiguration("testdata/go-daemon.yaml") != nil, "Configuration failed")
}

type failReader struct{}

func (a failReader) Read(read []byte) (int, error) {
	return 0, errors.New("Fake error")
}
func TestOpenReaderFails(t *testing.T) {
	config := c.NewConfiguration("testdata/go-daemon.yaml")
	_, err := config.OpenReader(failReader{})

	h.Assert(t, err != nil, "Expected error!")

}

func TestHassioOptionsConfig(t *testing.T) {
	options_json, _ := ioutil.ReadFile("testdata/options.json")
	options := &c.HassioOptionsConfig{}
	err := json.Unmarshal(options_json, options)
	h.Assert(t, err == nil, "Error parsing options")
	h.NotEquals(t, nil, options)
	h.NotEquals(t, nil, options.Persons)
	h.NotEquals(t, nil, options.Tracking)

	h.Equals(t, 2, len(options.Persons))
	h.Equals(t, 3, len(options.Persons[0].Devices))

	h.Equals(t, 300, options.Tracking.JustArrivedTime)
	h.Equals(t, 60, options.Tracking.JustLeftTime)
	h.Equals(t, "Home", options.Tracking.HomeState)
	h.Equals(t, "Away", options.Tracking.AwayState)
	h.Equals(t, "Just arrived", options.Tracking.JustArrivedState)
	h.Equals(t, "Just left", options.Tracking.JustLeftState)
}
