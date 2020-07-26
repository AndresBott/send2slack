package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"send2slack/internal/config"
	"send2slack/internal/daemon"
	"send2slack/internal/sender"
	"strconv"
	"strings"
)

type cmdParams struct {
	printExtHelp bool // print extended help
	verbose      bool
	version      bool
	configFile   string
	server       bool
	watcher      bool
	remote       string
	localRemote  bool
	channel      string
	color        string
}

var (
	Version = ""
	Commit  = ""
	Date    = ""
)

// execute the normal mode command, NOT replacing sendmail
func Run() {

	params := cmdParams{}

	cmd := &cobra.Command{
		Short: "Send messages to slack",
		Long: ` == send2slack v` + Version + ` ==
The send2slack behaves differently depending on configuration and invocation:
- send2slack can be started as http server to receive json payloads that will be delivered to slack
- TODO: can be started as file watcher, and all changes will be sent to slack ( /var/mail )
- a cli that sends json payloads to the json server
- a cli that (given the corresponding configuration) can send the messages directly without the server
- is a sendmail binary replacement, it accepts input streams in mail format to be sent to slack
`,
		Use: "send2slack (message)",
		Run: func(cmd *cobra.Command, args []string) {

			if params.version {
				printVersion()
			} else if params.printExtHelp {
				extendedHelp()
			} else if params.server || params.watcher {
				// start in server mode
				send2SlackDaemon(params)
			} else {
				// use the client to send a message
				send2SlackCli(cmd, args, params)
			}

		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.Flags().BoolVarP(&params.printExtHelp, "extended-help", "H", false, "print the extended help")
	cmd.Flags().BoolVarP(&params.verbose, "verbose", "v", false, "verbose mode")
	cmd.Flags().BoolVarP(&params.version, "version", "V", false, "print version")

	cmd.Flags().StringVarP(&params.configFile, "config", "f", "", "config file")

	cmd.Flags().BoolVarP(&params.server, "server", "s", false, "run in server mode")
	cmd.Flags().BoolVarP(&params.watcher, "watch", "w", false, "run in mbox watcher mode")

	cmd.Flags().StringVarP(&params.remote, "remote", "r", "", "send message to remote proxy server")
	cmd.Flags().BoolVarP(&params.localRemote, "local-remote", "R", false, "same as \"remote\" but uses \"127.0.0.1:"+strconv.Itoa(config.DefaultPort)+"\" as destination")

	cmd.Flags().StringVarP(&params.channel, "channel", "d", "", "destination channel to send the message")
	cmd.Flags().StringVarP(&params.color, "color", "c", "", "color")

	// normal cli mode
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func extendedHelp() {
	fmt.Println("todo: Extended help")
}

func printVersion() {

	v := `Version: ` + Version + `
Commit id: ` + Commit + `
Build date: ` + Date
	fmt.Println(v)
}

// getSend2SlackDaemonConfig uses the cli parameters to create a daemon config
func getSend2SlackDaemonConfig(params cmdParams) *config.DaemonConfig {

	if params.configFile == "" {
		cfgFile := " /etc/send2slack/server.yaml | $HOME/.send2slack/server.yaml | ./server.yaml "
		logrus.Infof("loading server configuration file from: %s", cfgFile)
	} else {
		logrus.Infof("loading server configuration file: \"%s\"", params.configFile)
	}

	cfg, err := config.NewDaemonConfig(params.configFile)
	HandleErr(err)

	if cfg.IsDefault {
		logrus.Infof("configuration file not found, using default values")
	} else {
		logrus.Infof("using configuration file: %s", viper.ConfigFileUsed())
	}

	// disable watcher if not enabled with flag
	if !params.watcher {
		cfg.WatchDir = "false"
	}

	// disable server if not enabled with flag
	if !params.server {
		cfg.ListenUrl = "false"
	}

	return cfg
}

// send to slack command normally executed, see sendmail for exception
func send2SlackDaemon(params cmdParams) {
	// start in server mode
	daemonCfg := getSend2SlackDaemonConfig(params)

	dmn, err := daemon.NewDaemon(daemonCfg)
	HandleErr(err)
	dmn.Start()
}

// getSend2SlackConfig uses the cli parameters to create a send2slack config
func getSend2SlackClientConfig(params cmdParams) *config.ClientConfig {
	if params.verbose {
		if params.configFile == "" {
			cfgFile := " /etc/send2slack/client.yaml | $HOME/.send2slack/client.yaml | ./client.yaml "
			fmt.Printf("loading configuration file from: %s \n", cfgFile)
		} else {
			fmt.Printf("loading configuration file: \"%s\"\n", params.configFile)
		}
	}

	slackCfg, err := config.NewClientConfig(params.configFile)
	HandleErr(err)

	if params.verbose {
		if slackCfg.IsDefault {
			fmt.Printf("configuration file not found, using default values\n")
		} else {
			fmt.Printf("using configuration file: %s\n", viper.ConfigFileUsed())
		}
	}

	return slackCfg
}

//// send to slack command normally executed, see sendmail for exception
func send2SlackCli(cmd *cobra.Command, args []string, params cmdParams) {

	slackCfg := getSend2SlackClientConfig(params)

	var inText string
	var err error

	if len(args) > 0 {
		inText = args[0]
	} else {
		inText, err = sender.InReader(1000000) // 1MB
		if err != nil && err.Error() != "io reader not started" {
			HandleErr(err)
		}
	}

	// set the default channel if none is provided
	if params.channel == "" {
		params.channel = slackCfg.DefChannel
	}

	msg := sender.Message{
		Destination: params.channel,
		Color:       params.color,
		Text:        inText,
	}

	// ============================================
	// the message has been composed, now we send it
	// depending on the invocation, either in direct mode or in client mode.

	if params.remote != "" || params.localRemote {

		if params.localRemote {
			params.remote = "http://127.0.0.1:" + strconv.Itoa(config.DefaultPort)
		}

		// prepend http:// if not provided already
		if !strings.HasPrefix(params.remote, "http://") && !strings.HasPrefix(params.remote, "https://") {
			params.remote = "http://" + params.remote
		}

		u, err := url.ParseRequestURI(params.remote)
		if err != nil {
			HandleErr(fmt.Errorf("invalid url: %v", err))
		}

		slackCfg.Mode = config.ModeHttpClient
		slackCfg.Url = u

		if params.verbose {
			fmt.Printf("sending message in \"client mode\" to channel: \"%s\" using server: \"%s\" \n", params.channel, params.remote)
		}

	} else {
		// if uri has been defined in configuration we ignore the direct cli fallback
		if slackCfg.Url == nil {
			if params.verbose {
				fmt.Printf("sending message in \"direct mode\" to channel: \"%s\"\n", params.channel)
			}
			slackCfg.Mode = config.ModeDirectCli
		}

	}

	slackSender, err := sender.NewSlackSender(slackCfg)
	HandleErr(err)

	err = slackSender.SendMessage(&msg)
	HandleErr(err)

	if err != nil && err.Error() == "unable to send empty message" {
		fmt.Print("> Unable to send empty message \n\n")
		cmd.Help()
	} else {
		HandleErr(err)
	}
}

func HandleErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v \n", err.Error())
		os.Exit(1)
	}
}
