package send2slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"net/http"
	"net/url"
	"text/template"
)

type SlackSender struct {
	// todo add destination for error sending
	client        *slack.Client
	mode          Mode
	url           *url.URL
	emailTemplate string
}

type slackMessage struct {
	Message
	att *slack.Attachment
}

func NewSlackSender(cfg *Config) (*SlackSender, error) {

	if cfg.Mode == ModeHttpClient {
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
	err := msg.Validate()
	if err != nil {
		return err
	}

	switch c.mode {
	case ModeDirectCli:

		slkMsg, err := c.transformMsg(msg)
		if err != nil {
			return err
		}
		return c.sendMsgDirecCli(slkMsg)

	case ModeHttpClient:
		return c.sendMsgHttpClient(msg)

	default:
		return errors.New("SlackSlackSender mode not found")
	}
}

// SendError send an error to the default destination
func (c *SlackSender) SendError(err error) {
	msg := Message{
		Text:  err.Error(),
		Color: "red",
	}
	c.SendMessage(&msg)
}

// default template used to generate the slack message based on an email
const DefaultMailTemplate = `*[EMAIL]* from: _{{ index .Meta "from" }}_ ` + "```" + `{{ .Text }}` + "```"

func (c *SlackSender) transformMsg(msg *Message) (*slackMessage, error) {

	slkMsg := slackMessage{
		att: &slack.Attachment{},
	}

	switch msg.origin {
	case "mail":
		if c.emailTemplate == "" {
			c.emailTemplate = DefaultMailTemplate
		}

		tmpl, err := template.New("test").Parse(c.emailTemplate)
		if err != nil {
			return nil, err
		}
		var tplOut bytes.Buffer
		err = tmpl.Execute(&tplOut, msg)
		if err != nil {
			return nil, err
		}

		slkMsg.Text = tplOut.String()
		slkMsg.Color = ""
		slkMsg.att = nil

		break
	default:

		// if color is defined send the message as attachment
		if msg.getColor() != "" {
			slkMsg.att.Text = msg.Text
			slkMsg.att.Color = msg.getColor()
			msg.Text = ""
		}

		break
	}

	return &slkMsg, nil

}

// internal method to send a message directly using the slack api
func (c *SlackSender) sendMsgDirecCli(msg *slackMessage) error {

	_, _, err := c.client.PostMessage(msg.Destination, slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionAttachments(*msg.att))
	if err != nil {
		return fmt.Errorf("error sending slack message: %s\n", err)
	}

	return nil
}

// internal method to send a message to a send2slack server
func (c *SlackSender) sendMsgHttpClient(msg *Message) error {

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url.String(), bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}

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
