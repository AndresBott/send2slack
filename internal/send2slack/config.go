package send2slack

import (
	"github.com/spf13/viper"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

const Version = "0.2.0"

type Mode int

const (
	ModeDirectCli       = 1
	ModeHttpClient      = 2
	ModeServerNoWatch   = 3
	ModeNoServerWatch   = 4
	ModeServerWatch     = 5
	ModeNoServerNoWatch = 6
	ModeMailSending     = 7
)

const DefaultPort = 4789

type Config struct {
	IsDefault       bool     // set to true if no configuration file could be loaded
	ListenUrl       string   // used by the server, listen address
	WatchDir        string   // used by the server, watch for mbox dir
	URL             *url.URL // used by the client, send url
	Token           string
	DefChannel      string
	SendmailChannel string
	Mode            Mode
}

func NewConfig(cfgFile string) (*Config, error) {

	if cfgFile != "" {
		absPath, err := filepath.Abs(cfgFile)
		if err != nil {
			return nil, err
		}
		viper.AddConfigPath(filepath.Dir(absPath))
		viper.SetConfigName(filepath.Base(absPath))
	} else {
		viper.SetConfigName("config.yaml")
		viper.AddConfigPath("/etc/send2slack/")
		viper.AddConfigPath("$HOME/.send2slack")
		viper.AddConfigPath(".")
	}

	viper.SetConfigType("yaml")

	viper.SetDefault("default_channel", "general")
	viper.SetDefault("sendmail_channel", "general")

	defaultConfg := false
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			defaultConfg = true
		} else {
			return nil, err
		}
	}

	// overwrite configuration token if env "SLACK_TOKEN" is set
	slackToken := viper.GetString("slack.token")
	if envSlackToken := os.Getenv("SLACK_TOKEN"); envSlackToken != "" {
		slackToken = envSlackToken
	}

	listenUrl := viper.GetString("daemon.listen_url")
	// populate default config
	if defaultConfg {
		listenUrl = "127.0.0.1:" + strconv.Itoa(DefaultPort)
	}

	cfg := Config{
		IsDefault:       defaultConfg,
		Token:           slackToken,
		DefChannel:      viper.GetString("slack.default_channel"),
		SendmailChannel: viper.GetString("slack.sendmail_channel"),
		WatchDir:        viper.GetString("daemon.mbox_watch"),
		ListenUrl:       listenUrl,
	}
	return &cfg, nil
}
