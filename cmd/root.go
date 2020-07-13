package cmd

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"send2slack/internal/send2slack"
	"strconv"
	"strings"
)

func Run() {

	// disable the whole sendmail implementation for now

	//binName := filepath.Base(os.Args[0])
	//// handle the case where the app replaces sendmail binary
	//if binName == "sendmail" || os.Getenv("SENDMAIL") == "debug" {
	//	// sendmail mode
	//	SendmailCmd()
	//} else {
	send2SlackCmd()
	//}
}

type slackCmdConf struct {
	configFile string

	printExtHelp bool
	verbose      bool
	channel      string
	color        string
	remote       string

	server       bool
	serverListen string

	localRemote bool
}

// execute the normal mode command, NOT replacing sendmail
func send2SlackCmd() {

	cfg := slackCmdConf{}

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

			if cfg.printExtHelp {
				extendedHelp()
			} else if cfg.server {
				// start in server mode
				send2SlackServer(cfg)
			} else {
				// use the client to send a message
				send2SlackCli(cmd, args, cfg)
			}

		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.Flags().BoolVarP(&cfg.printExtHelp, "extended-help", "H", false, "print the extended help")
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", false, "verbose mode")

	cmd.Flags().StringVarP(&cfg.configFile, "config", "f", "", "config file")

	cmd.Flags().BoolVarP(&cfg.server, "server", "s", false, "run in server mode")
	cmd.Flags().StringVarP(&cfg.serverListen, "listen", "l", "127.0.0.1:"+strconv.Itoa(send2slack.DefaultPort), "listen on <url> | \"false\" to disable")

	cmd.Flags().StringVarP(&cfg.remote, "remote", "r", "", "send message to remote proxy server")
	cmd.Flags().BoolVarP(&cfg.localRemote, "local-remote", "R", false, "same as \"remote\" but uses \"127.0.0.1:"+strconv.Itoa(send2slack.DefaultPort)+"\" as destination")

	cmd.Flags().StringVarP(&cfg.channel, "channel", "d", "", "destination channel to send the message")
	cmd.Flags().StringVarP(&cfg.color, "color", "c", "", "color")

	// normal cli mode
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func extendedHelp() {
	fmt.Println("Extended help")
}

func getSend2SlackConfig(cfg slackCmdConf) *send2slack.Config {
	if cfg.verbose {

		if cfg.configFile == "" {
			cfgFile := " /etc/send2slack/config.yaml | $HOME/.send2slack/config.yaml | ./config.yaml "
			fmt.Printf("loading configuration file from: %s \n", cfgFile)
		} else {
			fmt.Printf("loading configuration file: \"%s\"\n", cfg.configFile)
		}

	}
	slackCfg, err := send2slack.NewConfig(cfg.configFile)
	HandleErr(err)

	if slackCfg.IsDefault && cfg.verbose {
		fmt.Printf("configuration file not found, using default values\n")
	}

	return slackCfg
}

//// send to slack command normally executed, see sendmail for exception
func send2SlackCli(cmd *cobra.Command, args []string, cfg slackCmdConf) {

	slackCfg := getSend2SlackConfig(cfg)

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
	if cfg.channel == "" {
		cfg.channel = slackCfg.DefChannel
	}

	slackMsg := send2slack.Message{
		Destination: cfg.channel,
		Color:       cfg.color,
		Text:        inText,
	}

	// ============================================
	// the message has been composed, now we send it
	// depending on the invocation, either in direct mode or in client mode.

	if cfg.remote != "" || cfg.localRemote {

		if cfg.localRemote {
			cfg.remote = "http://127.0.0.1:" + strconv.Itoa(send2slack.DefaultPort)
		}

		// prepend http:// if not provided already
		if !strings.HasPrefix(cfg.remote, "http://") && !strings.HasPrefix(cfg.remote, "https://") {
			cfg.remote = "http://" + cfg.remote
		}

		u, err := url.ParseRequestURI(cfg.remote)
		if err != nil {
			HandleErr(fmt.Errorf("invalid url: %v", err))
		}

		slackCfg.Mode = send2slack.ModeHttpClient
		slackCfg.URL = u

		if cfg.verbose {
			fmt.Printf("sending message in \"client mode\" to channel: \"%s\" using server: \"%s\" \n", cfg.channel, cfg.remote)
		}

	} else {
		if cfg.verbose {
			fmt.Printf("sending message in \"direct mode\" to channel: \"%s\"\n", cfg.channel)
		}
		slackCfg.Mode = send2slack.ModeDirectCli
	}

	slackSender, err := send2slack.NewSlackSender(slackCfg)
	HandleErr(err)

	err = slackSender.SendMessage(&slackMsg)
	HandleErr(err)

	if err != nil && err.Error() == "unable to send empty message" {
		fmt.Print("> Unable to send empty message \n\n")
		cmd.Help()
	} else {
		HandleErr(err)
	}
}

func send2SlackServer(cfg slackCmdConf) {
	slackCfg := getSend2SlackConfig(cfg)
	_ = slackCfg

	if cfg.serverListen == "false" {
		if cfg.verbose {
			fmt.Printf("slack server has been disabled by \"--listen false\", not starting\n")
		}
		slackCfg.Mode = send2slack.ModeNoServerWatch
	} else {
		if cfg.verbose {
			fmt.Printf("starting slack server on %s, \n", cfg.serverListen)
		}
		slackCfg.Mode = send2slack.ModeServerNoWatch
		slackCfg.ListenUrl = cfg.serverListen
	}

	slackServer, err := send2slack.NewServer(slackCfg)
	HandleErr(err)

	slackServer.Start()

}

// send to slack command executed when the binary is called sendmail or the debug ENV is set.
// this removes all cobra features and makes send2slack compatible with sendmail by avoiding flag parses
func SendmailCmd() {

	//["/usr/sbin/sendmail", "-FCronDaemon", "-i", "-B8BITMIME", "-oem", "vagrant"]

	//var err error
	//slack := send2slack.NewSlackMsg(cfg.Token)
	//slack.Channel = cfg.SendmailChannel
	//slack.Text = ""
	//slack.Detail, err = send2slack.InReader(1000000) // 1MB
	//if err != nil {
	//	HandleErr(err)
	//}
	//err = slack.SendMail()
	//if err != nil {
	//	HandleErr(err)
	//}

	slackCfg := getSend2SlackConfig(slackCmdConf{})

	spew.Dump(slackCfg)

	slackSender, err := send2slack.NewSlackSender(slackCfg)
	HandleErr(err)
	_ = slackSender
	//err = slackSender.SendMessage(&slackMsg)
	//HandleErr(err)

}

func HandleErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v \n", err.Error())
		os.Exit(1)
	}
}
