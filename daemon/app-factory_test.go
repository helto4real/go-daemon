package daemon_test

import (
	"testing"

	"github.com/helto4real/go-daemon/daemon"
	h "github.com/helto4real/go-daemon/daemon/test"
)

func TestGetInstance(t *testing.T) {

	app, ok := daemon.NewDaemonApp("testapp")
	h.Equals(t, true, ok)
	h.NotEquals(t, nil, app)
}

func TestGetInstanceTypeMissing(t *testing.T) {

	app, ok := daemon.NewDaemonApp("testapp_not_found")
	h.Equals(t, false, ok)
	h.Equals(t, nil, app)
}

type testapp struct {
}

func (a testapp) Initialize(helper daemon.DaemonAppHelper) bool {
	return true
}

func (a testapp) NewInstance() daemon.DaemonApplication {
	return testapp{}
}

func init() {
	daemon.RegisterDaemonApp("testapp", &testapp{})
}
