package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Data    []byte
	Created int64 // created time acts as the etag
}

type Channel struct {
	Name     string                       `json:"name"`
	Size     uint                         `json:"size"`
	Life     time.Duration                `json:"life"`
	Key      string                       `json:"key,omitempty"`
	Clients  map[string]chan ChannelEvent `json:"-"`
	Messages *CircularMessageArray        `json:"-"`
	One2One  bool                         `json:"one2one"`
	lock     sync.RWMutex                 `json:"-"`
	inited   bool
}

type ChannelEvent struct {
	Chan *Channel
	Mesg *Message
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
		ch.Life = life
		ch.One2One = one2one
		ch.Key = key
		ch.Messages = NewCircularMessageArray(size)
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
		ch = &Channel{Name: name, Clients: make(map[string]chan ChannelEvent)}
		Channels[name] = ch

		// TODO spawn a goroutine to delete this channel?
	}
	return ch
}

func (c *Channel) Pub(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	m := &Message{Data: data, Created: time.Now().UnixNano()}
	c.Messages.Push(m)

	for _, evch := range c.Clients {
		evch <- ChannelEvent{c, m}
	}

	c.Clients = make(map[string]chan ChannelEvent)

	fmt.Println("After publish", *c.Messages)
	return nil
}

func (c *Channel) HasNew(cid string, etag int64) (bool, uint) {
	/*
		etag symantics: if someone has passed etag != 0, means they have some
		old data, and want everything since then. we may have lost some data
		by then, but we should not lose anything more.
	*/
	c.lock.Lock()
	defer c.lock.Unlock()

	if etag != 0 && c.Messages != nil && c.Messages.Length() > 0 {
		oldest, _ := c.Messages.PeekOldest() // TODO, handle error?
		if oldest.Created > etag {
			return true, 0 // oldest
		}

		ml := c.Messages.Length()

		// find the first message in the channel with .Created == etag.
		for i := uint(0); i < ml-1; i++ {
			ith, _ := c.Messages.Ith(i)
			if etag == ith.Created {
				return true, i + 1
			}
		}
	}
	return false, 0
}

func (c *Channel) Sub(cid string, evch chan ChannelEvent) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Clients[cid] = evch
}

func (c *Channel) UnSub(cid string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.Clients, cid)
}

func (c *Channel) Json() ([]byte, error) {
	m, _ := c.Messages.PeekNewest() // TODO handle error
	return json.MarshalIndent(
		map[string]int64{"etag": m.Created}, " ", "    ",
	)
}

func (ch *Channel) Append(resp *SubResponse, ith uint) {
	payload := []string{}
	etag := int64(0)
	ml := ch.Messages.Length()
	for i := ith; i < ml; i++ {
		ithm, _ := ch.Messages.Ith(i)
		payload = append(payload, string(ithm.Data))
		etag = ithm.Created
	}
	resp.Channels[ch.Name] = &ChanResponse{etag, payload}
}
