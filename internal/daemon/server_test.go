package daemon_test

import (
	"bytes"
	"encoding/json"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"send2slack/internal/config"
	"send2slack/internal/daemon"
	"send2slack/internal/sender"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartAndStopServer(t *testing.T) {

	// don't print log messages during tests
	logrus.SetLevel(logrus.FatalLevel)

	// get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	// start the server
	cfg := config.DaemonConfig{
		ListenUrl: ":" + strconv.Itoa(port),
	}
	srv, err := daemon.NewServer(&cfg)
	if err != nil {
		t.Fatal(err)
	}
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
			t.Fatal(err)
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
}

type serverTc struct {
	name         string
	method       string
	contentType  string
	expectedCode int
	expectedBody string
	msg          sender.Message
}

func TestSeverMessages(t *testing.T) {
	// get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	// start the server
	cfg := config.DaemonConfig{
		ListenUrl: ":" + strconv.Itoa(port),
	}
	srv, err := daemon.NewServer(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv.StartBackground()
	// wait for server to start
	time.Sleep(200 * time.Microsecond)

	logrus.SetLevel(logrus.ErrorLevel)

	tcs := []serverTc{
		{
			name:         "Submit a message",
			method:       "POST",
			contentType:  "application/json",
			expectedCode: 202,
			msg: sender.Message{
				Debug: true,
				Text:  "sample",
			},
		},
		{
			name:         "invalid message",
			method:       "POST",
			contentType:  "application/json",
			expectedCode: 400,
			expectedBody: "400: error validating message",
			msg: sender.Message{
				Debug: true,
			},
		},
		{
			name:         "invalid content type",
			method:       "POST",
			contentType:  "invalid",
			expectedCode: 400,
			expectedBody: "400: content type is not \"application/json\"",
			msg: sender.Message{
				Debug: true,
			},
		},
		{
			name:         "misconfigured slack client",
			method:       "POST",
			contentType:  "application/json",
			expectedCode: 500,
			expectedBody: "500: unable to send slack message",
			msg: sender.Message{
				Text: "sample",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			jsonMsg, err := json.Marshal(tc.msg)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(tc.method, "http://localhost:"+strconv.Itoa(port), bytes.NewBuffer(jsonMsg))
			req.Header.Set("Content-Type", tc.contentType)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			expected := tc.expectedCode
			got := resp.StatusCode
			if expected != got {
				t.Errorf("wrong status code: got %d, expected %d", got, expected)
			}

			if tc.expectedBody != "" {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				gotBody := string(bodyBytes)
				expectedBody := tc.expectedBody

				if expectedBody != gotBody {
					t.Errorf("wrong status code: got \"%s\", expected \"%s\"", gotBody, expectedBody)
				}
			}
		})
	}

	srv.Stop()
	// wait for server to stop
	time.Sleep(100 * time.Microsecond)

}

type ParseListenAddressTc struct {
	name        string
	in          string
	expectedErr string
	expected    string
}

func TestParseListenAddress(t *testing.T) {
	tcs := []ParseListenAddressTc{
		{
			name:        "only port",
			in:          ":123",
			expectedErr: "",
			expected:    ":123",
		},
		{
			name:        "ip and port port",
			in:          "127.0.0.1:123",
			expectedErr: "",
			expected:    "127.0.0.1:123",
		},
		{
			name:        "wrong port",
			in:          ":bla",
			expectedErr: "unable to parse port",
			expected:    "",
		},
		{
			name:        "wrong port with an ip",
			in:          "127.0.0.1:bla",
			expectedErr: "unable to parse port",
			expected:    "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			host, port, err := daemon.ParseListenAddress(tc.in)
			out := host + ":" + strconv.Itoa(port)
			if tc.expectedErr != "" {
				if err != nil {
					if err.Error() != tc.expectedErr {
						t.Errorf("unexpected error, got \"%s\", expected \"%s\"", err.Error(), tc.expectedErr)
					}
				} else {
					t.Errorf("expecting error but not got, expected %s", tc.expectedErr)
				}
			} else {
				if out != tc.expected {
					t.Errorf("unexpected output, got \"%s\" expected \"%s\"", out, tc.expected)
				}
			}
		})
	}
}
