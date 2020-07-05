package send2slack

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type Server struct {
	listen      string
	sever       *http.Server
	slackSender *Sender
	running     int32
	done        chan interface{}
}

func NewServer(cfg *Config) (*Server, error) {

	host, port, err := ParseListenAddress(cfg.ListenUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to validate listen url: %v", err)
	}

	if port == 0 {
		port = DefaultPort
	}

	done := make(chan interface{}, 1)

	cfg.Mode = ModeDirectCli
	sender, err := NewSender(cfg)
	if err != nil {
		return nil, err
	}

	srv := Server{
		listen:      host + ":" + strconv.Itoa(port),
		done:        done,
		slackSender: sender,
	}

	httpServer := &http.Server{
		Addr: srv.listen,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.mainHandlerFunc)
	httpServer.Handler = mux

	srv.sever = httpServer
	return &srv, nil
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
		log.Infof("[%s] responded with 404", r.Method)
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
		log.Warnf("unable to send message: %v", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "ok")
	log.Infof("message submitted to channel: #%s", msg.Channel)
	return

}

// ParseListenAddress takes a listen address in format <ip>:<port> or :<port> and validates correct values
// returns listen string if validated correctly
func ParseListenAddress(in string) (string, int, error) {

	s := strings.TrimSpace(in)

	if s == "" {
		return "", -1, fmt.Errorf("expecting listen pattern like \"<ip>:<port>\" or \":<port>\"")
	}

	var port int
	var host string
	var err error

	// if string starts with  " : " we assume something like :<port>
	if s[0:1] == ":" {

		PortStr := s[1:]
		port, err = strconv.Atoi(PortStr)
		if err != nil {
			return "", -1, fmt.Errorf("unable to parse port")
		}
	} else {
		// if not, we assume <ip>:<port>
		spl := strings.Split(in, ":")
		if len(spl) != 2 {
			return "", -1, fmt.Errorf("expecting listen pattern like \"<ip>:<port>\" or \":<port>\"")
		}
		host = spl[0]
		port, err = strconv.Atoi(spl[1])
		if err != nil {
			return "", -1, fmt.Errorf("unable to parse port")
		}
	}

	return host, port, nil
}
