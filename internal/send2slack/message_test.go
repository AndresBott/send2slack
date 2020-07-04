package send2slack_test

import (
	"send2slack/internal/send2slack"
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
			expected: "from:root@mail.amelia.rivervps.com (root)|original:root|body:this ist the message body\nand another lines \n\nsignature\n\n",
			template: `from:{{- index .Headers "from" }}|original:{{- index .Headers "x-original-to" }}|body:{{ .Body }}`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.senario, func(t *testing.T) {
			out, err := send2slack.NewMessageFromMailStr(tc.in, tc.template)
			if err != nil {
				t.Fatal(err)
			}

			if out.Text != tc.expected {
				t.Errorf("the message got does not match expected, got: \"%s\" expected: \"%s\"", out.Text, tc.expected)
			}

		})
	}
}

type valdationScenatio struct {
	name        string
	msg         send2slack.Message
	expectError string
}

func TestMessage_Validate(t *testing.T) {

	tcs := []valdationScenatio{
		{
			name:        "empty text",
			msg:         send2slack.Message{},
			expectError: send2slack.EmptyBodyError,
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
