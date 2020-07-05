package main_test

import (
	"bytes"
	"flag"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"send2slack/internal/send2slack"
	"strconv"
	"strings"
	"testing"
)

var rune2eTests bool

// read the cli flags
func init() {
	if flag.Lookup("e2e") == nil {
		flag.BoolVar(&rune2eTests, "e2e", false, "set flag to execute performance tests")
	}
}

func configExists() bool {
	filename, err := filepath.Abs("./config.yaml")
	if err != nil {
		return false
	}
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type e2eTestCase struct {
	name             string
	cmd              []string
	expectedExitCode int
	expectedErrorStr string
	token            string
}

func TestE2eCliDirectMode(t *testing.T) {
	if !rune2eTests {
		t.Log("Skipping e2e tests")
		t.Skip("Use -e2e flag to run")
	}

	if !configExists() {
		t.Fatal("config.yaml does not exists")
	}
	filename, _ := filepath.Abs("./config.yaml")

	slackCfg, err := send2slack.NewConfig(filename)
	if err != nil {
		t.Fatalf("unable to load config file: %v", err)
	}

	ts := "*[DIRECT MODE]* test message from e2e test in send2slack =>"

	tcs := []e2eTestCase{
		{
			name: "expect error empty message",
			cmd: []string{
				"go", "run", "main.go",
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: unable to send empty message",
		},
		{
			name: "direct Cli",
			cmd: []string{
				"go", "run", "main.go",
				ts + " simple color blue",
				"-d", "#trash", "-c", "blue",
			},
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "direct Cli different config file",
			cmd: []string{
				"go", "run", "main.go",
				ts + " nonexistent file",
				"-d", "#trash", "-c", "blue", "-f", "inexistent",
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: error sending slack message: not_authed",
		},

		{
			name: "direct Cli unsing env token: $S2S_TOKEN",
			cmd: []string{
				"go", "run", "main.go",
				ts + " using env token",
				"-d", "#trash", "-c", "orange", "-f", "inexistent",
			},
			expectedExitCode: 0,
			expectedErrorStr: "",
			token:            slackCfg.Token,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			exitCode := 0
			cmdFailure := false

			if tc.token != "" {
				os.Setenv("S2S_TOKEN", tc.token)
			}

			cmd := exec.Command(tc.cmd[0], tc.cmd[1:len(tc.cmd)]...)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			err := cmd.Run()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
					cmdFailure = true
				}
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("expected exit code does not match, got %d expected %d ", exitCode, tc.expectedExitCode)
			}

			if tc.expectedExitCode == 0 && cmdFailure {
				t.Errorf("expected command to work but it failed")
				t.Logf("Stdout %s", outb.String())
				t.Logf("Sterr %s", errb.String())
			}

			gotErr := errb.String()
			if !strings.Contains(gotErr, tc.expectedErrorStr) {
				t.Errorf("unexpected error string, got: %s expected %s", gotErr, tc.expectedErrorStr)
			}

			os.Unsetenv("S2S_TOKEN")

		})
	}
}

func TestE2eCliClientServerMode(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	if !rune2eTests {
		t.Log("Skipping e2e tests")
		t.Skip("Use -e2e flag to run")
	}

	if !configExists() {
		t.Fatal("config.yaml does not exists")
	}
	filename, _ := filepath.Abs("./config.yaml")

	slackCfg, err := send2slack.NewConfig(filename)
	if err != nil {
		t.Fatalf("unable to load config file: %v", err)
	}

	// get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	slackCfg.ListenUrl = ":" + strconv.Itoa(port)

	remoteUrl := "localhost:" + strconv.Itoa(port)

	server, err := send2slack.NewServer(slackCfg)
	if err != nil {
		t.Fatalf("unable to crete s2s server: %v", err)
	}
	server.StartBackground()

	ts := "*[CLIENT-SERVER MODE]* test message from e2e test in send2slack =>"
	tcs := []e2eTestCase{
		{
			name: "expect error empty message",
			cmd: []string{
				"go", "run", "main.go", "-r", remoteUrl,
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: unable to send empty message",
		},
		{
			name: "direct Cli",
			cmd: []string{
				"go", "run", "main.go",
				ts + " simple color blue",
				"-d", "#trash", "-c", "blue", "-r", remoteUrl,
			},
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "direct Cli different config file",
			cmd: []string{
				"go", "run", "main.go",
				ts + " nonexistent file",
				"-d", "#trash", "-c", "blue", "-f", "inexistent", "-r", remoteUrl,
			},
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			exitCode := 0
			cmdFailure := false

			if tc.token != "" {
				os.Setenv("S2S_TOKEN", tc.token)
			}

			cmd := exec.Command(tc.cmd[0], tc.cmd[1:len(tc.cmd)]...)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			err := cmd.Run()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
					cmdFailure = true
				}
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("expected exit code does not match, got %d expected %d ", exitCode, tc.expectedExitCode)
			}

			if tc.expectedExitCode == 0 && cmdFailure {
				t.Errorf("expected command to work but it failed")
				t.Logf("Stdout %s", outb.String())
				t.Logf("Sterr %s", errb.String())
			}

			gotErr := errb.String()
			if !strings.Contains(gotErr, tc.expectedErrorStr) {
				t.Errorf("unexpected error string, got: %s expected %s", gotErr, tc.expectedErrorStr)
			}

			os.Unsetenv("S2S_TOKEN")

		})
	}

	server.Stop()
}

