package sender_test

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"net/url"
	"send2slack/internal/config"
	"send2slack/internal/sender"
	"strings"
	"testing"
)

type test struct {
	description  string
	responseCode int
	msg          sender.Message
	errorString  string
}

func TestNewSlackSenderClient(t *testing.T) {
	tcs := []test{
		{
			description:  "simple message",
			responseCode: http.StatusAccepted,
			msg: sender.Message{
				Text: "test",
			},
			errorString: "",
		},
	}

	for _, test := range tcs {
		t.Run(test.description, func(t *testing.T) {

			// start the sample server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// test request header to contain Content-Type application/json
				gotCT := r.Header.Get("Content-Type")
				if diff := cmp.Diff("application/json", gotCT); diff != "" {
					t.Errorf("header mismatch (-want +got):\n%s", diff)
				}

				// verify the method
				if r.Method != http.MethodPost {
					t.Errorf("method mismatch want: %s, got %s", http.MethodPost, r.Method)
				}

				decoder := json.NewDecoder(r.Body)

				//here the httpclint should send a normal menssage,
				//and it is the server that needs to then transform that to slak mesage or other in the send2slack.EmptyBodyError
				//
				//therefore the server needs also to implement the sender interface
				var got sender.Message
				err := decoder.Decode(&got)
				if err != nil {
					t.Fatal("error decoding request body")
					return
				}

				err = got.Validate()
				if err != nil {
					t.Errorf("submitted payload does not pass message validation")
				}

				//send test response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.responseCode)

				fmt.Fprintln(w, "ok")

			}))
			defer ts.Close()

			u, _ := url.ParseRequestURI(ts.URL)
			c, err := sender.NewSlackSender(&config.ClientConfig{
				Url:  u,
				Mode: config.ModeHttpClient,
			})
			if err != nil {
				t.Fatal(err)
			}

			err = c.SendMessage(&test.msg)
			//expecting an error
			if test.errorString != "" {
				if err == nil {
					t.Errorf("error expected")
				}
				if !strings.Contains(err.Error(), test.errorString) {
					t.Errorf("error does not contain expected error string, expected '%v', got '%v'", test.errorString, err.Error())
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}

		})
	}
}
