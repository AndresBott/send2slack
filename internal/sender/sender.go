package sender

import "strings"

type MessageSender interface {
	SendMessage(msg *Message) error
	SendError(err error)
}

type DummyMessageSender struct {
	Msg string
}

func (sndr *DummyMessageSender) SendMessage(msg *Message) error {
	s := strings.Trim(msg.Text, "\n")
	s = strings.TrimSpace(s)

	if s != "" {
		sndr.Msg = sndr.Msg + "|" + s
	}

	return nil
}
func (sndr *DummyMessageSender) SendError(err error) {
	msg := Message{
		Text:  err.Error(),
		Color: "red",
	}
	sndr.SendMessage(&msg)
}
