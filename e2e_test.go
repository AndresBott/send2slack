package main_test

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"send2slack/internal/send2slack"
	"strconv"
	"strings"
	"testing"
	"time"
)

var rune2eTests bool

// read the cli flags
func init() {
	if flag.Lookup("e2e") == nil {
		flag.BoolVar(&rune2eTests, "e2e", false, "set flag to execute performance tests")
	}
}

type e2eTestCase struct {
	name             string
	cmd              []string
	copyCfg          bool // copy the generated config file from /tmp to current folder
	text             string
	expectedExitCode int
	expectedErrorStr string
	token            string
}

// prePateStage will read the slack token from a file "./token" and generate a tmp dir that contains
// a valid configuration file as well as an mbox directory,
//
// return is an instance of slack config, the string path to the mbox folder and a free port to be used
func prePateStage() (path string, port int, er error) {

	// check for file with token
	filename, err := filepath.Abs("./token")
	if err != nil {
		return "", -1, err
	}
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return "", -1, err
	}
	if info.IsDir() {
		return "", -1, fmt.Errorf("token is not a file")
	}

	// generate a tmp dir
	tmpDir, err := ioutil.TempDir("/tmp", "s2s_e2e_")
	if err != nil {
		return "", -1, err
	}

	// read token
	token, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", -1, err
	}

	// get a free port
	newPort, err := freeport.GetFreePort()
	if err != nil {
		return "", -1, err
	}
	listenUrl := "localhost:" + strconv.Itoa(newPort)

	// create the config file
	cfgStr := `---
slack:
  token: "` + string(token) + `"
  default_channel: "general"
  sendmail_channel: "general"
daemon:
  listen_url: "` + listenUrl + `"
  mbox_watch: "` + tmpDir + `/mbox"

`
	err = ioutil.WriteFile(tmpDir+"/config.yaml", []byte(cfgStr), 0644)
	if err != nil {
		return "", -1, err
	}

	// create the mbox
	err = os.MkdirAll(tmpDir+"/mbox", os.ModePerm)
	if err != nil {
		return "", -1, err
	}
	return tmpDir, newPort, nil
}

// e2e test to send a message directly by loading the configuration file / or using an env var to populate the token
func TestE2eCliDirectMode(t *testing.T) {
	if !rune2eTests {
		t.Log("Skipping e2e tests")
		t.Skip("Use -e2e flag to run")
	}

	tmpPath, _, err := prePateStage()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpPath)

	// load the config
	slackCfg, err := send2slack.NewConfig(tmpPath + "/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	tcs := []e2eTestCase{
		{
			name: "no config file nor message",
			cmd: []string{
				"go", "run", "main.go",
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: text cannot be empty",
		},
		{
			name: "expect error empty message",
			cmd: []string{
				"go", "run", "main.go", "-f", tmpPath + "/config.yaml",
			},
			copyCfg:          true,
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: text cannot be empty",
		},
		{
			name: "text without color",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-f", tmpPath + "/config.yaml",
			},
			text:             "text without color",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "setting a color",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-c", "orange", "-f", tmpPath + "/config.yaml",
			},
			text:             "setting a color",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "non existent config file",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-c", "orange", "-f", "inexistent",
			},
			text:             "non existent config file",
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: error sending slack message: not_authed",
		},
		{
			name: "config file loaded from current dir",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-c", "orange",
			},
			copyCfg:          true,
			text:             "config file loaded from current dir in ./config.yaml",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "using env variable SLACK_TOKEN",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-c", "orange", "-f", "inexistent",
			},
			text:             "using env variable SLACK_TOKEN",
			expectedExitCode: 0,
			expectedErrorStr: "",
			token:            slackCfg.Token,
		},
	}

	i := 1
	testCount := 0
	for _, tc := range tcs {
		if tc.expectedExitCode == 0 {
			testCount++
		}
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			exitCode := 0
			cmdFailure := false

			// if using env token
			if tc.token != "" {
				os.Setenv("SLACK_TOKEN", tc.token)
			}
			// copy config to local dir
			if tc.copyCfg {
				input, err := ioutil.ReadFile(tmpPath + "/config.yaml")
				if err != nil {
					t.Fatal(err)

				}
				err = ioutil.WriteFile("./config.yaml", input, 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					os.Remove("./config.yaml")
				}()
			}

			// modify the test text
			msg := tc.text
			if tc.text != "" {
				msg = "*[E2E test] [DIRECT CLI MODE] (" + strconv.Itoa(i) + "/" + strconv.Itoa(testCount) + ")* => " + tc.text
			}
			tc.cmd = append(tc.cmd, msg)

			// execute the test
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

			outStr := outb.String()
			if outStr != "" {
				t.Log("cmd output")
				t.Log(outStr)
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("expected exit code does not match, got %d expected %d ", exitCode, tc.expectedExitCode)
			}

			// only increase test counter for tests that are expected to be sent
			if tc.expectedExitCode == 0 && exitCode == 0 {
				i++
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

			os.Unsetenv("SLACK_TOKEN")

		})
	}
}

