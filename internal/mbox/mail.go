package mbox

import (
	"bytes"
	"strings"
)

type Mail struct {
	Headers map[string]string
	Body    string
}

// NewMailFromBytes takes an slice of slice of byte as input and composes a mail struct
func NewMailFromBytes(input [][]byte) *Mail {
	m := Mail{
		Headers: map[string]string{},
		Body:    "",
	}

	isBody := false
	var sb strings.Builder

	inputLen := len(input)
	lastLine := inputLen - 1
	for i := 0; i < inputLen; i++ {
		if !isBody {
			// headers
			if tr := bytes.TrimSpace(input[i]); len(tr) == 0 {
				// starting from here we are dealing with the body part
				isBody = true
				continue
			}
			h := string(bytes.TrimSpace(input[i]))
			splitHeader := strings.SplitN(h, ":", 2)
			if len(splitHeader) <= 1 {
				continue
			}
			m.Headers[strings.ToLower(strings.TrimSpace(splitHeader[0]))] = strings.TrimSpace(splitHeader[1])
		} else {
			// body
			sb.Write(input[i])
			if i != lastLine {
				sb.WriteString("\n")
			}
		}
	}

	body := sb.String()
	// remove last \n from mails
	if len(body) > 0 && body[len(body)-1:] == "\n" {
		body = body[0 : len(body)-1]
	}
	m.Body = body
	return &m
}
