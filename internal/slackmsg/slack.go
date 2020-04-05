package slackmsg

import (
	"fmt"
	"github.com/slack-go/slack"
)

type slackmsg struct {
	client  *slack.Client
	Channel string
}

func NewSlackMsg(token string) *slackmsg {
	sl := slackmsg{
		client: slack.New(token),
	}

	return &sl
}

func (c *slackmsg) SendMsg(text string) {
	attachment := slack.Attachment{
		//Pretext: "some pretext",
		//Text:    "some text",
		// Uncomment the following part to send a field too

		//Fields: []slack.AttachmentField{
		//	slack.AttachmentField{
		//		Title: "a",
		//		Value: "no",
		//	},
		//},

	}

	channelID, timestamp, err := c.client.PostMessage(c.Channel, slack.MsgOptionText(text, false), slack.MsgOptionAttachments(attachment))

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

}

func (c *slackmsg) Warn(text string, detail string) {
	attachment := slack.Attachment{
		//Pretext: "some pretext",
		Text: detail,
		// Uncomment the following part to send a field too

		//Fields: []slack.AttachmentField{
		//	slack.AttachmentField{
		//		Title: "a",
		//		Value: "no",
		//	},
		//},

	}

	channelID, timestamp, err := c.client.PostMessage(c.Channel, slack.MsgOptionText(text, false), slack.MsgOptionAttachments(attachment))

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

}
