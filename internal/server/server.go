package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

const defPort = 4789

type S2sConfig struct {
	Port int
}

func NewS2sServer(cfg S2sConfig) *s2sServer {
	if cfg.Port == 0 {
		cfg.Port = defPort
	}
	done := make(chan interface{}, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandlerFunc)

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: mux,
	}

	srv := s2sServer{
		Port:  cfg.Port,
		sever: httpServer,
		done:  done,
	}
	return &srv
}

type s2sServer struct {
	Port    int
	sever   *http.Server
	running int32
	done    chan interface{}
}

// returns true if the server is currently running
func (srv *s2sServer) IsRunning() bool {
	if atomic.LoadInt32(&srv.running) == 0 {
		return false
	} else {
		return true
	}
}

func (srv *s2sServer) Start() {
	if atomic.CompareAndSwapInt32(&srv.running, 0, 1) {
		if err := srv.sever.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}
}

// Start the Server in a non blocking way in a separate routine
func (srv *s2sServer) StartBackground() {
	if atomic.LoadInt32(&srv.running) == 0 {
		go func() {
			srv.Start()
		}()
	}
}

func (srv *s2sServer) Stop() {
	if atomic.LoadInt32(&srv.running) != 0 {
		srv.sever.Shutdown(context.Background())
		srv.sever = nil
	}
}

func mainHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404")
	}
}
