package daemon

import (
	"path/filepath"
	"testing"

	h "github.com/helto4real/go-daemon/daemon/test"
)

func TestGetAllApplicationConfigFilePaths(t *testing.T) {
	daemon := ApplicationDaemon{
		configPath: "testdata"}

	files := daemon.getAllApplicationConfigFilePaths()
	h.Equals(t, 3, len(files))
	h.Equals(t, filepath.Join("testdata", "app", "app.yaml"), files[0])
	h.Equals(t, filepath.Join("testdata", "app", "folder", "myapp.yaml"), files[1])
	h.Equals(t, filepath.Join("testdata", "app", "folder2", "myapp2.yaml"), files[2])
}

func TestGetConfigFromFile(t *testing.T) {

	daemon := ApplicationDaemon{
		configPath: "testdata"}

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

type testapp struct {
}

func (a testapp) Initialize(helper DaemonAppHelper, config DeamonAppConfig) bool {
	return true
}

func (a testapp) Cancel() {

}
