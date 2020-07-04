package send2slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type client struct {
	Url string
}

type ClientConfig struct {
	Url string
}

func NewClient(cfg ClientConfig) (*client, error) {

	if cfg.Url == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}

	c := client{
		Url: cfg.Url,
	}

	return &c, nil
}

func (c *client) Send(msg Message) error {

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.Url, bytes.NewBuffer(jsonMsg))
	//req.Header.Set("X-Custom-Header", "myvalue")
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
