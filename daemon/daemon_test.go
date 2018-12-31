package daemon_test

import (
	"testing"

	h "github.com/helto4real/go-daemon/daemon/test"
	"github.com/helto4real/go-hassclient/client"
)

func TestGetAllApplicationConfigFilePaths(t *testing.T) {
	h.Equals(t, true, true)
	x := client.HassEntity{}
	h.NotEquals(t, nil, x)
}
