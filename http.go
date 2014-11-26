package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/amitu/gutils"
	"github.com/dustin/go-humanize"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type ChanResponse struct {
	Etag    int64    `json:"etag"`
	Payload []string `json:"payload"`
}

type SubResponse struct {
	Channels map[string]*ChanResponse `json:"channels,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

var (
	HostPort    string
	Debug       bool
	ServerStart time.Time
	nSub        int
	CIDM_lock   sync.RWMutex
)

func init() {
	flag.StringVar(&HostPort, "http", ":54321", "HTTP Host:Port")
	flag.BoolVar(&Debug, "debug", false, "Debug.")
	ServerStart = time.Now()
}

func reject(w http.ResponseWriter, reason string) {
	j, err := json.Marshal(SubResponse{Error: reason})
	if err != nil {
		log.Println("Error during json.Marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, string(j), http.StatusBadRequest)
}

func respond(w http.ResponseWriter, resp *SubResponse) {
	j, err := json.Marshal(resp)
	if err != nil {
		log.Println("Error during json.Marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func PubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		reject(w, err.Error())
		return
	}

	channel := r.FormValue("channel") // TODO: support multiple channels
	size_s := r.FormValue("size")
	life_s := r.FormValue("life")
	one2one := r.FormValue("one2one") == "true"
	key := r.FormValue("key")

	if channel == "" {
		reject(w, "channel is required")
		return
	}

	size := uint(10)
	if size_s != "" {
		_, err := fmt.Sscan(size_s, &size)
		if err != nil {
			reject(w, "invalid size: "+err.Error())
			return
		}
	}

	life := time.Second * 10
	if life_s != "" {
		_, err := fmt.Sscan(life_s, &life)
		if err != nil {
			reject(w, "invalid life: "+err.Error())
			return
		}
	}

	ch, err := GetOrCreateChannel(channel, size, life, one2one, key)
	if err != nil {
		reject(w, err.Error())
		return
	}

	if len(body) != 0 {
		err := ch.Pub(body)
		if err != nil {
			reject(w, err.Error())
			return
		}
	}

	j, err := ch.Json()
	if err != nil {
		reject(w, err.Error())
		return
	}

	fmt.Fprintf(w, "%s", j)
}

func IncrCCount(by int) {
	CIDM_lock.Lock()
	defer CIDM_lock.Unlock()

	nSub += by
}

func SubHandler(w http.ResponseWriter, r *http.Request) {
	IncrCCount(1)
	defer IncrCCount(-1)

	r.ParseForm()
	evch := make(chan *ChannelEvent)

	subs := make([]*Channel, 0)
	resp := &SubResponse{make(map[string]*ChanResponse), ""}

	for k, _ := range r.Form {
		if k == "cid" {
			continue
		}
		v := r.FormValue(k)
		if v == "" {
			reject(w, k+" has no etag")
			return
		}

		etag := int64(0)
		_, err := fmt.Sscan(v, &etag)
		if err != nil {
			reject(w, "invalid etag: "+err.Error())
			return
		}

		ch := GetChannel(k)
		has, ith := ch.HasNew(etag)
		if has {
			ch.Append(resp, ith)
		} else {
			subs = append(subs, ch)
		}
	}

	if len(resp.Channels) != 0 {
		respond(w, resp)
		return
	}

	// sub everything
	for _, ch := range subs {
		ch.Sub(evch)
	}

	cner, ok := w.(http.CloseNotifier)
	if !ok {
		reject(w, "server issue, handler does not support CloseNotifier")
		return
	}

	select {
	case cm := <-evch:
		resp.Channels[cm.Chan.Name] = &ChanResponse{
			cm.Mesg.Created, []string{string(cm.Mesg.Data)},
		}
		respond(w, resp)
	case <-cner.CloseNotify():
	}

	for _, ch := range subs {
		ch.UnSub(evch)
	}
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	CIDM_lock.Lock()
	defer CIDM_lock.Unlock()

	ChannelLock.Lock()
	defer ChannelLock.Unlock()

	fmt.Fprintf(
		w, `Started: %s (%s)
Channels: %d
Clients: %d
NumGoroutine: %d
`,
		ServerStart,
		humanize.Time(ServerStart),
		len(Channels),
		nSub,
		runtime.NumGoroutine(),
	)
}

func ServeHTTP() {
	http.HandleFunc("/pub", PubHandler)
	http.HandleFunc("/sub", SubHandler)
	http.HandleFunc("/stats", StatsHandler)
	http.Handle("/", http.FileServer(FS(Debug)))

	log.Printf("Started HTTP Server on %s.", HostPort)
	logger := gutils.NewApacheLoggingHandler(http.DefaultServeMux, os.Stderr)
	fmt.Println(logger)
	server := &http.Server{
		Addr:    HostPort,
		Handler: http.DefaultServeMux, // logger,
	}
	log.Fatal(server.ListenAndServe())
}
