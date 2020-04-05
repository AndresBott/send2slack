package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
`,
	Use: "send2slack <message>",
	Run: func(cmd *cobra.Command, args []string) {

		slack := send2slack.NewSlackMsg(cfg.Token)
		slack.Channel = channel

		d, err := send2slack.InReader(1000000) // 1MB
		if err != nil && err.Error() != "io reader not started" {
			HandleErr(err)
		}

		msg := ""
		if len(args) > 0 {
			msg = args[0]
		}

		err = slack.SendMsg(msg, d, color)
		if err != nil && err.Error() == "unable to send empty message" {
			cmd.Help()
		} else {
			HandleErr(err)
		}
	},
}

func RootCmd() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func HandleErr(err error) {
	if err != nil {
		fmt.Println("Fatal error: " + err.Error())
		os.Exit(0)
	}
}
