package send2slack

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
)

const Version = "0.1.1"

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

func translateColor(color string) string {
	switch color {
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
	}

	return color
}

func (c *slackmsg) SendMsg(text string, detail string, color string) error {

	if text == "" && detail == "" {
		return errors.New("unable to send empty message")
	}

	if color != "" || detail != "" {
		detail = text + detail
		text = ""
	}

	color = translateColor(color)

	attachment := slack.Attachment{
		//Pretext: "some pretext",
		Text:  detail,
		Color: color,
		//AuthorName: "Bobby Tables",
		//AuthorLink: "http://flickr.com/bobby/",
		//AuthorIcon: "https://img.icons8.com/material/4ac144/256/user-male.png",
		//ImageURL: "https://andresbott.com/AndresBott_Silk.jpg",
		//ThumbURL: "https://andresbott.com/AndresBott_Silk.jpg",

		// Uncomment the following part to send a field too

		//Fields: []slack.AttachmentField{
		//	slack.AttachmentField{
		//		Title: "a",
		//		Value: "no",
		//	},
		//},

	}
	_, _, err := c.client.PostMessage(c.Channel, slack.MsgOptionText(text, false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		return errors.New(fmt.Sprintf("error sending slack message: %s\n", err))
	}
	return nil
}
