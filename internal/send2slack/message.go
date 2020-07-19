package send2slack

import (
	"bufio"
	"errors"
	"strings"
)

type Message struct {
	origin      string
	Destination string
	Text        string
	Color       string
	Debug       bool
	Meta        map[string]string
}

type Email struct {
	Headers map[string]string
	Body    string
}

// set the color of the message attachment field
func (m *Message) getColor() string {
	switch m.Color {
	case "red":
		return "#FF5640"
	case "green":
		return "#2eb886"
	case "orange":
		return "#FF9C40"
	case "blue":
		return "#3270A5"
	case "lime":
		return "#B3EE3C"
	default:
		return m.Color
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
func NewMessageFromMailStr(in string) (*Message, error) {

	m := Email{
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

	//// check that we don't have a empty body
	//if len(m.Body) <= 20 {
	//	s := strings.TrimSpace(m.Body)
	//	if len(s) == 0 {
	//		return nil, errors.New("message body cannot be empty")
	//	}
	//}

	return NewMessageFromMail(m)

}

// NewMessageFromMail parses an Email struct and returns a s2s Message
func NewMessageFromMail(m Email) (*Message, error) {

	msg := Message{
		Meta:   m.Headers,
		Text:   m.Body,
		origin: "email",
	}

	// check for a header "channel"
	if c := getMapString(m.Headers, "x-slack-channel"); c != "" {
		msg.Destination = c
	}

	// check for a header "color"
	if c := getMapString(m.Headers, "color"); c != "" {
		msg.Color = c
	} else if c := getMapString(m.Headers, "x-slack-color"); c != "" {
		msg.Color = c
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
