package send2slack_test

import (
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"send2slack/internal/send2slack"
	"testing"
	"time"
)

func TestStartAndStopDirWatcher(t *testing.T) {

	// don't print log messages during tests
	//logrus.SetLevel(logrus.FatalLevel)

	dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(dir)
	spew.Dump(dir)

	// start the server
	cfg := send2slack.Config{
		WatchDir: dir,
	}

	dw, err := send2slack.NewDirWatcher(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	dummySender := send2slack.DummyMessageSender{}
	dummySender.Msg = "TestStartAndStopDirWatcher"

	// set the sender to use dummy
	dw.MsgSender = &dummySender

	dw.StartBackground()
	// wait for watcher to start
	time.Sleep(100 * time.Microsecond)

	t.Run("test watcher is running", func(t *testing.T) {
		// check running state
		if dw.IsRunning() != true {
			t.Error("expected serv to be running, server is not running")
		}

		s := []byte("hello\ngo\n")
		err := ioutil.WriteFile(dir+"/file1", s, 0644)
		if err != nil {
			t.Fatal(err)
		}

		t.Error(dummySender.Msg)

	})

	dw.Stop()
	// wait for watcher to stop
	time.Sleep(100 * time.Microsecond)

	t.Run("test watcher is stopped", func(t *testing.T) {
		if dw.IsRunning() != true {
			t.Error("expected serv to be stopped, server is running")
		}

	})
}
