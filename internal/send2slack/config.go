package send2slack

import (
	"github.com/spf13/viper"
	"net/url"
	"path/filepath"
)

const Version = "0.2.0"

type Mode int

const (
	ModeDirectCli     = 1
	ModeClientCli     = 2
	ModeServerNoWatch = 3
	ModeNoServerWatch = 4
	ModeServerWatch   = 5
)

const DefaultPort = 4789

type Config struct {
	IsDefault       bool   // set to true if no configuration file could be loaded
	ListenUrl       string // used by the server, listen address
	WatchDir        string
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

	viper.SetDefault("default_channel", "general")
	viper.SetDefault("sendmail_channel", "general")

	viper.SetEnvPrefix("s2s")
	viper.BindEnv("token")

	cfg := Config{
		IsDefault:       defaultConfg,
		Token:           viper.GetString("token"),
		DefChannel:      viper.GetString("default_channel"),
		SendmailChannel: viper.GetString("sendmail_channel"),
	}
	return &cfg, nil
}
