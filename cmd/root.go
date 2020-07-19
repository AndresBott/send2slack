package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"send2slack/internal/send2slack"
	"strconv"
	"strings"
)

func Run() {

	// disable the whole sendmail implementation for now
	// mayne never added again

	//binName := filepath.Base(os.Args[0])
	//// handle the case where the app replaces sendmail binary
	//if binName == "sendmail" || os.Getenv("SENDMAIL") == "debug" {
	//	// sendmail mode
	//	SendmailCmd()
	//} else {
	send2SlackCmd()
	//}
}

type cmdParams struct {
	printExtHelp bool // print extended help
	verbose      bool
	configFile   string
	server       bool
	remote       string
	localRemote  bool
	channel      string
	color        string
}

// execute the normal mode command, NOT replacing sendmail
func send2SlackCmd() {

	params := cmdParams{}

	cmd := &cobra.Command{
		Short: "Send messages to slack",
		Long: ` == send2slack v` + send2slack.Version + ` ==
The send2slack behaves differently depending on configuration and invocation:
- send2slack can be started as http server to receive json payloads that will be delivered to slack
- TODO: can be started as file watcher, and all changes will be sent to slack ( /var/mail )
- a cli that sends json payloads to the json server
- a cli that (given the corresponding configuration) can send the messages directly without the server
- is a sendmail binary replacement, it accepts input streams in mail format to be sent to slack
`,
		Use: "send2slack (message)",
		Run: func(cmd *cobra.Command, args []string) {

			if params.printExtHelp {
				extendedHelp()
			} else if params.server {
				// start in server mode
				slackCfg := getSend2SlackConfig(params)

				daemon, err := send2slack.NewDaemon(slackCfg)
				HandleErr(err)
				daemon.Start()

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

	cmd.Flags().StringVarP(&params.configFile, "config", "f", "", "config file")

	cmd.Flags().BoolVarP(&params.server, "server", "s", false, "run in server mode")

	cmd.Flags().StringVarP(&params.remote, "remote", "r", "", "send message to remote proxy server")
	cmd.Flags().BoolVarP(&params.localRemote, "local-remote", "R", false, "same as \"remote\" but uses \"127.0.0.1:"+strconv.Itoa(send2slack.DefaultPort)+"\" as destination")

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

// getSend2SlackConfig uses the cli parameters to create a send2slack config
func getSend2SlackConfig(params cmdParams) *send2slack.Config {
	if params.verbose {
		if params.configFile == "" {
			cfgFile := " /etc/send2slack/config.yaml | $HOME/.send2slack/config.yaml | ./config.yaml "
			fmt.Printf("loading configuration file from: %s \n", cfgFile)
		} else {
			fmt.Printf("loading configuration file: \"%s\"\n", params.configFile)
		}
	}

	slackCfg, err := send2slack.NewConfig(params.configFile)
	HandleErr(err)

	if slackCfg.IsDefault && params.verbose {
		fmt.Printf("configuration file not found, using default values\n")
	}

	return slackCfg
}

//// send to slack command normally executed, see sendmail for exception
func send2SlackCli(cmd *cobra.Command, args []string, params cmdParams) {

	slackCfg := getSend2SlackConfig(params)

	var inText string
	var err error

	if len(args) > 0 {
		inText = args[0]
	} else {
		inText, err = send2slack.InReader(1000000) // 1MB
		if err != nil && err.Error() != "io reader not started" {
			HandleErr(err)
		}
	}

	// set the default channel if none is provided
	if params.channel == "" {
		params.channel = slackCfg.DefChannel
	}

	msg := send2slack.Message{
		Destination: params.channel,
		Color:       params.color,
		Text:        inText,
	}

	// ============================================
	// the message has been composed, now we send it
	// depending on the invocation, either in direct mode or in client mode.

	if params.remote != "" || params.localRemote {

		if params.localRemote {
			params.remote = "http://127.0.0.1:" + strconv.Itoa(send2slack.DefaultPort)
		}

		// prepend http:// if not provided already
		if !strings.HasPrefix(params.remote, "http://") && !strings.HasPrefix(params.remote, "https://") {
			params.remote = "http://" + params.remote
		}

		u, err := url.ParseRequestURI(params.remote)
		if err != nil {
			HandleErr(fmt.Errorf("invalid url: %v", err))
		}

		slackCfg.Mode = send2slack.ModeHttpClient
		slackCfg.URL = u

		if params.verbose {
			fmt.Printf("sending message in \"client mode\" to channel: \"%s\" using server: \"%s\" \n", params.channel, params.remote)
		}

	} else {
		if params.verbose {
			fmt.Printf("sending message in \"direct mode\" to channel: \"%s\"\n", params.channel)
		}
		slackCfg.Mode = send2slack.ModeDirectCli
	}

	slackSender, err := send2slack.NewSlackSender(slackCfg)
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

// send to slack command executed when the binary is called sendmail or the debug ENV is set.
// this removes all cobra features and makes send2slack compatible with sendmail by avoiding flag parses
//func SendmailCmd() {
//
//	//["/usr/sbin/sendmail", "-FCronDaemon", "-i", "-B8BITMIME", "-oem", "vagrant"]
//
//	//var err error
//	//slack := send2slack.NewSlackMsg(cfg.Token)
//	//slack.Channel = cfg.SendmailChannel
//	//slack.Text = ""
//	//slack.Detail, err = send2slack.InReader(1000000) // 1MB
//	//if err != nil {
//	//	HandleErr(err)
//	//}
//	//err = slack.SendMail()
//	//if err != nil {
//	//	HandleErr(err)
//	//}
//
//	slackCfg := getSend2SlackConfig(slackCmdConf{})
//
//
//	slackSender, err := send2slack.NewSlackSender(slackCfg)
//	HandleErr(err)
//	_ = slackSender
//	//err = slackSender.SendMessage(&slackMsg)
//	//HandleErr(err)
//
//}

func HandleErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v \n", err.Error())
		os.Exit(1)
	}
}