// the whole sendmail implementation is disabled for now
//
//func TestE2eSendmailMode(t *testing.T)  {
//	if !rune2eTests {
//		t.Log("Skipping e2e tests")
//		t.Skip("Use -e2e flag to run")
//	}
//
//	if !configExists() {
//		t.Fatal("config.yaml does not exists")
//	}
//	filename, _ := filepath.Abs("./config.yaml")
//	spew.Dump(filename)
//
//	os.Setenv("SENDMAIL", "debug")
//
//
//	slackCfg,err := send2slack.NewConfig(filename)
//	_= slackCfg
//	if err != nil{
//		t.Fatalf("unable to load config file: %v",err)
//	}
//
//	ts := "this is a test message from e2e test in send2slack =>"
//
//
//	tcs := []e2eTestCase{
//
//		{
//			name : "sendmail",
//			cmd: []string{
//				"go", "run", "main.go",
//				ts+ " testing direct CLI and color blue",
//				"-d", "#trash", "-c", "blue",
//			},
//			expectedExitCode: 1,
//			expectedErrorStr: "sfsgads",
//		},
//		//{
//		//	name : "direct Cli different config file",
//		//	cmd: []string{
//		//		"go", "run", "main.go",
//		//		ts+ " testing direct CLI and color blue",
//		//		"-d", "#trash", "-c", "blue", "-f", "inexistent",
//		//	},
//		//	expectedExitCode: 1,
//		//	expectedErrorStr: "Fatal error: error sending slack message: not_authed",
//		//},
//		//
//		//{
//		//	name : "direct Cli unsing env token: $S2S_TOKEN",
//		//	cmd: []string{
//		//		"go", "run", "main.go",
//		//		ts+ " testing direct CLI with environmental variable $S2S_TOKEN and color orange",
//		//		"-d", "#trash", "-c", "orange", "-f", "inexistent",
//		//	},
//		//	expectedExitCode: 0,
//		//	expectedErrorStr: "",
//		//	token: slackCfg.Token,
//		//},
//	}
//
//
//
//	for _, tc := range tcs{
//		t.Run(tc.name, func(t *testing.T) {
//
//			exitCode := 0
//			cmdFailure := false
//
//			if tc.token != ""{
//				os.Setenv("S2S_TOKEN", tc.token)
//			}
//
//			cmd := exec.Command(tc.cmd[0], tc.cmd[1:len(tc.cmd)]...)
//
//			var outb, errb bytes.Buffer
//			cmd.Stdout = &outb
//			cmd.Stderr = &errb
//
//			err :=  cmd.Run()
//			if err != nil {
//				if exitError, ok := err.(*exec.ExitError); ok {
//					exitCode = exitError.ExitCode()
//					cmdFailure = true
//				}
//			}
//
//			//fmt.Println("out:", outb.String(), "err:", errb.String())
//
//			if exitCode != tc.expectedExitCode{
//				t.Errorf("expected exit code does not match, got %d expected %d ", exitCode,tc.expectedExitCode)
//			}
//
//			if tc.expectedExitCode == 0 && cmdFailure{
//				t.Errorf("expected command to work but it failed")
//				t.Logf("Stdout %s",outb.String())
//				t.Logf("Sterr %s",errb.String())
//			}
//
//			gotErr := errb.String()
//			if ! strings.Contains(gotErr, tc.expectedErrorStr){
//				t.Errorf("unexpected error string, got: \"%s\" expected \"%s\"", gotErr, tc.expectedErrorStr )
//			}
//
//			os.Unsetenv("S2S_TOKEN")
//
//		})
//	}
//
//	os.Unsetenv("SENDMAIL")
//}
//
//
