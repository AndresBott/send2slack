package main

import "send2slack/cmd"

var (
	version = "v0.2.0-dev"
	commit  = ""
	date    = ""
)

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.Run()
}
