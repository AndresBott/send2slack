package send2slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"net/http"
	"net/url"
)

type SlackSender struct {
	client *slack.Client
	mode   Mode
	url    *url.URL
}

func NewSlackSender(cfg *Config) (*SlackSender, error) {

	if cfg.Mode == ModeClientCli {
		if cfg.URL == nil {
			return nil, fmt.Errorf("url cannot be empty")
		}
	}

	sl := SlackSender{
		client: slack.New(cfg.Token),
		mode:   cfg.Mode,
		url:    cfg.URL,
	}
	return &sl, nil
}

// SendMessage depending on the configured mode
func (c *SlackSender) SendMessage(msg *Message) error {
	switch c.mode {
	case ModeDirectCli:
		return c.sendMsgDirecCli(msg)
	case ModeClientCli:
		return c.sendMsgClientCli(msg)
	default:
		return errors.New("SlackSlackSender mode not found")
	}
}

// internal method to send a message directly using the slack api
func (c *SlackSender) sendMsgDirecCli(msg *Message) error {
	if msg.Text == "" && msg.Detail == "" {
		return errors.New("unable to send empty message")
	}

	if msg.getColor() != "" || msg.Detail != "" {
		msg.Detail = msg.Text + msg.Detail
		msg.Text = ""
	}

	msg.Att = &slack.Attachment{
		Text:  msg.Detail,
		Color: msg.getColor(),
	}

	_, _, err := c.client.PostMessage(msg.Channel, slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAttachments(*msg.Att))
	if err != nil {
		return fmt.Errorf("error sending slack message: %s\n", err)
	}

	return nil
}

// internal method to send a message to a send2slack server
func (c *SlackSender) sendMsgClientCli(msg *Message) error {
	if msg.Text == "" && msg.Detail == "" {
		return errors.New("unable to send empty message")
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url.String(), bytes.NewBuffer(jsonMsg))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("message not submitted")
	}

	return nil
}
