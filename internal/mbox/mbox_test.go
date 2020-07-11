package mbox_test

import (
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
