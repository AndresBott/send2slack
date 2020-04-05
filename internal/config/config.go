package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

type Cfg struct {
	Token string
}

func NewConfig() (*Cfg, error) {

	viper.SetConfigName("config")           // name of config file (without extension)
	viper.SetConfigType("yaml")             // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/send2slack/") // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")   // call multiple times to add many search paths
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	err := viper.ReadInConfig()             // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		return nil, errors.New(fmt.Sprintf("Fatal error config file: %s \n", err))

	}

	cfg := Cfg{
		Token: viper.GetString("token"),
	}
	return &cfg, nil
}
