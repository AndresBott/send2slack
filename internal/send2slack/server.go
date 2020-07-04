package send2slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

const defPort = 4789

type ServerConfig struct {
	Port       int
	SlackToken string
}

type Server struct {
	port        int
	sever       *http.Server
	slackSender *Sender
	running     int32
	done        chan interface{}
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.Port == 0 {
		cfg.Port = defPort
	}
	done := make(chan interface{}, 1)

	srv := Server{
		port:        cfg.Port,
		done:        done,
		slackSender: NewSender(cfg.SlackToken),
	}

	httpServer := &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.mainHandlerFunc)
	httpServer.Handler = mux

	srv.sever = httpServer
	return &srv
}

// returns true if the server is currently running
func (srv *Server) IsRunning() bool {
	if atomic.LoadInt32(&srv.running) == 0 {
		return false
	} else {
		return true
	}
}

func (srv *Server) Start() {
	if atomic.CompareAndSwapInt32(&srv.running, 0, 1) {
		if err := srv.sever.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}
}

// Start the Server in a non blocking way in a separate routine
func (srv *Server) StartBackground() {
	if atomic.LoadInt32(&srv.running) == 0 {
		go func() {
			srv.Start()
		}()
	}
}

func (srv *Server) Stop() {
	if atomic.LoadInt32(&srv.running) != 0 {
		srv.sever.Shutdown(context.Background())
		srv.sever = nil
	}
}

func (srv *Server) mainHandlerFunc(w http.ResponseWriter, r *http.Request) {

	// discard any request that is not a POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404")
		return
	}

	ContentType := r.Header.Get("Content-Type")
	if ContentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400: content type is not \"application/json\"")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400: error decoding json body")
		return
	}

	err = msg.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400: error validating message")
		return
	}

	// return ok without sending the message if message is debug
	if msg.Debug {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "ok")
		return
	}

	sender := srv.slackSender

	err = sender.SendMessage(&msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500: unable to send slack message")
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "ok")
	return

}
