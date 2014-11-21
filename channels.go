package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Data    []byte
	Created time.Time // created time acts as the etag
}

type Channel struct {
	Name     string                `json:"name"`
	Size     uint                  `json:"size"`
	Life     time.Duration         `json:"life"`
	Key      string                `json:"key,omitempty"`
	Clients  ClientList            `json:"-"`
	Messages *CircularMessageArray `json:"-"`
	One2One  bool                  `json:"one2one"`
	lock     sync.RWMutex          `json:"-"`
}

var (
	Channels    map[string]*Channel
	ChannelLock sync.RWMutex
)

func init() {
	Channels = make(map[string]*Channel)
}

func NewChannel(
	name string, size uint, life time.Duration, one2one bool, key string,
) (*Channel, error) {
	ChannelLock.Lock()
	defer ChannelLock.Unlock()

	ch, ok := Channels[name]
	if !ok {
		ch = &Channel{
			Name:     name,
			Size:     size,
			Messages: NewCircularMessageArray(size),
			Life:     life,
			One2One:  one2one,
			Key:      key,
		}
		Channels[name] = ch

		// Spaws a goroutine to delete this channel?
	}

	j, _ := json.MarshalIndent(Channels, " ", "    ")
	fmt.Println("After New", string(j))
	return ch, nil
}

func (c *Channel) Publish(data []byte) error {
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

	fmt.Println("After publish", *c.Messages)
	return nil
}

func (c *Channel) AddClient(client Client) error {
	return nil
}

func (c *Channel) RemoveClient(client Client) error {
	return nil
}

func (c *Channel) Json() ([]byte, error) {
	return json.MarshalIndent(c, " ", "    ")
}
