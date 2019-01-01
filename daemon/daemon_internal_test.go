package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
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
		ID:   "testentity",
		Name: "Hello"}

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 2),
				make(chan client.HassEntity, 2)}}}

	daemon.handleEntity(&entity)

	e := <-daemon.stateListeners["testentity"][0]
	e2 := <-daemon.stateListeners["testentity"][1]
	h.NotEquals(t, nil, e)
	h.NotEquals(t, nil, e2)
}

func TestHandleEntityFullChannel(t *testing.T) {
	mockStdErr := strings.Builder{}
	logrus.SetOutput(&mockStdErr)
	defer func() {
		os.Stderr.WriteString(mockStdErr.String())
		logrus.SetOutput(os.Stderr)
	}()

	entity := client.HassEntity{
		ID:   "testentity",
		Name: "Hello"}

	daemon := ApplicationDaemon{
		stateListeners: map[string][]chan client.HassEntity{
			"testentity": []chan client.HassEntity{
				make(chan client.HassEntity, 1)}}}

	daemon.handleEntity(&entity)
	daemon.handleEntity(&entity)

	<-daemon.stateListeners["testentity"][0]
	h.Equals(t, true, strings.Contains(mockStdErr.String(), "Channel full for entity: testentity"))

}

type testapp struct {
}

func (a testapp) Initialize(helper DaemonAppHelper, config DeamonAppConfig) bool {
	return true
}

func (a testapp) Cancel() {

}
