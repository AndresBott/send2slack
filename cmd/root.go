package cmd

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	"send2slack/internal/send2slack"
)

//var cfg *config.Cfg
//
//
//func init() {
//
//	var err error
//	cfg, err = config.NewConfig()
//	HandleErr(err)
//
//
//}

func Run() {

	cfg, err := send2slack.NewConfig()
	HandleErr(err)

	spew.Dump(cfg)

	binName := filepath.Base(os.Args[0])
	// handle the case where the app replaces sendmail binary
	if binName == "sendmail" || os.Getenv("SENDMAIL") == "debug" {
		// sendmail mode
		Sendmail() // todo move to s2slack package
	} else {

		rootCmd := &cobra.Command{
			Short: "Send messages to slack",
			Long: ` - send2slack v` + send2slack.Version + ` -
The send2slack behaves differently depending on configuration and invocation:
- send2slack can be started as http server to receive json payloads that will be delivered to slack
- TODO: can be started as file watcher, and all changes will be sent to slack ( /var/mail )
- a cli that sends json payloads to the json server
- a cli that (given the corresponding configuration) can send the messages directly without the server
- is a sendmail binary replacement, it accepts input streams in mail format to be sent to slack
`,
			Use: "send2slack (message)",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("send2slack")
				//send2SlackCli(cmd, args)
			},
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
		}

		rootCmd.AddCommand(serverCmd())

		// normal cli mode
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// send to slack command executed when the binary is called sendmail or the debug ENV is set.
// this removes all cobra features and makes send2slack compatible with sendmail by avoiding flag parses
func Sendmail() {
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
}

func HandleErr(err error) {
	if err != nil {
		fmt.Println("Fatal error: " + err.Error())
		os.Exit(0)
	}
}
