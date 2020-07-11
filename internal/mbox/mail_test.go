package mbox_test

import (
	"github.com/google/go-cmp/cmp"
	"send2slack/internal/mbox"
	"testing"
)

type TestMailFromBytesTc struct {
	name     string
	file     string
	expected *mbox.Mail
}

func TestNewMailFromBytes(t *testing.T) {

	tcs := []TestMailFromBytesTc{
		{
			name: "simple mail 1",
			file: "test-data/simple-mail-1.txt",
			expected: &mbox.Mail{
				Headers: map[string]string{
					"from": "root@amelia.com (Cron Daemon)",
					"to":   "www-data@amelia.com",
				},
				Body: "This is the email body\nwe add just some more\n\nlines\n\n:)",
			},
		},
		{
			name: "simple mail 2",
			file: "test-data/simple-mail-2.txt",
			expected: &mbox.Mail{
				Headers: map[string]string{
					"from": "root@amelia.com (Cron Daemon)",
					"to":   "www-data@amelia.com",
				},
				Body: "This email is slightly different\n\n\n\nand contains some empty lines as well as ending on empty lines\n\n\n",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			hndl, err := mbox.NewHandler(tc.file)
			if err != nil {
				t.Fatal(err)
			}

			bts := hndl.ReadLastMail()
			got := mbox.NewMailFromBytes(bts)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("Mail mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

//func TestNewMailFromStr(t *testing.T) {
//
//	filename := "test-data/simple-mail.txt"
//
//	content, err := ioutil.ReadFile(filename)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	got := mbox.NewMailFromStr(string(content))
//
//	expected := mbox.Mail{
//		Headers: map[string]string{
//			"from":"root@amelia.com (Cron Daemon)",
//			"to":"www-data@amelia.com",
//		},
//		Body: "This is the email body\nwe add just some more\nlines\n",
//	}
//
//	if diff := cmp.Diff(expected, got); diff != "" {
//		t.Errorf("Mail mismatch (-want +got):\n%s", diff)
//	}
//}
