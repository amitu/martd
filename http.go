package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/amitu/gutils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// type ChanResponse struct {
// 	Etag    int64    `json:"etag"`
// 	Payload []string `json:"payload"`
// }

type SubResponse struct {
	// Channels map[string]Channel `json:"channels"`
	Etag    int64    `json:"etag"`
	Payload []string `json:"payload"`
	Error   string   `json:"error"`
}

var (
	HostPort string
	Debug    bool
)

func init() {
	flag.StringVar(&HostPort, "http", ":54321", "HTTP Host:Port")
	flag.BoolVar(&Debug, "debug", false, "Debug.")
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

func respond(w http.ResponseWriter, m *Message) {
	etag := fmt.Sprintf("%d", m.Created)
	w.Header().Add("Etag", etag)
	j, err := json.Marshal(SubResponse{m.Created, []string{string(m.Data)}, ""})
	if err != nil {
		log.Println("Error during json.Marshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func respondMany(w http.ResponseWriter, ch *Channel, ith uint) {
	payload := []string{}
	etag := int64(0)
	ml := ch.Messages.Length()
	for i := ith; i < ml; i++ {
		ithm, _ := ch.Messages.Ith(i)
		payload = append(payload, string(ithm.Data))
		etag = ithm.Created
	}
	j, err := json.Marshal(SubResponse{etag, payload, ""})
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

func SubHandler(w http.ResponseWriter, r *http.Request) {
	channel := r.FormValue("channel")
	if channel == "" {
		reject(w, "channel is required")
		return
	}

	cid := r.FormValue("cid")
	if cid == "" {
		reject(w, "client id(cid) is required")
		return
	}

	etag_s := r.FormValue("etag")
	if etag_s == "" {
		etag_s = r.Header.Get("If-None-Match")
	}

	// TODO: whats the symantics for etag=0? the whole point of keeping old
	// messages is to send them to clients[1]. if we do that then there can be
	// duplicates. but if we dont do that then there can be data loss.
	//
	// one can argue that [1] is not correct and point is to not lose data mid
	// stream, eg get 1 and 3 but not 2 (unless of course 2 has expired by the
	// time 2nd request comes (which we cant help, increase life))

	etag := int64(0)
	if etag_s != "" {
		_, err := fmt.Sscan(etag_s, &etag)
		if err != nil {
			reject(w, "invalid etag: "+err.Error())
			return
		}
	}

	cner, ok := w.(http.CloseNotifier)
	if !ok {
		reject(w, "channel is required")
		return
	}

	ch := GetChannel(channel)
	sub, ith := ch.Sub(cid, etag)

	if sub == nil {
		respondMany(w, ch, ith)
		return
	}

	select {
	case m := <-sub:
		if m == nil {
			// new guy came with same client id, kill this connection
			reject(w, "oops, new client")
		} else {
			respond(w, m)
		}
	case <-cner.CloseNotify():
		ch.UnSub(cid)
	}
}

func OKHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func ServeHTTP() {
	// http.HandleFunc("/", OKHandler)
	http.HandleFunc("/pub", PubHandler)
	http.HandleFunc("/sub", SubHandler)
	http.HandleFunc("/stats", OKHandler)
	http.Handle("/", http.FileServer(FS(Debug)))

	log.Printf("Started HTTP Server on %s.", HostPort)
	logger := gutils.NewApacheLoggingHandler(http.DefaultServeMux, os.Stderr)
	server := &http.Server{
		Addr:    HostPort,
		Handler: logger,
	}
	log.Fatal(server.ListenAndServe())
}
