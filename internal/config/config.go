package config

import (
	"github.com/spf13/viper"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Mode int

const (
	ModeDirectCli   = 1
	ModeHttpClient  = 2
	ModeMailSending = 3
)
const DefaultPort = 4789

func readConfigFile(cfgFile string) (bool, error) {

	if cfgFile != "" {
		absPath, err := filepath.Abs(cfgFile)
		if err != nil {
			return false, err
		}
		viper.AddConfigPath(filepath.Dir(absPath))
		viper.SetConfigName(filepath.Base(absPath))
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.send2slack")
		viper.AddConfigPath("/etc/send2slack/")
	}

	viper.SetConfigType("yaml")

	viper.SetDefault("default_channel", "general")
	viper.SetDefault("sendmail_channel", "general")

	fileRead := true
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fileRead = false
		} else {
			return false, err
		}
	}
	return fileRead, nil
}

type DaemonConfig struct {
	IsDefault       bool   // set to true if no configuration file could be loaded
	ListenUrl       string // used by the server, listen address
	WatchDir        string // used by the server, watch for mbox dir
	Token           string
	DefChannel      string
	SendmailChannel string
	MailThrottling  int
}

func NewDaemonConfig(cfgFile string) (*DaemonConfig, error) {

	viper.SetConfigName("server.yaml")

	fileRead, err := readConfigFile(cfgFile)
	if err != nil {
		return nil, err
	}

	// overwrite configuration token if env "SLACK_TOKEN" is set
	slackToken := viper.GetString("slack.token")
	if envSlackToken := os.Getenv("SLACK_TOKEN"); envSlackToken != "" {
		slackToken = envSlackToken
	}
	// if we loaded config from a file, we are not using the default values
	defaultConfg := true
	if fileRead {
		defaultConfg = false
	}

	listenUrl := viper.GetString("daemon.listen_url")
	if defaultConfg {
		listenUrl = "127.0.0.1:" + strconv.Itoa(DefaultPort)
	}

	watchDir := viper.GetString("daemon.mbox_watch")
	if defaultConfg {
		watchDir = "false"
	}

	cfg := DaemonConfig{
		IsDefault:       defaultConfg,
		Token:           slackToken,
		DefChannel:      viper.GetString("slack.default_channel"),
		SendmailChannel: viper.GetString("slack.email_channel"),
		WatchDir:        watchDir,
		ListenUrl:       listenUrl,
		MailThrottling:  1000,
	}
	return &cfg, nil
}

type ClientConfig struct {
	IsDefault  bool // set to true if no configuration file could be loaded
	Mode       Mode
	Url        *url.URL
	Token      string
	DefChannel string
}

func NewClientConfig(cfgFile string) (*ClientConfig, error) {

	viper.SetConfigName("client.yaml")
	fileRead, err := readConfigFile(cfgFile)

	if err != nil {
		return nil, err
	}
	// overwrite configuration token if env "SLACK_TOKEN" is set
	slackToken := viper.GetString("slack.token")
	if envSlackToken := os.Getenv("SLACK_TOKEN"); envSlackToken != "" {
		slackToken = envSlackToken
	}

	// if we loaded config from a file, we are not using the default values
	defaultConfg := true
	if fileRead {
		defaultConfg = false
	}

	mode := Mode(ModeDirectCli)
	var u *url.URL

	remoteUrl := viper.GetString("client.remote_url")
	if remoteUrl != "" && remoteUrl != "false" {

		if !strings.HasPrefix(remoteUrl, "http://") && !strings.HasPrefix(remoteUrl, "https://") {
			remoteUrl = "http://" + remoteUrl
		}

		u, err = url.ParseRequestURI(remoteUrl)
		if err != nil {
			u = nil
			mode = ModeDirectCli
		}
		mode = ModeHttpClient
	}

	cfg := ClientConfig{
		IsDefault:  defaultConfg,
		Token:      slackToken,
		DefChannel: viper.GetString("slack.default_channel"),
		Url:        u,
		Mode:       mode,
	}
	return &cfg, nil
}
