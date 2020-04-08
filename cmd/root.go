package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"send2slack/internal/config"
	"send2slack/internal/send2slack"
)

var cfg *config.Cfg
var color string
var channel string

func init() {

	var err error
	cfg, err = config.NewConfig()
	HandleErr(err)

	rootCmd.Flags().StringVarP(&channel, "channel", "d", cfg.DefChannel, "destination channel to send the message")
	rootCmd.Flags().StringVarP(&color, "color", "c", "", "color")
}

var rootCmd = &cobra.Command{
	Short: "Send messages to a slack channel",
	Long: ` - send2slack v` + send2slack.Version + ` -
Send messages to a slack channel either by inline message or by input stream.
send2slack supports hex based colors like #ff00ff or the following shortcuts:red, green, blue, range, lime
usage samples:

  send2slack -c red -d general '<!here> Warning the silence is among us' // ! note the single quotes

  send2slack -c red <<EOF
  why so serious :smile:
  EOF

if this binary is renamed to sendmail (i.e. /usr/sbin/sendmail) it will ignore commandline parameters, 
but still send the stdin to slack, to the configured channel 

`,
	Use: "send2slack <message>",
	Run: func(cmd *cobra.Command, args []string) {
		send2SlackCli(cmd, args)
	},
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func Run() {

	binName := filepath.Base(os.Args[0])

	if binName == "sendmail" || os.Getenv("SENDMAIL") == "debug" {
		// sendmail mode
		Sendmail()
	} else {
		// normal cli mode
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// send to slack command normally executed, see sendmail for exception
func send2SlackCli(cmd *cobra.Command, args []string) {
	slack := send2slack.NewSlackMsg(cfg.Token)
	slack.Channel = channel
	var err error

	slack.Detail, err = send2slack.InReader(1000000) // 1MB
	if err != nil && err.Error() != "io reader not started" {
		HandleErr(err)
	}

	slack.Color(color)

	if len(args) > 0 {
		slack.Text = args[0]
	}
	err = slack.SendMsg()
	if err != nil && err.Error() == "unable to send empty message" {
		cmd.Help()
	} else {
		HandleErr(err)
	}
}

// send to slack command executed when the binary is called sendmail or the debug ENV is set.
// this removes all cobra features and makes send2slack compatible with sendmail by avoiding flag parses
func Sendmail() {
	//["/usr/sbin/sendmail", "-FCronDaemon", "-i", "-B8BITMIME", "-oem", "vagrant"]

	var err error
	slack := send2slack.NewSlackMsg(cfg.Token)
	slack.Channel = cfg.SendmailChannel
	slack.Text = ""
	slack.Detail, err = send2slack.InReader(1000000) // 1MB
	if err != nil {
		HandleErr(err)
	}
	err = slack.SendMail()
	if err != nil {
		HandleErr(err)
	}
}

func HandleErr(err error) {
	if err != nil {
		fmt.Println("Fatal error: " + err.Error())
		os.Exit(0)
	}
}
