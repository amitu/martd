package main

import (
	"sync"
	"time"
)

type Message struct {
	Data    []byte
	Created time.Time // created time acts as the etag
}

type Channel struct {
	Name     string
	Size     uint
	Life     time.Duration
	Key      string
	Clients  ClientList
	Messages *CircularMessageArray
	One2One  bool
	lock     sync.RWMutex
}

func NewChannel(name string, size uint) *Channel {
	return &Channel{
		Name:     name,
		Size:     size,
		Messages: NewCircularMessageArray(size),
		Life:     time.Second * 10,
		One2One:  false,
	}
}

func (c *Channel) Publish(data []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	m := &Message{Data: data, Created: time.Now()}
	c.Messages.Push(m)

	for _, client := range c.Clients {
		err := client.Publish(m)
		if err != nil {
			c.RemoveClient(client)
		}
	}
}

func (c *Channel) AddClient(client Client) error {
	return nil
}

func (c *Channel) RemoveClient(client Client) error {
	return nil
}
