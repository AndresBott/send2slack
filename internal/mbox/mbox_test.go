package mbox_test

import (
	"io/ioutil"
	"os"
	"send2slack/internal/mbox"
	"strings"
	"testing"
)

type TestHandlerTc struct {
	name           string
	emailCount     int
	mboxFilename   string
	expectedString string
}

func TestHandler(t *testing.T) {

	tcs := []TestHandlerTc{
		{
			name:           "read 3 mails",
			mboxFilename:   "test-data/three-mail.mbox",
			emailCount:     3,
			expectedString: "|Email 3|Email 2|Email 1",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			hndl, err := mbox.NewHandler(tc.mboxFilename)
			if err != nil {
				t.Fatal(err)
			}
			defer hndl.Close()

			cmpString := ""
			mails := []mbox.Mail{}
			for hndl.HasMails() {
				mailBytes := hndl.ReadLastMail()
				mail := mbox.NewMailFromBytes(mailBytes)
				mails = append(mails, *mail)

				cmpString = cmpString + "|" + strings.Trim(strings.TrimSpace(mail.Body), "\n")
			}

			if len(mails) != tc.emailCount {
				t.Errorf("expected mail count does not match, got: %d expected: %d", len(mails), tc.emailCount)
			}

			if cmpString != tc.expectedString {
				t.Errorf("expected mail string does not match, got: \"%s\" expected: \"%s\"", cmpString, tc.expectedString)
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

func TestHandler_DeleteLastMail(t *testing.T) {

	// Prepare the stage to start testing
	dir, err := ioutil.TempDir("/tmp", "mbox_handler")
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(dir)
	defer os.RemoveAll(dir)

	mboxFile := dir + "/sample_mbox"

	mailsToWrite := []string{
		"mail 1",
		"mail 2",
		"mail 3",
		"mail 4",
		"mail 5",
	}

	for _, mail := range mailsToWrite {
		err = writeMailToMbox(mboxFile, mail)
		if err != nil {
			t.Fatal(err)
		}
	}

	// now we can star testing

	hndl, err := mbox.NewHandler(mboxFile)
	if err != nil {
		t.Fatal(err)
	}
	defer hndl.Close()

	// count al emails
	i := 0
	for hndl.HasMails() {
		_ = hndl.ReadLastMail()
		i++
	}
	if i != len(mailsToWrite) {
		t.Errorf("unexpected lenght in mmbox file, got: %d expected: %d", i, len(mailsToWrite))
	}
	hndl.Reset()

	// remove the last email
	_, err = hndl.ConsumeLastMail()
	if err != nil {
		t.Fatal(err)
	}
	hndl.Reset()

	// iterate again over the resulting mails
	i = 0
	mailStr := ""
	for hndl.HasMails() {

		m := mbox.NewMailFromBytes(hndl.ReadLastMail())

		mailStr = mailStr + "|" + strings.Trim(strings.TrimSpace(m.Body), "\n")
		i++
	}
	expectedLen := len(mailsToWrite) - 1

	if i != expectedLen {
		t.Errorf("unexpected lenght in mmbox file, got: %d expected: %d", i, expectedLen)
	}

	expectedMailStr := "|mail 4|mail 3|mail 2|mail 1"

	if mailStr != expectedMailStr {
		t.Errorf("expected mail string does not match, got: \"%s\" expected: \"%s\"", mailStr, expectedMailStr)
	}

}
