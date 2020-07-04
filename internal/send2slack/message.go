package send2slack

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/slack-go/slack"
	"strings"
	"text/template"
)

type Message struct {
	Channel string
	Text    string
	Detail  string
	color   string
	Debug   bool
	Att     *slack.Attachment
}

// set the color of the message attachment field
func (m *Message) Color(color string) {
	switch color {
	case "red":
		m.color = "#FF5640"
		break
	case "green":
		m.color = "#2eb886"
		break
	case "orange":
		m.color = "#FF9C40"
		break
	case "blue":
		m.color = "#3270A5"
		break
	case "lime":
		m.color = "#B3EE3C"
		break
	default:
		m.color = color
	}
}

const (
	EmptyBodyError = "text cannot be empty"
)

// validates if the message fulfils the minimal requirement to be sent
func (m *Message) Validate() error {

	if m.Text == "" {
		return errors.New(EmptyBodyError)
	}

	return nil
}

// takes a string input that is expected to be a plain text email and transforms this to a slack message
// using a template to from the main slack message
// example template usage:
// {{- index .Headers "from" }} to access any of the headers
//
func NewMessageFromMailStr(in string, tpl string) (*Message, error) {

	type email struct {
		Headers map[string]string
		Body    string
	}

	m := email{
		Headers: map[string]string{},
	}

	scanner := bufio.NewScanner(strings.NewReader(in))
	// get the headers
	for scanner.Scan() {
		s := scanner.Text()

		// exit the head section on the first empty line
		if strings.TrimSpace(s) == "" {
			break
		}
		s = strings.Trim(s, "\n")

		splited := strings.SplitN(s, ":", 2)
		// slit character not found
		if len(splited) <= 1 {
			continue
		}
		m.Headers[strings.ToLower(strings.TrimSpace(splited[0]))] = strings.TrimSpace(splited[1])
	}

	var sb strings.Builder
	for scanner.Scan() {
		s := scanner.Text()
		sb.WriteString(s + "\n")
	}
	m.Body = sb.String()

	// check that we don't have a empty body
	if len(m.Body) <= 20 {
		s := strings.TrimSpace(m.Body)
		if len(s) == 0 {
			return nil, errors.New("message body cannot be empty")
		}
	}

	msg := Message{}

	tmpl, err := template.New("test").Parse(tpl)
	if err != nil {
		return nil, err
	}
	var tplOut bytes.Buffer
	err = tmpl.Execute(&tplOut, m)
	if err != nil {
		return nil, err
	}

	msg.Text = tplOut.String()

	//	msg.Text =
	//`*[sendmail]* from: _"` + getMapString(headers,"from") + `"_ Subject: _"` + getMapString(headers,"subject") + `"_
	//` + "```" + strings.Join(body,"\n") + "```"

	msg.Detail = ""

	// check for a header "channel"
	if c := getMapString(m.Headers, "x-slack-channel"); c != "" {
		msg.Channel = c
	}

	// check for a header "color"
	if c := getMapString(m.Headers, "color"); c != "" {
		msg.Color(c)
	} else if c := getMapString(m.Headers, "x-slack-color"); c != "" {
		msg.Color(c)
	} else {
		msg.Color("blue")
	}

	return &msg, nil

}

// getMapString looks int the provided map if the key exists and returns it
// it returns an empty string if it does not exist
func getMapString(m map[string]string, key string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}
