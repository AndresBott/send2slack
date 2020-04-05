package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"send2slack/internal/config"
	"send2slack/internal/slackmsg"
)

var text string
var details string
var channel string

func init() {
	rootCmd.Flags().StringVarP(&channel, "channel", "c", "general", "channel to send the message to")
	rootCmd.Flags().StringVarP(&text, "text", "t", "", "text to send")
	rootCmd.Flags().StringVarP(&details, "details", "d", "", "details")
}

var rootCmd = &cobra.Command{
	Use:   "send2slack",
	Short: "todo ",
	Long:  `todo`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.NewConfig()
		_ = cfg

		slack := slackmsg.NewSlackMsg(cfg.Token)
		slack.Channel = channel

		slack.Warn(text, details)
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
