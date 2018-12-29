package config_test

import (
	"errors"
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
