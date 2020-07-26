package daemon

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"send2slack/internal/config"
	"send2slack/internal/mbox"
	"send2slack/internal/sender"
	"sync/atomic"
	"time"
)

type item struct {
	name string
	val  *int32
}
type itemList struct {
	items []*item
}

func newItemList() *itemList {
	return &itemList{
		items: []*item{},
	}
}

// flags the file in the list as active
// this is an atomic operation
func (i *itemList) active(in string) bool {

	var sel *item
	var val *int32
	for j := range i.items {
		if i.items[j].name == in {
			sel = i.items[j]
			val = sel.val
		}
	}

	if sel == nil {
		zero := int32(0)
		i.items = append(i.items, &item{
			name: in,
			val:  &zero,
		})
		val = &zero
	}
	if atomic.CompareAndSwapInt32(val, 0, 1) {
		return true
	}
	return false
}

func (i *itemList) disable(in string) bool {
	var sel *item
	var val *int32

	for j := range i.items {
		if i.items[j].name == in {
			sel = i.items[j]
			val = sel.val
		}
	}

	if sel != nil {
		if atomic.CompareAndSwapInt32(val, 1, 0) {
			return true
		}
	}

	return false
}

type DirWatcher struct {
	path           string
	MsgSender      sender.MessageSender
	watcher        *fsnotify.Watcher
	running        int32
	filesConsuming *itemList
	throttling     int
}

func NewDirWatcher(cfg *config.DaemonConfig) (*DirWatcher, error) {

	if cfg.WatchDir == "" {
		return nil, fmt.Errorf("watching dir cannot be empty")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	senderCfg := &config.ClientConfig{
		Token:      cfg.Token,
		IsDefault:  cfg.IsDefault,
		DefChannel: cfg.SendmailChannel,
		Mode:       config.ModeDirectCli,
	}

	sndr, err := sender.NewSlackSender(senderCfg)
	if err != nil {
		return nil, err
	}

	dw := DirWatcher{
		watcher:        watcher,
		path:           cfg.WatchDir,
		filesConsuming: newItemList(),
		MsgSender:      sndr,
		throttling:     cfg.MailThrottling,
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
		log.Info("Starting mbox watcher on path:" + dw.path)

		// consume any messages present when starting the watcher
		dw.ConsumeMboxDir()

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
						dw.ConsumeMbox(event.Name, false)
						time.Sleep(10 * time.Microsecond)
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

func (dw *DirWatcher) ConsumeMboxDir() {

	files := map[string]os.FileInfo{}
	err := filepath.Walk(dw.path, func(path string, info os.FileInfo, err error) error {
		files[path] = info
		return nil
	})
	if err != nil {
		panic(err)
	}

	for path, info := range files {
		if !info.IsDir() {
			log.Debug("Consuming mails in file: " + path)
			dw.ConsumeMbox(path, true)
		}
	}
}

// ConsumeMbox will consume all mbox emails in a file
//
func (dw *DirWatcher) ConsumeMbox(file string, blockExec bool) {

	fi, err := os.Stat(file)
	if err != nil {
		log.Error(err)
		return
	}

	if fi.Size() == 0 {
		return
	}

	if dw.filesConsuming.active(file) { // atomic operation
		done := make(chan interface{}, 1)

		go func(file string) {

			hand, err := mbox.New(file)
			if err != nil {
				log.Error(err)
			}

			for hand.HasMails() {
				mailBytes, err := hand.ReadLastMail(true)

				if err != nil {
					nErr := errors.New("Error while reading mbox: " + err.Error())
					// if err try to send the error to slack
					dw.MsgSender.SendError(nErr)
					log.Error("error reading mbox:" + err.Error())
					continue
				}

				mail := mbox.NewMailFromBytes(mailBytes)
				if mail.Body == "" && len(mail.Headers) == 0 {
					// avoid reading double mail
					continue
				}

				msg, err := sender.NewMessageFromMail(sender.Email(*mail))

				if err != nil {
					dw.MsgSender.SendError(err)
					log.Error(err)
				}

				err = dw.MsgSender.SendMessage(msg)
				if err != nil {
					log.Error(err)
				}
				// throttle email submissions
				time.Sleep(time.Duration(dw.throttling) * time.Millisecond)
			}

			log.Info("finished processing file: " + file)
			dw.filesConsuming.disable(file)
			done <- true
		}(file)

		if blockExec {
			<-done
		}
	}
}

func here() {
	fmt.Println("=>> HERE")
}
