package send2slack

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"strings"
)

const Version = "0.1.2"

type slackmsg struct {
	client  *slack.Client
	Channel string
	Text    string
	Detail  string
	color   string
	att     *slack.Attachment
}

func NewSlackMsg(token string) *slackmsg {
	sl := slackmsg{
		client: slack.New(token),
	}
	return &sl
}

func (c *slackmsg) Color(color string) {
	switch color {
	case "red":
		c.color = "#FF5640"
		break
	case "green":
		c.color = "#2eb886"
		break
	case "orange":
		c.color = "#FF9C40"
		break
	case "blue":
		c.color = "#3270A5"
		break
	case "lime":
		c.color = "#B3EE3C"
		break
	default:
		c.color = color
	}
}

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
	Headers []string
}

func parseMailString(in string) *Mail {
	//	mail:=`
	//
	//MIME-Version: 1.0
	//Content-Type: text/plain; charset=UTF-8
	//Content-Transfer-Encoding: 8bit
	//X-Cron-Env: <SHELL=/bin/sh>
	//X-Cron-Env: <HOME=/home/vagrant>
	//X-Cron-Env: <PATH=/usr/bin:/bin>
	//X-Cron-Env: <LOGNAME=vagrant>
	//test from cron
	//`
	m := Mail{}

	scanner := bufio.NewScanner(strings.NewReader(in))
	// get the headers
	for scanner.Scan() {
		s := scanner.Text()
		if strings.TrimSpace(s) == "" {
			break
		}

		search := "From:"
		if strings.HasPrefix(s, search) {
			m.From = strings.TrimSpace(s[len(search):])
			continue
		}

		search = "To: "
		if strings.HasPrefix(s, search) {
			m.To = strings.TrimSpace(s[len(search):])
			continue
		}

		search = "Subject: "
		if strings.HasPrefix(s, search) {
			m.Subject = strings.TrimSpace(s[len(search):])
			continue
		}

		m.Headers = append(m.Headers, s)

	}
	for scanner.Scan() {
		s := scanner.Text()
		m.Body = m.Body + s + "\n"
	}

	m.Body = strings.TrimSpace(m.Body)

	return &m

}

func (c *slackmsg) send() error {
	_, _, err := c.client.PostMessage(c.Channel, slack.MsgOptionText(c.Text, false), slack.MsgOptionAttachments(*c.att))
	if err != nil {
		return errors.New(fmt.Sprintf("error sending slack message: %s\n", err))
	}
	return nil
}

func (c *slackmsg) SendMail() error {
	if c.Detail == "" {
		return errors.New("unable to send empty message")
	}
	mail := parseMailString(c.Detail)

	c.Text = `*[sendmail]* from: _"` + mail.From + `"_ Subject: _"` + mail.Subject + `"_
` + "```" + mail.Body + "```" + `
`
	//c.Detail = `*Details:*
	//From: `+mail.From+`
	//To: `+mail.To+`
	//Subject: `+mail.Subject+`
	//`

	c.Detail = ""
	c.Color("blue")

	c.att = &slack.Attachment{
		Text:  c.Detail,
		Color: c.color,
	}
	return c.send()
}

func (c *slackmsg) SendMsg() error {
	if c.Text == "" && c.Detail == "" {
		return errors.New("unable to send empty message")
	}

	if c.color != "" || c.Detail != "" {
		c.Detail = c.Text + c.Detail
		c.Text = ""
	}

	c.att = &slack.Attachment{
		Text:  c.Detail,
		Color: c.color,
	}

	return c.send()
}
