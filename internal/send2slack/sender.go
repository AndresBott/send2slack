package send2slack

type MessageSender interface {
	SendMessage(msg *Message) error
}

type DummyMessageSender struct {
	Msg string
}

func (sndr *DummyMessageSender) SendMessage(msg *Message) error {
	sndr.Msg = sndr.Msg + "|" + msg.Text
	return nil
}
