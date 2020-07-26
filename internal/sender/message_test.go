package sender_test

import (
	"send2slack/internal/sender"
	"testing"
)

func TestMessageFromMailStr(t *testing.T) {
	type tcase struct {
		senario  string
		in       string
		template string
		expected string
	}

	tcs := []tcase{
		{
			senario: "system mail to root",
			in: `Return-Path: <root@mail.amelia.rivervps.com>
X-Original-To: root
Delivered-To: root@mail.amelia.rivervps.com
Received: by mail.amelia.rivervps.com (Postfix, from userid 0)
id C54BD7EB7; Wed,  4 Mar 2020 23:47:34 +0100 (CET)
Auto-Submitted: auto-generated
Subject: =?utf-8?q?apt-listchanges=3A_news_for_amelia?=
To: root@mail.amelia.rivervps.com
MIME-Version: 1.0
Content-Type: text/plain; charset="utf-8"
Content-Transfer-Encoding: 7bit
Message-Id: <20200304224734.C54BD7EB7@mail.amelia.rivervps.com>
Date: Wed,  4 Mar 2020 23:47:34 +0100 (CET)
From: root@mail.amelia.rivervps.com (root)

this ist the message body
and another lines 

signature

`,
			expected: "from|root@mail.amelia.rivervps.com (root)|to|root@mail.amelia.rivervps.com|this ist the message body\nand another lines \n\nsignature\n\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.senario, func(t *testing.T) {
			out, err := sender.NewMessageFromMailStr(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			// create a composite string with meta values and text to make sure both are parsed correctly
			cmprStr := "from|" + out.Meta["from"] + "|to|" + out.Meta["to"] + "|" + out.Text

			if cmprStr != tc.expected {
				t.Errorf("the message got does not match expected, got: \"%s\" expected: \"%s\"", cmprStr, tc.expected)
			}
		})
	}
}

type valdationScenatio struct {
	name        string
	msg         sender.Message
	expectError string
}

func TestMessage_Validate(t *testing.T) {

	tcs := []valdationScenatio{
		{
			name:        "empty text",
			msg:         sender.Message{},
			expectError: sender.EmptyBodyError,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()

			// expecting invalid messages
			if tc.expectError != "" {
				if err == nil {
					t.Fatal("expecting an error but not got ")
				}
				if err.Error() != tc.expectError {
					t.Errorf("error message mismatch, got: %s expected %s", err.Error(), tc.expectError)
				}

			} else {
				// expected a valid message
				if err != nil {
					t.Error(err)
				}
			}
		})
	}

}
