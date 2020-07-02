package server_test

import (
	"github.com/phayes/freeport"
	"log"
	"net/http"
	"send2slack/internal/server"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartAndStopServer(t *testing.T) {

	// get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	// start the server
	cfg := server.S2sConfig{
		Port: port,
	}
	srv := server.NewS2sServer(cfg)
	srv.StartBackground()
	// wait for server to start
	time.Sleep(100 * time.Microsecond)

	t.Run("test server is running", func(t *testing.T) {
		// check running state
		if srv.IsRunning() != true {
			t.Error("expected serv to be running, server is not running")
		}

		resp, err := http.Get("http://localhost:" + strconv.Itoa(port))
		if err != nil {
			print(err)
		}

		expected := 404
		got := resp.StatusCode

		if expected != got {
			t.Errorf("wrong status code: got %d, expected %d", got, expected)
		}
	})

	srv.Stop()
	// wait for server to stop
	time.Sleep(100 * time.Microsecond)

	t.Run("test server is stopped", func(t *testing.T) {
		if srv.IsRunning() != true {
			t.Error("expected serv to be stopped, server is running")
		}

		_, err := http.Get("http://localhost:" + strconv.Itoa(port))

		expected := "connect: connection refused"
		got := err.Error()

		if !strings.Contains(got, expected) {
			t.Errorf("unexpected returned error: got \"%s\", expected \"%s\"", got, expected)
		}

	})

	//expected = 404
	//got = resp.StatusCode
	//
	//if expected != got{
	//	t.Errorf("wrong status code: got %d, expected %d",got,expected)
	//}

}
