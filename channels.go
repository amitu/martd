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
	Name     string                   `json:"name"`
	Size     uint                     `json:"size"`
	Life     time.Duration            `json:"life"`
	Key      string                   `json:"key,omitempty"`
	Clients  map[string]chan *Message `json:"-"`
	Messages *CircularMessageArray    `json:"-"`
	One2One  bool                     `json:"one2one"`
	lock     sync.RWMutex             `json:"-"`
	inited   bool
}

var (
	Channels    map[string]*Channel
	ChannelLock sync.RWMutex
)

func init() {
	Channels = make(map[string]*Channel)
}

func GetOrCreateChannel(
	name string, size uint, life time.Duration, one2one bool, key string,
) (*Channel, error) {
	ChannelLock.Lock()
	defer ChannelLock.Unlock()

	ch := GetChannel_(name)

	if !ch.inited {
		ch.inited = true
		ch.Size = size
		ch.Messages = NewCircularMessageArray(size)
		ch.Life = life
		ch.One2One = one2one
		ch.Key = key
	}

	j, _ := json.MarshalIndent(Channels, " ", "    ")
	fmt.Println("After New", string(j))
	return ch, nil
}

func GetChannel(name string) *Channel {
	ChannelLock.Lock()
	defer ChannelLock.Unlock()
	return GetChannel_(name)
}

func GetChannel_(name string) *Channel {
	ch, ok := Channels[name]
	if !ok {
		ch = &Channel{Name: name, Clients: make(map[string]chan *Message)}
		Channels[name] = ch

		// Spaws a goroutine to delete this channel?
	}
	return ch
}

func (c *Channel) Publish(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	m := &Message{Data: data, Created: time.Now()}
	c.Messages.Push(m)

	for _, client := range c.Clients {
		client <- m
	}

	fmt.Println("After publish", *c.Messages)
	return nil
}

func (c *Channel) Subscribe(cid string) chan *Message {
	existing, ok := c.Clients[cid]
	if ok {
		existing <- nil
	}
	ch := make(chan *Message)
	c.Clients[cid] = ch
	return ch
}

func (c *Channel) UnSubscribe(cid string) {
	delete(c.Clients, cid)
}

func (c *Channel) Json() ([]byte, error) {
	return json.MarshalIndent(c, " ", "    ")
}
