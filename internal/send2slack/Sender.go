package send2slack

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
)

type Sender struct {
	client *slack.Client
}

func NewSender(token string) *Sender {
	sl := Sender{
		client: slack.New(token),
	}
	return &sl
}

// SendMessage uses the slack library to send a message
func (c *Sender) SendMessage(msg *Message) error {

	if msg.Text == "" && msg.Detail == "" {
		return errors.New("unable to send empty message")
	}

	if msg.color != "" || msg.Detail != "" {
		msg.Detail = msg.Text + msg.Detail
		msg.Text = ""
	}

	msg.Att = &slack.Attachment{
		Text:  msg.Detail,
		Color: msg.color,
	}

	_, _, err := c.client.PostMessage(msg.Channel, slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAttachments(*msg.Att))
	if err != nil {
		return fmt.Errorf("error sending slack message: %s\n", err)
	}

	return nil

}
