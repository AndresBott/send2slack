package send2slack

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"send2slack/internal/mbox"
	"sync/atomic"
)

type DirWatcher struct {
	path      string
	MsgSender MessageSender
	watcher   *fsnotify.Watcher
	running   int32
	done      chan interface{}
}

func NewDirWatcher(cfg *Config) (*DirWatcher, error) {

	if cfg.WatchDir == "" {
		return nil, fmt.Errorf("watching dir cannot be empty")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	dw := DirWatcher{
		watcher: watcher,
		path:    cfg.WatchDir,
	}
	return &dw, nil

}

// returns true if the server is currently running
func (dw *DirWatcher) IsRunning() bool {
	if atomic.LoadInt32(&dw.running) == 0 {
		return false
	} else {
		return true
	}
}

func (dw *DirWatcher) Start() {
	if atomic.CompareAndSwapInt32(&dw.running, 0, 1) {

		go func() {
			for {
				select {
				case event, ok := <-dw.watcher.Events:
					if !ok {
						return
					}
					//log.Println("event:", event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						//log.Println("modified file:", event.Name)
						err := dw.ConsumeMbox(event.Name)
						if err != nil {
							log.Error(err)
						}
					}

				case err, ok := <-dw.watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()
		err := dw.watcher.Add(dw.path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Start the Server in a non blocking way in a separate routine
func (dw *DirWatcher) StartBackground() {
	if atomic.LoadInt32(&dw.running) == 0 {
		go func() {
			dw.Start()
		}()
	}
}

// Stop the dir watcher
func (dw *DirWatcher) Stop() {
	if atomic.LoadInt32(&dw.running) != 0 {
		dw.watcher.Close()
		dw.watcher = nil
	}
}

// ConsumeMbox will consume all mbox emails in a file
//
func (dw *DirWatcher) ConsumeMbox(file string) error {

	hand, err := mbox.NewHandler(file)
	if err != nil {
		return err
	}

	for hand.HasMails() {

		mailBytes, err := hand.ConsumeLastMail()

		if err != nil {
			// if err try to send the error to slack
			errMsg := Message{
				Text:  "Error while reading mbox: " + err.Error(),
				Color: "red",
			}
			_ = dw.MsgSender.SendMessage(&errMsg)

			log.Error("error reading mbox:" + err.Error())
		}

		mail := mbox.NewMailFromBytes(mailBytes)
		msg, err := NewMessageFromMail(Email(*mail), "")

		err = dw.MsgSender.SendMessage(msg)
		if err != nil {
			return err
		}

	}

	return nil
}
