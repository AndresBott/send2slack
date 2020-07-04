package cmd

//rootCmd.Flags().StringVarP(&channel, "channel", "d", cfg.DefChannel, "destination channel to send the message")
//rootCmd.Flags().StringVarP(&color, "color", "c", "", "color")
//
//
//// send to slack command normally executed, see sendmail for exception
//func send2SlackCli(cmd *cobra.Command, args []string) {
//	slack := send2slack.NewSlackMsg(cfg.Token)
//	slack.Channel = channel
//	var err error
//
//	slack.Detail, err = send2slack.InReader(1000000) // 1MB
//	if err != nil && err.Error() != "io reader not started" {
//		HandleErr(err)
//	}
//
//	slack.Color(color)
//
//	if len(args) > 0 {
//		slack.Text = args[0]
//	}
//	err = slack.SendMsg()
//	if err != nil && err.Error() == "unable to send empty message" {
//		cmd.Help()
//	} else {
//		HandleErr(err)
//	}
//}
