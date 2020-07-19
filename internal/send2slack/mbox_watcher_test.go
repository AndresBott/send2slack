package send2slack_test

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"send2slack/internal/send2slack"
	"testing"
	"time"
)

func TestStartAndStopDirWatcher(t *testing.T) {

	// don't print log messages during tests
	logrus.SetLevel(logrus.WarnLevel)

	dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// start the server
	cfg := send2slack.Config{
		WatchDir: dir,
	}

	dw, err := send2slack.NewDirWatcher(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	dw.StartBackground()
	// wait for watcher to start
	time.Sleep(20 * time.Millisecond)

	t.Run("test watcher is running ", func(t *testing.T) {
		// check running state
		if dw.IsRunning() != true {
			t.Error("expected serv to be running, server is not running")
		}
	})

	dw.Stop()
	// wait for watcher to stop
	time.Sleep(20 * time.Microsecond)

	t.Run("test watcher is stopped", func(t *testing.T) {
		if dw.IsRunning() != true {
			t.Error("expected serv to be stopped, server is running")
		}
	})
}

func TestDirWatcher_ConsumeMbox(t *testing.T) {
	// don't print log messages during tests
	logrus.SetLevel(logrus.DebugLevel)

	dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	//spew.Dump(dir)

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
	time.Sleep(20 * time.Millisecond)

	t.Run("test watcher is running ", func(t *testing.T) {

		file := dir + "/file1"

		err := writeMailToMbox(file, "started")
		if err != nil {
			t.Fatal(err)
		}
		// let the watcher consume the events
		//time.Sleep(1 * time.Millisecond)

		err = writeMailToMbox(dir+"/file1", "msg2")
		if err != nil {
			t.Fatal(err)
		}

		// let the watcher consume the events
		time.Sleep(20 * time.Millisecond)

		expected := "TestStartAndStopDirWatcher|msg2|started"
		if dummySender.Msg != expected {
			t.Errorf("consumed message does not match expected, got \"%s\" expected: \"%s\"", dummySender.Msg, expected)
		}

		// test file size
		fi, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Size() != 0 {
			t.Errorf("expected file size after consumption no 0, got %d bytes", fi.Size())
		}
	})
}

func TestDirWatcher_ConsumeMboxDir(t *testing.T) {
	// don't print log messages during tests
	logrus.SetLevel(logrus.DebugLevel)

	t.Run("empty mbox dir", func(t *testing.T) {

		dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		//spew.Dump(dir)

		cfg := send2slack.Config{
			WatchDir: dir,
		}
		dw, err := send2slack.NewDirWatcher(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		dummySender := send2slack.DummyMessageSender{}
		dummySender.Msg = "ConsumeMboxDir"

		// set the sender to use dummy
		dw.MsgSender = &dummySender

		dw.ConsumeMboxDir()

		expected := "ConsumeMboxDir"
		if dummySender.Msg != expected {
			t.Errorf("consumed message does not match expected, got \"%s\" expected: \"%s\"", dummySender.Msg, expected)
		}

	})

	t.Run("mbox dir with empty files", func(t *testing.T) {

		dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		//spew.Dump(dir)

		cfg := send2slack.Config{
			WatchDir: dir,
		}
		dw, err := send2slack.NewDirWatcher(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		dummySender := send2slack.DummyMessageSender{}
		dummySender.Msg = "ConsumeMboxDir"

		// set the sender to use dummy
		dw.MsgSender = &dummySender

		// create the empty mbox file
		emptyFile, err := os.Create(dir + "/empty.mbox")
		if err != nil {
			t.Fatal(err)
		}
		emptyFile.Close()

		dw.ConsumeMboxDir()

		expected := "ConsumeMboxDir"
		if dummySender.Msg != expected {
			t.Errorf("consumed message does not match expected, got \"%s\" expected: \"%s\"", dummySender.Msg, expected)
		}
	})

	t.Run("consume existing mbox", func(t *testing.T) {

		dir, err := ioutil.TempDir("/tmp", "s2s_watcher")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		//spew.Dump(dir)

		cfg := send2slack.Config{
			WatchDir: dir,
		}
		dw, err := send2slack.NewDirWatcher(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		dummySender := send2slack.DummyMessageSender{}
		dummySender.Msg = "ConsumeMboxDir"

		// set the sender to use dummy
		dw.MsgSender = &dummySender

		writeMailToMbox(dir+"/file1", "msg1")
		writeMailToMbox(dir+"/file1", "msg2")

		dw.ConsumeMboxDir()
		writeMailToMbox(dir+"/file2", "msg3")
		dw.ConsumeMboxDir()

		expected := "ConsumeMboxDir|msg2|msg1|msg3"
		if dummySender.Msg != expected {
			t.Errorf("consumed message does not match expected, got \"%s\" expected: \"%s\"", dummySender.Msg, expected)
		}
	})
}

func writeMailToMbox(file string, body string) error {

	m :=
		`From www-data@amelia.com  Thu Dec 21 05:00:01 2017
From: root@amelia.com (Cron Daemon)
To: www-data@amelia.com

`
	m = m + body + "\n\n"

	f, err := os.OpenFile(file,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(m)); err != nil {

		return err
	}

	return nil
}
