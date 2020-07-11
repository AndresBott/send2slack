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

	m.Body = sb.String()
	return &m
}

// TODO: uncomment and solve if needed in the future
//func NewMailFromStr( input string)*Mail  {
//	m := Mail{
//		Headers: map[string]string{},
//		Body: "",
//	}
//
//	scanner := bufio.NewScanner(strings.NewReader(input))
//	// get the Headers
//	for scanner.Scan() {
//		s := scanner.Text()
//
//		// exit the head section on the first empty line
//		if strings.TrimSpace(s) == "" {
//			break
//		}
//		s = strings.Trim(s, "\n")
//
//		splited := strings.SplitN(s, ":", 2)
//		// slit character not found
//		if len(splited) <= 1 {
//			continue
//		}
//		m.Headers[strings.ToLower(strings.TrimSpace(splited[0]))] = strings.TrimSpace(splited[1])
//	}
//
//	var sb strings.Builder
//	for scanner.Scan() {
//		s := scanner.Text()
//		sb.WriteString(s + "\n")
//	}
//	m.Body = sb.String()
//
//	return &m
//}
