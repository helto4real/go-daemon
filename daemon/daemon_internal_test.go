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
	h.Equals(t, 2, len(files))
	h.Equals(t, filepath.Join("testdata", "app", "app.yaml"), files[0])
	h.Equals(t, filepath.Join("testdata", "app", "folder", "myapp.yaml"), files[1])
}

func TestGetConfigFromFile(t *testing.T) {

	daemon := ApplicationDaemon{
		configPath: "testdata"}

	files := daemon.getAllApplicationConfigFilePaths()
	h.Equals(t, 2, len(files))

	c, ok := daemon.getConfigFromFile(files[1])
	h.Equals(t, true, ok)
	h.Equals(t, 1, len(c))

}
