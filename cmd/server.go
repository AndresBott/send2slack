package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func serverCmd() *cobra.Command {

	var subCmd = &cobra.Command{
		Use:   "server",
		Short: "Start send2slack server",
		//PreRun: func(cmd *cobra.Command, args []string) {
		//	fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
		//},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Inside subCmd Run with args: %v\n", args)
		},
		//PostRun: func(cmd *cobra.Command, args []string) {
		//	fmt.Printf("Inside subCmd PostRun with args: %v\n", args)
		//},
		//PersistentPostRun: func(cmd *cobra.Command, args []string) {
		//	fmt.Printf("Inside subCmd PersistentPostRun with args: %v\n", args)
		//},
	}

	return subCmd

}
