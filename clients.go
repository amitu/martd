package main

type Client struct {
}

type ClientList []Client

func (c *Client) Publish(m *Message) error {
	return nil
}