// e2e test to send messages from the cli to a http server listening on localhost, the server then forwards the
// messages to slack.
func TestE2EHttpServerMode(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	if !rune2eTests {
		t.Log("Skipping e2e tests")
		t.Skip("Use -e2e flag to run")
	}

	tmpPath, port, err := prePateStage()
	if err != nil {
		t.Fatal(err)
	}
	remoteUrl := "localhost:" + strconv.Itoa(port)

	// load the config
	slackCfg, err := send2slack.NewConfig(tmpPath + "/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	server, err := send2slack.NewServer(slackCfg)
	if err != nil {
		t.Fatalf("unable to crete s2s server: %v", err)
	}
	server.StartBackground()

	tcs := []e2eTestCase{
		{
			name: "no config file nor message",
			cmd: []string{
				"go", "run", "main.go",
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: text cannot be empty",
		},
		{
			name: "expect error empty message",
			cmd: []string{
				"go", "run", "main.go", "-r", remoteUrl,
			},
			expectedExitCode: 1,
			expectedErrorStr: "Fatal error: text cannot be empty",
		},
		{
			name: "text without color",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-r", remoteUrl,
			},
			text:             "text without color",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "setting a color",
			cmd: []string{
				"go", "run", "main.go",

				"-d", "#trash", "-c", "blue", "-r", remoteUrl,
			},
			text:             "setting a color",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		{
			name: "missing configuration file should be ignored",
			cmd: []string{
				"go", "run", "main.go",
				"-d", "#trash", "-c", "blue", "-f", "inexistent", "-r", remoteUrl,
			},
			text:             "missing configuration file should be ignored",
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
	}

	i := 1
	testCount := 0
	for _, tc := range tcs {
		if tc.expectedExitCode == 0 {
			testCount++
		}
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			exitCode := 0
			cmdFailure := false

			// modify the test text
			msg := tc.text
			if tc.text != "" {
				msg = "*[E2E test] [DIRECT CLI MODE] (" + strconv.Itoa(i) + "/" + strconv.Itoa(testCount) + ")* => " + tc.text
			}
			tc.cmd = append(tc.cmd, msg)

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

			outStr := outb.String()
			if outStr != "" {
				t.Log("cmd output")
				t.Log(outStr)
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("expected exit code does not match, got %d expected %d ", exitCode, tc.expectedExitCode)
			}

			// only increase test counter for tests that are expected to be sent
			if tc.expectedExitCode == 0 && exitCode == 0 {
				i++
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
		})
	}

	server.Stop()
}

func TestE2EFileWatchMode(t *testing.T) {

	// skip test if flag is not present
	if !rune2eTests {
		t.Log("Skipping e2e tests")
		t.Skip("Use -e2e flag to run")
	}

	//logrus.SetLevel(logrus.FatalLevel)

	tmpPath, _, err := prePateStage()
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(tmpPath)

	fmt.Println(tmpPath)

	mboxDir := tmpPath + "/mbox"

	ts := "*[E2E test] [MBOX WATCH MODE]* => "
	mailsToWrite := []string{
		ts + "mail 1",
		ts + "mail 2",
	}

	for _, mail := range mailsToWrite {
		err = writeMailToMbox(mboxDir+"/sample_mbox", mail)
		if err != nil {
			t.Fatal(err)
		}
	}

	tcs := []e2eTestCase{
		{
			name: "expect error empty message",
			cmd: []string{
				"go", "run", "main.go", "-s", "-f", tmpPath + "/config.yaml", "-v",
			},
			expectedExitCode: 0,
			expectedErrorStr: "",
		},
		//{
		//	name: "text without color",
		//	cmd: []string{
		//		"go", "run", "main.go",
		//		ts + " text without color",
		//		"-d", "#trash",
		//	},
		//	expectedExitCode: 0,
		//	expectedErrorStr: "",
		//},
		//{
		//	name: "direct Cli",
		//	cmd: []string{
		//		"go", "run", "main.go",
		//		ts + " simple color blue",
		//		"-d", "#trash", "-c", "blue", "-r", remoteUrl,
		//	},
		//	expectedExitCode: 0,
		//	expectedErrorStr: "",
		//},
		//{
		//	name: "different config file",
		//	cmd: []string{
		//		"go", "run", "main.go",
		//		ts + " nonexistent file",
		//		"-d", "#trash", "-c", "blue", "-f", "inexistent", "-r", remoteUrl,
		//	},
		//	expectedExitCode: 0,
		//	expectedErrorStr: "",
		//},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			exitCode := 0
			cmdFailure := false

			cmd := exec.Command(tc.cmd[0], tc.cmd[1:len(tc.cmd)]...)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			go func() {
				err := cmd.Run()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						exitCode = exitError.ExitCode()
						cmdFailure = true
					}
				}
			}()

			// wait for watcher to process
			time.Sleep(1000 * time.Millisecond)

			fmt.Println(errb.String())

			if err := cmd.Process.Kill(); err != nil {
				t.Fatal("failed to kill process: ", err)
			}

			t.Logf("Stdout %s", outb.String())

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

		})
	}

}

func writeMailToMbox(file string, body string) error {

	m :=
		`From www-data@amelia.com  Thu Dec 21 05:00:01 2017
From: root@amelia.com (Cron Daemon)
To: www-data@amelia.com

`
	m = m + body + "\n\n"

	f, err := os.OpenFile(file,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(m)); err != nil {

		return err
	}
	return nil
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
