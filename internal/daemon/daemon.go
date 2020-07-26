package daemon

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"send2slack/internal/config"
	"sync/atomic"
)

type daemon struct {
	cfg     *config.DaemonConfig
	done    chan interface{}
	running int32
}

func NewDaemon(cfg *config.DaemonConfig) (*daemon, error) {

	if cfg.ListenUrl == "false" && cfg.WatchDir == "false" {
		return nil, errors.New("both mbox-watch and server have been disabled")
	}

	d := daemon{
		cfg:  cfg,
		done: make(chan interface{}, 1),
	}

	return &d, nil
}

// returns true if the server is currently running
func (d *daemon) IsRunning() bool {
	if atomic.LoadInt32(&d.running) == 0 {
		return false
	} else {
		return true
	}
}

func (d *daemon) Start() {

	if atomic.CompareAndSwapInt32(&d.running, 0, 1) {

		if d.cfg.ListenUrl != "false" {
			server, err := NewServer(d.cfg)
			if err != nil {
				log.Fatal(err)
			}
			server.StartBackground()
		}

		if d.cfg.WatchDir != "false" {
			watcher, err := NewDirWatcher(d.cfg)

			if err != nil {
				log.Fatal(err)
			}
			watcher.StartBackground()
		}

		<-d.done
	}
}

// Start the Server in a non blocking way in a separate routine
func (d *daemon) StartBackground() {
	if atomic.LoadInt32(&d.running) == 0 {
		go func() {
			d.Start()
		}()
	}
}

func (d *daemon) Stop() {
	if atomic.LoadInt32(&d.running) != 0 {
		d.done <- true
		close(d.done)
	}
}
