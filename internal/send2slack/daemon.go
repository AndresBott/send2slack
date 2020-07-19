package send2slack

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
)

type daemon struct {
	cfg     *Config
	done    chan interface{}
	running int32
}

func NewDaemon(cfg *Config) (*daemon, error) {

	if cfg.ListenUrl == "false" {
		if cfg.WatchDir == "false" {
			cfg.Mode = ModeNoServerNoWatch
		} else {
			cfg.Mode = ModeNoServerWatch
		}
	} else {
		cfg.Mode = ModeServerNoWatch
		if cfg.WatchDir != "false" {
			cfg.Mode = ModeServerWatch
		}
	}

	if cfg.Mode == ModeNoServerNoWatch {
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

		if d.cfg.Mode == ModeServerWatch || d.cfg.Mode == ModeServerNoWatch {
			server, err := NewServer(d.cfg)
			if err != nil {
				log.Fatal(err)
			}
			server.StartBackground()
		}

		if d.cfg.Mode == ModeServerWatch || d.cfg.Mode == ModeNoServerWatch {
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
