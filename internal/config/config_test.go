package config_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"net/url"
	"send2slack/internal/config"
	"strings"
	"testing"
)

type configTc struct {
	name           string
	file           string
	ClientExpected *config.ClientConfig
	DaemonExpected *config.DaemonConfig
	expectedErr    string
}

func getUrl(s string) *url.URL {
	u, _ := url.ParseRequestURI(s)
	return u
}

func TestNewClientConfig(t *testing.T) {

	tcs := []configTc{
		{
			name: "test default config on non existent file",
			file: "sampledata/doesNotExist",
			ClientExpected: &config.ClientConfig{
				IsDefault: true,
				Mode:      config.ModeDirectCli,
			},
			expectedErr: "",
		},
		{
			name: "test sample file",
			file: "sampledata/client.yaml",
			ClientExpected: &config.ClientConfig{
				IsDefault:  false,
				DefChannel: "general",
				Token:      "my_token",
				Mode:       config.ModeHttpClient,
				Url:        getUrl("http://127.0.0.1:4789"),
			},
			expectedErr: "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			viper.Reset()

			cfg, err := config.NewClientConfig(tc.file)

			if tc.expectedErr != "" && err != nil {

				got := err.Error()

				if !strings.Contains(got, tc.expectedErr) {
					if diff := cmp.Diff(tc.ClientExpected, got); diff != "" {
						t.Errorf("%s: my test: (-got +want)\n%s", tc.name, diff)
					}
				}
			} else {
				// we don't expect an error
				got := cfg

				if diff := cmp.Diff(tc.ClientExpected, got); diff != "" {
					t.Errorf("%s: my test: (-got +want)\n%s", tc.name, diff)
				}

			}

		})
	}
}

func TestNewDaemonConfig(t *testing.T) {
	tcs := []configTc{
		{
			name: "test default config on non existent file",
			file: "sampledata/doesNotExist",
			DaemonExpected: &config.DaemonConfig{
				IsDefault:      true,
				ListenUrl:      "127.0.0.1:4789",
				WatchDir:       "false",
				MailThrottling: 1000,
			},
			expectedErr: "",
		},
		{
			name: "test sample file",
			file: "sampledata/server.yaml",
			DaemonExpected: &config.DaemonConfig{
				IsDefault:       false,
				ListenUrl:       "127.0.0.1:1234",
				WatchDir:        "/var/mail",
				Token:           "my_token",
				DefChannel:      "general",
				SendmailChannel: "general",
				MailThrottling:  1000,
			},
			expectedErr: "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			viper.Reset()

			cfg, err := config.NewDaemonConfig(tc.file)

			if tc.expectedErr != "" && err != nil {

				got := err.Error()

				if !strings.Contains(got, tc.expectedErr) {
					if diff := cmp.Diff(tc.DaemonExpected, got); diff != "" {
						t.Errorf("%s: my test: (-got +want)\n%s", tc.name, diff)
					}
				}
			} else {
				// we don't expect an error
				got := cfg

				if diff := cmp.Diff(tc.DaemonExpected, got); diff != "" {
					t.Errorf("%s: my test: (-got +want)\n%s", tc.name, diff)
				}

			}

		})
	}
}
