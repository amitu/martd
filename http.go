package main

import (
	"flag"
	"fmt"
	"github.com/amitu/gutils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	HostPort string
	Debug    bool
)

func init() {
	flag.StringVar(&HostPort, "http", ":5432", "HTTP Host:Port")
	flag.BoolVar(&Debug, "debug", false, "Debug.")
}

func PubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	channel := r.FormValue("channel") // TODO: support multiple channels
	size_s := r.FormValue("size")
	life_s := r.FormValue("life")
	one2one := r.FormValue("one2one") == "true"
	key := r.FormValue("key")

	if channel == "" {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}

	size := uint(10)
	if size_s != "" {
		_, err := fmt.Sscan(size_s, &size)
		if err != nil {
			http.Error(w, "invalid size: "+err.Error(), http.StatusBadRequest)
		}
	}

	life := time.Second * 10
	if life_s != "" {
		_, err := fmt.Sscan(life_s, &life)
		if err != nil {
			http.Error(w, "invalid life: "+err.Error(), http.StatusBadRequest)
		}
	}

	ch, err := GetOrCreateChannel(channel, size, life, one2one, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if len(body) != 0 {
		err := ch.Pub(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	j, err := ch.Json()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	fmt.Fprintf(w, "%s\n", j)
}

func SubHandler(w http.ResponseWriter, r *http.Request) {
	channel := r.FormValue("channel")
	if channel == "" {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}

	cid := r.FormValue("cid")
	if cid == "" {
		http.Error(w, "client id(cid) is required", http.StatusBadRequest)
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
			http.Error(w, "invalid etag: "+err.Error(), http.StatusBadRequest)
		}
	}

	cner, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}

	ch := GetChannel(channel)
	sub, m := ch.Sub(cid, etag)

	if m != nil {
		w.Header().Add("Etag", fmt.Sprintf("%d", m.Created))
		w.Write(m.Data)
		return
	}

	select {
	case m := <-sub:
		if m == nil {
			// new guy came with same client id, kill this connection
			fmt.Fprintf(w, "oops, new client")
		} else {
			w.Header().Add("Etag", fmt.Sprintf("%d", m.Created))
			w.Write(m.Data)
		}
	case <-cner.CloseNotify():
		ch.UnSub(cid)
	}
}

func OKHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func ServeHTTP() {
	http.Handle("/static/", http.FileServer(FS(Debug)))
	http.HandleFunc("/", OKHandler)
	http.HandleFunc("/pub", PubHandler)
	http.HandleFunc("/sub", SubHandler)
	http.HandleFunc("/stats", OKHandler)

	log.Printf("Started HTTP Server on %s.", HostPort)
	logger := gutils.NewApacheLoggingHandler(http.DefaultServeMux, os.Stderr)
	server := &http.Server{
		Addr:    HostPort,
		Handler: logger,
	}
	log.Fatal(server.ListenAndServe())
}
