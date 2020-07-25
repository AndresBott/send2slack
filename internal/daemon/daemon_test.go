package daemon_test

import (
	"github.com/phayes/freeport"
	"io/ioutil"
	"os"
	"send2slack/internal/config"
	"send2slack/internal/daemon"
	"strconv"
	"testing"
	"time"
)

func TestStartAndStopDaemon(t *testing.T) {

	// don't print log messages during tests
	//logrus.SetLevel(logrus.WarnLevel)

	tmpPath, _, err := prePateStage()
	if err != nil {
		t.Fatal(err)
	}

	dCfg, err := config.NewDaemonConfig(tmpPath + "/server.yaml")
	if err != nil {
		t.Fatal(err)
	}

	dmn, err := daemon.NewDaemon(dCfg)
	if err != nil {
		t.Fatal(err)
	}

	dmn.StartBackground()
	// wait for watcher to start
	time.Sleep(20 * time.Millisecond)

	t.Run("test watcher is running ", func(t *testing.T) {
		// check running state
		if dmn.IsRunning() != true {
			t.Error("expected serv to be running, server is not running")
		}
	})

	dmn.Stop()
	// wait for watcher to stop
	time.Sleep(20 * time.Microsecond)

	t.Run("test watcher is stopped", func(t *testing.T) {
		if dmn.IsRunning() != true {
			t.Error("expected serv to be stopped, server is running")
		}
	})
}

func prePateStage() (path string, port int, er error) {

	// generate a tmp dir
	tmpDir, err := ioutil.TempDir("/tmp", "s2s_daemon_")
	if err != nil {
		return "", -1, err
	}

	// get a free port
	newPort, err := freeport.GetFreePort()
	if err != nil {
		return "", -1, err
	}
	listenUrl := "localhost:" + strconv.Itoa(newPort)

	// create the config file
	cfgStr := `---
slack:
  token: "my_token"
  default_channel: "general"
  sendmail_channel: "general"
daemon:
  listen_url: "` + listenUrl + `"
  mbox_watch: "` + tmpDir + `/mbox"

`
	err = ioutil.WriteFile(tmpDir+"/server.yaml", []byte(cfgStr), 0644)
	if err != nil {
		return "", -1, err
	}

	// create the mbox
	err = os.MkdirAll(tmpDir+"/mbox", os.ModePerm)
	if err != nil {
		return "", -1, err
	}
	return tmpDir, newPort, nil
}
