package main

import (
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/amitu/gutils"
)

type ChanResponse struct {
	Etag    string    `json:"etag"`
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
	CIDM_lock   sync.RWMutex
	nSub        = expvar.NewInt("nSub")
	nList        = expvar.NewInt("nList")
	nSubAll     = expvar.NewInt("nSubAll")
	nPubAll     = expvar.NewInt("nPubAll")
)

func init() {
	flag.StringVar(&HostPort, "http", ":54321", "HTTP Host:Port")
	flag.BoolVar(&Debug, "debug", false, "Debug.")
	ServerStart = time.Now()

	expvar.Publish("stats", expvar.Func(stats))
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
	nPubAll.Add(1)

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

	life := time.Second * 60 * 60 // default expiry = one hr
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

	etag := int64(0)

	if len(body) != 0 {
		etag = ch.Pub(body)
	}

	j, err := json.MarshalIndent(
		map[string]string{"etag": fmt.Sprintf("%d", etag)}, " ", "    ",
	)

	if err != nil {
		reject(w, err.Error())
		return
	}

	fmt.Fprintf(w, "%s", j)
}

func SubHandler(w http.ResponseWriter, r *http.Request) {
	nSub.Add(1)
	nSubAll.Add(1)
	defer nSub.Add(-1)

	r.ParseForm()
	evch := make(chan *ChannelEvent)

	subs := make([]*Channel, 0)
	resp := &SubResponse{make(map[string]*ChanResponse), ""}

	for k := range r.Form {
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
			fmt.Sprintf("%d", cm.Mesg.Created), []string{string(cm.Mesg.Data)},
		}
		respond(w, resp)
	case <-cner.CloseNotify():
	}

	for _, ch := range subs {
		ch.UnSub(evch)
	}
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	nList.Add(1)
	DumpChannels()

	w.Write([]byte("ok\n"))
}

func ServeHTTP() {
	http.HandleFunc("/list", ListHandler)
	http.HandleFunc("/pub", PubHandler)
	http.HandleFunc("/sub", SubHandler)
	http.Handle("/", http.FileServer(FS(Debug)))

	log.Printf("Started HTTP Server on %s.", HostPort)
	logger := gutils.NewApacheLoggingHandler(http.DefaultServeMux, os.Stderr)
	server := &http.Server{
		Addr: HostPort,
		Handler:/*http.DefaultServeMux,*/ logger,
	}
	log.Fatal(server.ListenAndServe())
}
